package consents

import (
  "strings"
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"
  hydra "github.com/charmixer/hydra/client"

  "github.com/opensentry/aap/app"
  "github.com/opensentry/aap/config"
  "github.com/opensentry/aap/gateway/aap"
  "github.com/opensentry/aap/client"
  E "github.com/opensentry/aap/client/errors"

  bulky "github.com/charmixer/bulky/server"

  "github.com/neo4j/neo4j-go-driver/neo4j" // To remove, move private functions fetchSubscriptions, fetchConsents to gateways.
)

// Set Difference: A - B
func Difference(a, b []string) (diff []string) {
  m := make(map[string]bool)

  for _, item := range b {
    m[item] = true
  }

  for _, item := range a {
    if _, ok := m[item]; !ok {
      diff = append(diff, item)
    }
  }
  return
}

type ConsentChallenge struct {
  Skip bool

  Subject string
  SubjectName string
  SubjectEmail string

  ClientId string
  ClientName string

  RequestedScopes []string
  RequestedAudiences []string
}


func GetAuthorize(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetAuthorize",
    })

    var requests []client.ReadConsentsAuthorizeRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    // Create a new HTTP client to perform the request, to prevent serialization
    hydraClient := hydra.NewHydraClient(env.OAuth2Delegator.Config)

    var handleRequests = func(iRequests []*bulky.Request) {

      session, tx, err := aap.BeginReadTx(env.Driver)
      if err != nil {
        bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
        log.Debug(err.Error())
        return
      }
      defer tx.Close() // rolls back if not already committed/rolled back
      defer session.Close()

      for _, request := range iRequests {
        r := request.Input.(client.ReadConsentsAuthorizeRequest)

        log = log.WithFields(logrus.Fields{"challenge": r.Challenge})

        hydraConsentResponse, err := hydra.GetConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consent"), hydraClient, r.Challenge)
        if err != nil {
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
          log.Debug(err.Error())
          return
        }

        consentChallenge := parseConsentChallenge(hydraConsentResponse)

        // Prepare db lookup filters based on consent challenge.
        iFilterOwner := aap.Identity{Id:consentChallenge.Subject}
        iFilterSubscriber := aap.Identity{Id: consentChallenge.ClientId}

        var iFilterScopes []aap.Scope
        for _,scopeName := range consentChallenge.RequestedScopes {
          iFilterScopes = append(iFilterScopes, aap.Scope{Name: scopeName})
        }
        var iFilterPublishers []aap.Identity
        if len(consentChallenge.RequestedAudiences) > 0 {
          for _,publisherId := range consentChallenge.RequestedAudiences {
            iFilterPublishers = append(iFilterPublishers, aap.Identity{Id:publisherId})
          }
        }

        consentRequests, subscribedScopes, consentedScopes, consentedAudiences, err := fetchConsentRequests(tx, iFilterOwner, iFilterSubscriber, iFilterPublishers, iFilterScopes)
        if err != nil {
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
          log.Debug(err.Error())
          return
        }

        // Sanity check. Require atleast one subscription by client
        if len(subscribedScopes) <= 0 {
          log.WithFields(logrus.Fields{ "client_id":iFilterSubscriber.Id }).Debug("No subscriptions")
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests)
          request.Output = bulky.NewClientErrorResponse(request.Index, E.NO_SUBSCRIPTIONS)
          return
        }

        // Sanity check. All requested scopes must be subscribed to by client
        if len(subscribedScopes) < len(consentChallenge.RequestedScopes) {

          invalidScopes := Difference(consentChallenge.RequestedScopes, subscribedScopes)
          log.WithFields(logrus.Fields{ "client_id":iFilterSubscriber.Id, "scopes": strings.Join(invalidScopes, " ") }).Debug("Invalid scopes")

          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests)
          request.Output = bulky.NewClientErrorResponse(request.Index, E.INVALID_SCOPES)
          return
        }


        consentAuthorization := client.ReadConsentsAuthorizeResponse{
          Challenge: r.Challenge,
          Authorized: false,
          RedirectTo: "", // This can only come from a hydra accept consent call.

          ClientId: consentChallenge.ClientId,
          ClientName: consentChallenge.ClientName,

          Subject: consentChallenge.Subject,
          SubjectName: consentChallenge.SubjectName,
          SubjectEmail: consentChallenge.SubjectEmail,

          ConsentRequests: consentRequests,
        }

        var hydraAcceptConsent bool = false
        var hydraGrantScopes []string
        var hydraGrantAudience []string

        if consentChallenge.Skip == true {
          // Grant all scopes that have been requested - hydra already checked for us that no additional scopes are requested accidentally.
          hydraGrantScopes = consentChallenge.RequestedScopes
          hydraGrantAudience = consentChallenge.RequestedAudiences
          hydraAcceptConsent = true
        }

        // If not skip in hydra but all consented in db model, then accept consent.
        if len(consentedAudiences) == len(consentChallenge.RequestedAudiences) && len(consentedScopes) == len(consentChallenge.RequestedScopes) {
          hydraGrantScopes = consentChallenge.RequestedScopes
          hydraGrantAudience = consentChallenge.RequestedAudiences
          hydraAcceptConsent = true
        }

        if hydraAcceptConsent == true {

          hydraConsentAcceptResponse, err := hydra.AcceptConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consentAccept"), hydraClient, r.Challenge, hydra.ConsentAcceptRequest{
            GrantScope: hydraGrantScopes,
            Session: hydra.ConsentAcceptSession {
            },
            GrantAccessTokenAudience: hydraGrantAudience,
            Remember: true,
            RememberFor: 0, // Never expire consent in hydra. Control this from aap system
          })
          if err != nil {
            bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
            request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
            log.Debug(err.Error())
            return
          }

          // Consent to access
          consentAuthorization.Authorized = true
          consentAuthorization.RedirectTo = hydraConsentAcceptResponse.RedirectTo
        }

        request.Output = bulky.NewOkResponse(request.Index, consentAuthorization)
        continue
      }

      bulky.OutputValidateRequests(iRequests)
    }
    responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{MaxRequests: 1})
    c.JSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}

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

      session, tx, err := aap.BeginReadTx(env.Driver)
      if err != nil {
        bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
        log.Debug(err.Error())
        return
      }
      defer tx.Close() // rolls back if not already committed/rolled back
      defer session.Close()

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

        consentChallenge := parseConsentChallenge(hydraConsentResponse)

        // Prepare db lookup filters based on consent challenge.
        iFilterOwner := aap.Identity{Id:consentChallenge.Subject}
        iFilterSubscriber := aap.Identity{Id: consentChallenge.ClientId}

        var iFilterScopes []aap.Scope
        for _,scopeName := range consentChallenge.RequestedScopes {
          iFilterScopes = append(iFilterScopes, aap.Scope{Name: scopeName})
        }
        var iFilterPublishers []aap.Identity
        if len(consentChallenge.RequestedAudiences) > 0 {
          for _,publisherId := range consentChallenge.RequestedAudiences {
            iFilterPublishers = append(iFilterPublishers, aap.Identity{Id:publisherId})
          }
        }

        consentRequests, subscribedScopes, consentedScopes, consentedAudiences, err := fetchConsentRequests(tx, iFilterOwner, iFilterSubscriber, iFilterPublishers, iFilterScopes)
        if err != nil {
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
          log.Debug(err.Error())
          return
        }

        // Sanity check. Require atleast one subscription by client
        if len(subscribedScopes) <= 0 {
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests)
          request.Output = bulky.NewClientErrorResponse(request.Index, E.NO_SUBSCRIPTIONS)
          return
        }

        // Sanity check. All requested scopes must be subscribed to by client
        if len(subscribedScopes) < len(consentChallenge.RequestedScopes) {

          invalidScopes := Difference(consentChallenge.RequestedScopes, subscribedScopes)
          log.Debug(invalidScopes)

          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests)
          request.Output = bulky.NewClientErrorResponse(request.Index, E.INVALID_SCOPES)
          return
        }

        consentAuthorization := client.ReadConsentsAuthorizeResponse{
          Challenge: r.Challenge,
          Authorized: false,
          RedirectTo: "", // This can only come from a hydra accept consent call.

          ClientId: consentChallenge.ClientId,
          ClientName: consentChallenge.ClientName,

          Subject: consentChallenge.Subject,
          SubjectName: consentChallenge.SubjectName,
          SubjectEmail: consentChallenge.SubjectEmail,

          ConsentRequests: consentRequests,
        }

        var hydraGrantScopes []string = consentedScopes // Accept all consented scopes from DB model
        var hydraGrantAudience []string = consentedAudiences // Accept all consented audience from DB model

        hydraConsentAcceptResponse, err := hydra.AcceptConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consentAccept"), hydraClient, r.Challenge, hydra.ConsentAcceptRequest{
          GrantScope: hydraGrantScopes,
          Session: hydra.ConsentAcceptSession {
          },
          GrantAccessTokenAudience: hydraGrantAudience,
          Remember: true,
          RememberFor: 0, // Never expire consent in hydra. Control this from aap system
        })
        if err != nil {
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
          log.Debug(err.Error())
          return
        }

        // Consent to access
        consentAuthorization.Authorized = true
        consentAuthorization.RedirectTo = hydraConsentAcceptResponse.RedirectTo
        request.Output = bulky.NewOkResponse(request.Index, consentAuthorization)
        continue
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
        r := request.Input.(client.CreateConsentsRejectRequest)

        log = log.WithFields(logrus.Fields{"challenge": r.Challenge})

        hydraConsentResponse, err := hydra.GetConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consent"), hydraClient, r.Challenge)
        if err != nil {
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
          log.Debug(err.Error())
          return
        }

        consentChallenge := parseConsentChallenge(hydraConsentResponse)

        hydraConsentRejectResponse, err := hydra.RejectConsent(config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.consentReject"), hydraClient, r.Challenge, hydra.ConsentRejectRequest{
          Error: "Access Denied",
          ErrorDebug: "",
          ErrorDescription: "Subject did not consent to access",
          ErrorHint: "",
          StatusCode: http.StatusForbidden,
        })
        if err != nil {
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
          log.Debug(err.Error())
          return
        }

        // Reject access
        request.Output = bulky.NewOkResponse(request.Index, client.CreateConsentsRejectResponse{
          Challenge: r.Challenge,
          Authorized: false,
          RedirectTo: hydraConsentRejectResponse.RedirectTo,
          ClientId: consentChallenge.ClientId,
          Subject: consentChallenge.Subject,
        })
        continue
      }

      bulky.OutputValidateRequests(iRequests)
    }

    responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{MaxRequests: 1})
    c.JSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}

