package consents

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"
  hydra "github.com/charmixer/hydra/client"

  "github.com/charmixer/aap/app"
  "github.com/charmixer/aap/config"
  // "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"

  bulky "github.com/charmixer/bulky/server"
)

func PostAuthorize(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostAuthorize",
    })

    var requests []client.CreateConsentsAuthorizeRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    // Create a new HTTP client to perform the request, to prevent serialization
    hydraClient := hydra.NewHydraClient(env.OAuth2Delegator.Config)

    var handleRequests = func(iRequests []*bulky.Request) {

      for _, request := range iRequests {
        r := request.Input.(client.CreateConsentsAuthorizeRequest)

        log = log.WithFields(logrus.Fields{"challenge": r.Challenge})

        hydraConsentResponse, err := hydra.GetConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consent"), hydraClient, r.Challenge)
        if err != nil {
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
          log.Debug(err.Error())
          return
        }

        log.Debug(hydraConsentResponse)

        var grantedScopes []string = r.GrantScopes
        var subject string = hydraConsentResponse.Subject
        var requestedScopes []string = hydraConsentResponse.RequestedScopes
        var grantedAccessTokenAudiences []string = hydraConsentResponse.RequestedAccessTokenAudience

        var clientId string = hydraConsentResponse.Client.ClientId
        if clientId == "" {
          log.Debug("No client_id found")
        }

        var clientName string
        var subjectName string
        var subjectEmail string

        loginContext := hydraConsentResponse.Context
        if loginContext != nil {
          log.Debug(loginContext)
          clientName = loginContext["client_name"]
          subjectName = loginContext["subject_name"]
          subjectEmail = loginContext["subject_email"]
        }

        if hydraConsentResponse.Skip {
          // Grant all scopes that have been requested - hydra already checked for us that no additional scopes are requested accidentally.
          grantedScopes = requestedScopes
        }

        if len(grantedScopes) > 0 {

          consentAcceptRequest := hydra.ConsentAcceptRequest{
            GrantScope: grantedScopes,
            Session: hydra.ConsentAcceptSession {
            },
            GrantAccessTokenAudience: grantedAccessTokenAudiences,
            Remember: true,
            RememberFor: 0, // Never expire consent in hydra. Control this from aap system
          }
          hydraConsentAcceptResponse, err := hydra.AcceptConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consentAccept"), hydraClient, r.Challenge, consentAcceptRequest)
          if err != nil {
            bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
            request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
            log.Debug(err.Error())
            return
          }

          consent := client.CreateConsentsAuthorizeResponse{
            Challenge: r.Challenge,
            Authorized: true,
            RedirectTo: hydraConsentAcceptResponse.RedirectTo,

            ClientId: clientId,
            ClientName: clientName,

            Subject: subject,
            SubjectName: subjectName,
            SubjectEmail: subjectEmail,

            RequestedScopes: requestedScopes,
            GrantedScopes: grantedScopes,

            RequestedAudiences: grantedAccessTokenAudiences,
          }

          log.WithFields(logrus.Fields{ "challenge":consent.Challenge, "authorized":consent.Authorized }).Debug("Consent Accepted")
          request.Output = bulky.NewOkResponse(request.Index, consent)
          continue
        }

        // Deny by default (Read challenge data from hydra)
        consent := client.CreateConsentsAuthorizeResponse{
          Challenge: r.Challenge,
          Subject: subject,
          ClientId: clientId,
          Authorized: false,
          RequestedScopes: requestedScopes,
        }
        request.Output = bulky.NewOkResponse(request.Index, consent)
      }

      bulky.OutputValidateRequests(iRequests)
    }

    responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{MaxRequests: 1})
    c.JSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}

func PostReject(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostReject",
    })

    var requests []client.CreateConsentsRejectRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    // Create a new HTTP client to perform the request, to prevent serialization
    hydraClient := hydra.NewHydraClient(env.OAuth2Delegator.Config)

    var handleRequests = func(iRequests []*bulky.Request) {

      for _, request := range iRequests {
        r := request.Input.(client.CreateConsentsAuthorizeRequest)

        log = log.WithFields(logrus.Fields{"challenge": r.Challenge})

        hydraConsentResponse, err := hydra.GetConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consent"), hydraClient, r.Challenge)
        if err != nil {
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
          log.Debug(err.Error())
          return
        }

        var subject string = hydraConsentResponse.Subject
        var requestedScopes []string = hydraConsentResponse.RequestedScopes
        var clientId string = hydraConsentResponse.Client.ClientId
        var requestedAccessTokenAudiences []string = hydraConsentResponse.RequestedAccessTokenAudience

        var clientName string
        var subjectName string
        var subjectEmail string

        loginContext := hydraConsentResponse.Context
        if loginContext != nil {
          log.Debug(loginContext)
          clientName = loginContext["client_name"]
          subjectName = loginContext["subject_name"]
          subjectEmail = loginContext["subject_email"]
        }

        hydraConsentRejectResponse, err := hydra.RejectConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consentReject"), hydraClient, r.Challenge, hydra.ConsentRejectRequest{
          Error: "",
          ErrorDebug: "",
          ErrorDescription: "",
          ErrorHint: "",
          StatusCode: http.StatusForbidden,
        })
        if err != nil {
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
          log.Debug(err.Error())
          return
        }

        reject := client.CreateConsentsRejectResponse{
          Challenge: r.Challenge,
          Authorized: false,
          RedirectTo: hydraConsentRejectResponse.RedirectTo,

          ClientId: clientId,
          ClientName: clientName,

          Subject: subject,
          SubjectName: subjectName,
          SubjectEmail: subjectEmail,

          RequestedScopes: requestedScopes,
          GrantedScopes: []string{},

          RequestedAudiences: requestedAccessTokenAudiences,
        }

        log.WithFields(logrus.Fields{ "challenge":reject.Challenge, "authorized":reject.Authorized }).Debug("Consent Rejected")
        request.Output = bulky.NewOkResponse(request.Index, reject)
      }

      bulky.OutputValidateRequests(iRequests)
    }

    responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{MaxRequests: 1})
    c.JSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}

