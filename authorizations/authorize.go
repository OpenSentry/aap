package authorizations

import (
  "fmt"
  "net/http"
  "net/url"
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
      "route.logid": route.LogId,
      "component": "authorizations",
      "func": "PostAuthorize",
    })

    log.Debug("Received authorize request")

    var input AuthorizeRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      c.Abort()
      return
    }

    // Create a new HTTP client to perform the request, to prevent serialization
    hydraClient := hydra.NewHydraClient(env.HydraConfig)

    authorizeResponse, err := authorize(hydraClient, input)
    if err != nil {
      c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
      c.Abort()
      return
    }

    logMsg := fmt.Sprintf("client_id:%s subject:%s authorized:%s redirect_to:%s", authorizeResponse.ClientId, authorizeResponse.Subject, authorizeResponse.Authorized, authorizeResponse.RedirectTo)
    log.Debug(logMsg)
    c.JSON(http.StatusOK, authorizeResponse)
  }
  return gin.HandlerFunc(fn)
}

func PostReject(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "route.logid": route.LogId,
      "component": "authorizations",
      "func": "PostReject",
    })

    log.Debug("Received reject request")

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

    log.Debug("authorized:false redirect_to:" + hydraConsentRejectResponse.RedirectTo)
    c.JSON(http.StatusOK, gin.H{
      "authorized": false,
      "redirect_to": hydraConsentRejectResponse.RedirectTo,
    })
  }
  return gin.HandlerFunc(fn)
}


// helper
func authorize(client *hydra.HydraClient, authorizeRequest AuthorizeRequest) (AuthorizeResponse, error) {
  var authorizeResponse AuthorizeResponse

  hydraConsentResponse, err := hydra.GetConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consent"), client, authorizeRequest.Challenge)
  if err != nil {
    return authorizeResponse, err
  }

  // Extract client_id from RequestUrl
  // FIXME: Is there another way to find out on behalf of which client the request is made?
  var clientId string
  if hydraConsentResponse.RequestUrl != "" {
    u, err := url.Parse(hydraConsentResponse.RequestUrl)
    if err == nil {
      q := u.Query();
      clientId = q.Get("client_id")
    }
  }

  if hydraConsentResponse.Skip {
    hydraConsentAcceptRequest := hydra.ConsentAcceptRequest{
      GrantScope: hydraConsentResponse.RequestedScopes, // We can grant all scopes that have been requested - hydra already checked for us that no additional scopes are requested accidentally.
      Session: hydra.ConsentAcceptSession {
      },
      GrantAccessTokenAudience: hydraConsentResponse.GrantAccessTokenAudience,
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
      GrantScopes: hydraConsentResponse.RequestedScopes,
      RequestedScopes: authorizeRequest.GrantScopes,
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
    GrantAccessTokenAudience: hydraConsentResponse.GrantAccessTokenAudience,
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