func parseConsentChallenge(hydraConsentResponse hydra.ConsentResponse) (consentChallenge ConsentChallenge) {
  consentChallenge = ConsentChallenge{
    Skip: hydraConsentResponse.Skip,
    Subject: hydraConsentResponse.Subject,
    ClientId: hydraConsentResponse.Client.ClientId,
    RequestedScopes: hydraConsentResponse.RequestedScopes,
    RequestedAudiences: hydraConsentResponse.RequestedAccessTokenAudience,
  }

  loginContext := hydraConsentResponse.Context
  if loginContext != nil {
    consentChallenge.ClientName = loginContext["client_name"]
    consentChallenge.SubjectName = loginContext["subject_name"]
    consentChallenge.SubjectEmail = loginContext["subject_email"]
  }

  return consentChallenge
}

type PublisherScope struct {
  Publisher, Scope string
}

func fetchConsentRequests(tx neo4j.Transaction, iFilterOwner aap.Identity, iFilterSubscriber aap.Identity, iFilterPublishers []aap.Identity, iFilterScopes []aap.Scope) (consentRequests []client.ConsentRequest, subscribedScopes []string, consentedScopes []string, consentedAudiences []string, err error) {

  // No publisher given, so use the all publisher the one with Id = ""
  if len(iFilterPublishers) <= 0 {
    iFilterPublishers = append(iFilterPublishers, aap.Identity{})
  }

  subscriptions := make(map[string][]aap.Subscription)
  consents      := make(map[PublisherScope]bool)
  publishings   := make(map[PublisherScope]aap.Publish)

  for _,publisher := range iFilterPublishers {

    // Lookup definitions of scopes for each given publisher
    dbPublishes, err := aap.FetchPublishes(tx, publisher, iFilterScopes)
    if err != nil {
      return nil, nil, nil, nil, err
    }
    for _, pub := range dbPublishes {
      publishings[PublisherScope{pub.Publisher.Id, pub.Scope.Name}] = pub
    }

    // Lookup subscriptions for client to each publisher
    dbSubscriptions, err := aap.FetchSubscriptions(tx, iFilterSubscriber, publisher, iFilterScopes)
    if err != nil {
      return nil, nil, nil, nil, err
    }
    subscriptions[publisher.Id] = append(subscriptions[publisher.Id], dbSubscriptions...)

    // Initialize consent map for all subscriptions.
    for _, sub := range dbSubscriptions {
      consents[PublisherScope{sub.Publisher.Id, sub.Scope.Name}] = false
    }

    // Lookup consents already given to the client to publisher scope by subject
    dbConsents, err := aap.FetchConsents(tx, iFilterOwner, iFilterSubscriber, publisher, iFilterScopes)
    if err != nil {
      return nil, nil, nil, nil, err
    }
    for _, consent := range dbConsents {
      consentedScopes = append(consentedScopes, consent.Scope.Name)
      consents[PublisherScope{consent.Publisher.Id, consent.Scope.Name}] = true
    }
  }

  // Unique subscription scopes.
  // https://medium.com/@l.peppoloni/how-to-improve-your-go-code-with-empty-structs-3bd0c66bc531
  // "The cool thing about an empty structure is that it occupies zero bytes of storage."
  _subscribedScopes := make(map[string]struct{})
  for _, subs := range subscriptions {
    for _, sub := range subs {
      _subscribedScopes[sub.Scope.Name] = struct{}{}
    }
  }
  for scope,_ := range _subscribedScopes {
    subscribedScopes = append(subscribedScopes, scope)
  }

  // Build consent requests.
  for _, subs := range subscriptions {

    for _, sub := range subs {

      pub := publishings[PublisherScope{sub.Publisher.Id, sub.Scope.Name}]
      isConsented := consents[PublisherScope{sub.Publisher.Id, sub.Scope.Name}]

      consentRequest := client.ConsentRequest{
        Scope: sub.Scope.Name,
        Audience: sub.Publisher.Id,
        Title: pub.Rule.Title,
        Description: pub.Rule.Description,
        Consented: isConsented,
      }
      consentRequests = append(consentRequests, consentRequest)
    }

  }

  return consentRequests, subscribedScopes, consentedScopes, consentedAudiences, nil
}

