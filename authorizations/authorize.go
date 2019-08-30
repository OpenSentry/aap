package authorizations

import (
  "net/http"
  //"net/url"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"
  "github.com/CharMixer/hydra-client" // FIXME: Do not use upper case
  "golang-cp-be/config"
  "golang-cp-be/environment"
)

type AuthorizeRequest struct {
  Challenge                   string            `json:"challenge" binding:"required"`
  GrantScopes                 []string          `json:"grant_scopes,omitempty"`
}

type AuthorizeResponse struct {
  Challenge                   string            `json:"challenge" binding:"required"`
  Authorized                  bool              `json:"authorized" binding:"required"`
  GrantScopes                 []string          `json:"grant_scopes,omitempty"`
  RequestedScopes             []string          `json:"requested_scopes,omitempty"`
  RedirectTo                  string            `json:"redirect_to,omitempty"`
  Subject                     string            `json:"subject,omitempty"`
  ClientId                    string            `json:"client_id,omitempty"`
  RequestedAudiences          []string          `json:"requested_audiences,omitempty"` // requested_access_token_audience
}

type RejectRequest struct {
  Challenge                   string            `json:"challenge" binding:"required"`
}

type RejectResponse struct {
  Authorized                  bool              `json:"authorized" binding:"required"`
  RedirectTo                  string            `json:"redirect_to" binding:"required"`
}

func PostAuthorize(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostAuthorize",
    })

    var input AuthorizeRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      c.Abort()
      return
    }

    // Create a new HTTP client to perform the request, to prevent serialization
    hydraClient := hydra.NewHydraClient(env.HydraConfig)

    authorizeResponse, err := authorize(hydraClient, input, log)
    if err != nil {
      c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
      c.Abort()
      return
    }

    log.WithFields(logrus.Fields{
      "client_id": authorizeResponse.ClientId,
      "subject": authorizeResponse.Subject,
      "authorized": authorizeResponse.Authorized,
      "redirect_to": authorizeResponse.RedirectTo,
    }).Debug("Authorized authorization")
    c.JSON(http.StatusOK, authorizeResponse)
  }
  return gin.HandlerFunc(fn)
}

func PostReject(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostReject",
    })

    var input RejectRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      c.Abort()
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
      c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
      c.Abort()
      return
    }

    rejectResponse := RejectResponse{
      Authorized: false,
      RedirectTo: hydraConsentRejectResponse.RedirectTo,
    }

    log.WithFields(logrus.Fields{
      "authorized": rejectResponse.Authorized,
      "redirect_to": rejectResponse.RedirectTo,
    }).Debug("Rejected authorization")
    c.JSON(http.StatusOK, rejectResponse)
  }
  return gin.HandlerFunc(fn)
}


// helper
func authorize(client *hydra.HydraClient, authorizeRequest AuthorizeRequest, log *logrus.Entry) (AuthorizeResponse, error) {
  var authorizeResponse AuthorizeResponse

  hydraConsentResponse, err := hydra.GetConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consent"), client, authorizeRequest.Challenge)
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
    hydraConsentAcceptResponse, err := hydra.AcceptConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consentAccept"), client, authorizeRequest.Challenge, hydraConsentAcceptRequest)
    if err != nil {
      return authorizeResponse, err
    }

    authorizeResponse = AuthorizeResponse{
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
    authorizeResponse = AuthorizeResponse{
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
  hydraConsentAcceptResponse, err := hydra.AcceptConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consentAccept"), client, authorizeRequest.Challenge, hydraConsentAcceptRequest)
  if err != nil {
    return authorizeResponse, err
  }

  authorizeResponse = AuthorizeResponse{
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
