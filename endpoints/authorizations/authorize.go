package authorizations

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"
  hydra "github.com/charmixer/hydra/client"

  client "github.com/charmixer/aap/client"
  "github.com/charmixer/aap/config"
  "github.com/charmixer/aap/environment"
)

func PostAuthorize(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostAuthorize",
    })

    var input client.AuthorizeRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    // Create a new HTTP client to perform the request, to prevent serialization
    hydraClient := hydra.NewHydraClient(env.HydraConfig)

    authorizeResponse, err := authorize(hydraClient, input, log)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
      return
    }

    log.WithFields(logrus.Fields{
      "client_id": authorizeResponse.ClientId,
      "subject": authorizeResponse.Subject,
      "authorized": authorizeResponse.Authorized,
      "redirect_to": authorizeResponse.RedirectTo,
    }).Debug("Authorized authorization")
    c.AbortWithStatusJSON(http.StatusOK, authorizeResponse)
  }
  return gin.HandlerFunc(fn)
}

func PostReject(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostReject",
    })

    var input client.RejectRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    hydraClient := hydra.NewHydraClient(env.HydraConfig)

    hydraConsentRejectRequest := hydra.ConsentRejectRequest{
      Error: "",
      ErrorDebug: "",
      ErrorDescription: "",
      ErrorHint: "",
      StatusCode: 403,
    }
    hydraConsentRejectResponse, err := hydra.RejectConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consentReject"), hydraClient, input.Challenge, hydraConsentRejectRequest)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
      return
    }

    rejectResponse := client.RejectResponse{
      Authorized: false,
      RedirectTo: hydraConsentRejectResponse.RedirectTo,
    }

    log.WithFields(logrus.Fields{
      "authorized": rejectResponse.Authorized,
      "redirect_to": rejectResponse.RedirectTo,
    }).Debug("Rejected authorization")
    c.AbortWithStatusJSON(http.StatusOK, rejectResponse)
  }
  return gin.HandlerFunc(fn)
}


// helper
func authorize(hydraClient *hydra.HydraClient, authorizeRequest client.AuthorizeRequest, log *logrus.Entry) (client.AuthorizeResponse, error) {
  var authorizeResponse client.AuthorizeResponse

  hydraConsentResponse, err := hydra.GetConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consent"), hydraClient, authorizeRequest.Challenge)
  if err != nil {
    return authorizeResponse, err
  }

  clientId := hydraConsentResponse.Client.ClientId
  if clientId == "" {
    log.WithFields(logrus.Fields{"consent_challenge":authorizeRequest.Challenge}).Debug("No client_id found")
  }

  if hydraConsentResponse.Skip {
    hydraConsentAcceptRequest := hydra.ConsentAcceptRequest{
      GrantScope: hydraConsentResponse.RequestedScopes, // We can grant all scopes that have been requested - hydra already checked for us that no additional scopes are requested accidentally.
      Session: hydra.ConsentAcceptSession {
      },
      GrantAccessTokenAudience: hydraConsentResponse.RequestedAccessTokenAudience,
      Remember: true, // FIXME: Mindre timeout eller flere kald mod neo?
      RememberFor: 0, // Never expire consent in hydra. Control this from aap system
    }
    hydraConsentAcceptResponse, err := hydra.AcceptConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consentAccept"), hydraClient, authorizeRequest.Challenge, hydraConsentAcceptRequest)
    if err != nil {
      return authorizeResponse, err
    }

    authorizeResponse = client.AuthorizeResponse{
      Challenge: authorizeRequest.Challenge,
      Subject: hydraConsentResponse.Subject,
      ClientId: clientId,
      Authorized: true,
      GrantScopes: hydraConsentResponse.RequestedScopes,
      RequestedScopes: authorizeRequest.GrantScopes,
      RequestedAudiences: hydraConsentResponse.RequestedAccessTokenAudience,
      RedirectTo: hydraConsentAcceptResponse.RedirectTo,
    }
    return authorizeResponse, nil
  }

  // Require atleast one scope to grant or this is just a masked read.
  if len(authorizeRequest.GrantScopes) <= 0 {
    authorizeResponse = client.AuthorizeResponse{
      Challenge: authorizeRequest.Challenge,
      Subject: hydraConsentResponse.Subject,
      ClientId: clientId,
      Authorized: false,
      RequestedScopes: hydraConsentResponse.RequestedScopes,
    }
    return authorizeResponse, nil
  }

  hydraConsentAcceptRequest := hydra.ConsentAcceptRequest{
    GrantScope: authorizeRequest.GrantScopes,
    Session: hydra.ConsentAcceptSession {
    },
    GrantAccessTokenAudience: hydraConsentResponse.RequestedAccessTokenAudience, // FIXME this should be changed to allow for choosing audience (resource server) the user trust
    Remember: true,
    RememberFor: 0, // Never expire consent in hydra. Control this from aap system
  }
  hydraConsentAcceptResponse, err := hydra.AcceptConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consentAccept"), hydraClient, authorizeRequest.Challenge, hydraConsentAcceptRequest)
  if err != nil {
    return authorizeResponse, err
  }

  authorizeResponse = client.AuthorizeResponse{
    Challenge: authorizeRequest.Challenge,
    Subject: hydraConsentResponse.Subject,
    ClientId: clientId,
    Authorized: true,
    GrantScopes: authorizeRequest.GrantScopes,
    RequestedScopes: authorizeRequest.GrantScopes,
    RedirectTo: hydraConsentAcceptResponse.RedirectTo,
  }
  return authorizeResponse, nil
}
