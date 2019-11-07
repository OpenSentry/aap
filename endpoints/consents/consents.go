package consents

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/app"
  _ "github.com/charmixer/aap/config"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"
  E "github.com/charmixer/aap/client/errors"

  bulky "github.com/charmixer/bulky/server"
)

func GetConsents(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetConsents",
    })

    var requests []client.ReadConsentsRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

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

        var owner aap.Identity
        var subscriber aap.Identity
        var publisher aap.Identity
        var scopes []aap.Scope

        if request.Input != nil {
          r := request.Input.(client.ReadConsentsRequest)

          owner = aap.Identity{Id:r.Reference}

          if r.Subscriber != "" {
            subscriber = aap.Identity{Id:r.Subscriber}
          }

          if r.Publisher != "" {
            publisher = aap.Identity{Id:r.Publisher}
          }

          if r.Scopes != nil {
            for _,scopeName := range r.Scopes {
              scopes = append(scopes, aap.Scope{Name:scopeName})
            }
          }
        }

        dbConsents, err := aap.FetchConsents(tx, owner, subscriber, publisher, scopes)
        if err != nil {
          e := tx.Rollback()
          if e != nil {
            log.Debug(e.Error())
          }
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
          log.Debug(err.Error())
          return
        }

        if len(dbConsents) > 0 {
          var ok client.ReadConsentsResponse
          for _, d := range dbConsents {
            ok = append(ok, client.Consent{
              Reference: d.Identity.Id,
              Subscriber: d.Subscriber.Id,
              Publisher: d.Publisher.Id,
              Scope: d.Scope.Name,
            })
          }
          request.Output = bulky.NewOkResponse(request.Index, ok)
          continue
        }

        // Deny by default
        request.Output = bulky.NewOkResponse(request.Index, nil)
        continue
      }

      err = bulky.OutputValidateRequests(iRequests)
      if err == nil {
        tx.Commit()
        return
      }

      // Deny by default
      tx.Rollback()
    }

    responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{EnableEmptyRequest: true})
    c.JSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}


func PostConsents(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostConsents",
    })

    var requests []client.CreateConsentsRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var handleRequests = func(iRequests []*bulky.Request) {

      session, tx, err := aap.BeginWriteTx(env.Driver)
      if err != nil {
        bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
        log.Debug(err.Error())
        return
      }
      defer tx.Close() // rolls back if not already committed/rolled back
      defer session.Close()

      var newConsents []aap.Consent

      for _, request := range iRequests {
        r := request.Input.(client.CreateConsentsRequest)

        newConsent := aap.Consent{
          Identity: aap.Identity{Id:r.Reference},
          Subscriber: aap.Identity{Id:r.Subscriber},
          Publisher: aap.Identity{Id:r.Publisher},
          Scope: aap.Scope{Name:r.Scope},
        }

        consent, err := aap.CreateConsent(tx, newConsent.Identity, newConsent.Subscriber, newConsent.Publisher, newConsent.Scope)
        if err != nil {
          e := tx.Rollback()
          if e != nil {
            log.Debug(e.Error())
          }
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index)
          log.Debug(err.Error())
          return
        }

        if consent != (aap.Consent{}) {
          newConsents = append(newConsents, consent)

          ok := client.CreateConsentsResponse{
            Reference: consent.Identity.Id,
            Subscriber: consent.Subscriber.Id,
            Publisher: consent.Publisher.Id,
            Scope: consent.Scope.Name,
          }
          request.Output = bulky.NewOkResponse(request.Index, ok)
          aap.EmitEventConsentCreated(env.Nats, consent)
          continue
        }

        // Deny by default
        e := tx.Rollback()
        if e != nil {
          log.Debug(e.Error())
        }
        bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
        request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
        log.Debug(err.Error())
        return
      }

      err = bulky.OutputValidateRequests(iRequests)
      if err == nil {
        tx.Commit()
        // proxy to hydra. Not needed.
        return
      }

      // Deny by default
      tx.Rollback()
    }

    responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{})
    c.JSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}

func DeleteConsents(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "DeleteConsents",
    })

    var requests []client.DeleteConsentsRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var handleRequests = func(iRequests []*bulky.Request) {

      session, tx, err := aap.BeginWriteTx(env.Driver)
      if err != nil {
        bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
        log.Debug(err.Error())
        return
      }
      defer tx.Close() // rolls back if not already committed/rolled back
      defer session.Close()

      for _, request := range iRequests {
        r := request.Input.(client.DeleteConsentsRequest)

        var owner aap.Identity = aap.Identity{Id:r.Reference}

        var subscriber aap.Identity
        if r.Subscriber != "" {
          subscriber = aap.Identity{Id:r.Subscriber}
        }

        var publisher aap.Identity
        if r.Publisher != "" {
          publisher = aap.Identity{Id:r.Publisher}
        }

        var scopes []aap.Scope
        if r.Scope != "" {
          scopes = append(scopes, aap.Scope{Name:r.Scope})
        }

        dbConsents, err := aap.FetchConsents(tx, owner, subscriber, publisher, scopes)
        if err != nil {
          e := tx.Rollback()
          if e != nil {
            log.Debug(e.Error())
          }
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
          log.Debug(err.Error())
          return
        }

        if len(dbConsents) <= 0 {
          e := tx.Rollback()
          if e != nil {
            log.Debug(e.Error())
          }
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
          request.Output = bulky.NewClientErrorResponse(request.Index, E.CONSENT_NOT_FOUND)
          return
        }
        consentToDelete := dbConsents[0]

        if consentToDelete.Scope.Name != "" {

          _ /* deletedConsent */, err := aap.DeleteConsent(tx, consentToDelete.Identity, consentToDelete.Subscriber, consentToDelete.Publisher, consentToDelete.Scope)
          if err != nil {
            e := tx.Rollback()
            if e != nil {
              log.Debug(e.Error())
            }
            bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
            request.Output = bulky.NewInternalErrorResponse(request.Index)
            log.Debug(err.Error())
            return
          }

          ok := client.DeleteConsentsResponse{
            Reference: consentToDelete.Identity.Id,
            Subscriber: consentToDelete.Subscriber.Id,
            Publisher: consentToDelete.Publisher.Id,
            Scope: consentToDelete.Scope.Name,
          }
          request.Output = bulky.NewOkResponse(request.Index, ok)
          continue
        }

        // Deny by default
        e := tx.Rollback()
        if e != nil {
          log.Debug(e.Error())
        }
        bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
        request.Output = bulky.NewClientErrorResponse(request.Index, E.CONSENT_NOT_FOUND)
        log.Debug("Delete consent failed. Hint: Maybe input validation needs to be improved.")
        return
      }

      err = bulky.OutputValidateRequests(iRequests)
      if err == nil {
        tx.Commit()
        // proxy to hydra. Not needed
        return
      }

      // Deny by default
      tx.Rollback()
    }

    responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{})
    c.JSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}
