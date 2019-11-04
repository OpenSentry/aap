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

      // requestor := c.MustGet("sub").(string)
      // var requestedBy *idp.Identity
      // if requestor != "" {
      //   identities, err := idp.FetchIdentities(tx, []idp.Identity{ {Id:requestor} })
      //   if err != nil {
      //     bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
      //     log.Debug(err.Error())
      //     return
      //   }
      //   if len(identities) > 0 {
      //     requestedBy = &identities[0]
      //   }
      // }

      for _, request := range iRequests {

        var dbConsents []aap.Consent
        var err error
        var ok client.ReadConsentsResponse

        r := request.Input.(client.ReadConsentsRequest)

        var owner *aap.Identity = &aap.Identity{Id:r.Reference}

        var subscriber *aap.Identity
        if r.Subscriber != "" {
          subscriber = &aap.Identity{Id:r.Subscriber}
        }

        var publisher *aap.Identity
        if r.Publisher != "" {
          publisher = &aap.Identity{Id:r.Publisher}
        }

        var scopes []aap.Scope
        if r.Scope != "" {
          scopes = append(scopes, aap.Scope{Name:r.Scope})
        }

        dbConsents, err = aap.FetchConsents(tx, owner, subscriber, publisher, scopes)
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
        request.Output = bulky.NewClientErrorResponse(request.Index, E.CONSENT_NOT_FOUND)
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

      // requestor := c.MustGet("sub").(string)
      // var requestedBy *idp.Identity
      // if requestor != "" {
      //   identities, err := idp.FetchIdentities(tx, []idp.Identity{ {Id:requestor} })
      //   if err != nil {
      //     bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
      //     log.Debug(err.Error())
      //     return
      //   }
      //   if len(identities) > 0 {
      //     requestedBy = &identities[0]
      //   }
      // }

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
        log.WithFields(logrus.Fields{ "name": newConsent.Reference }).Debug(err.Error())
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

    responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{MaxRequests: 1})
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

    var input client.ConsentRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    if len(input.RequestedScopes) <= 0 {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing granted_scopes"})
      return
    }

    var grantPermissions []aap.Scope
    for _, scope := range input.GrantedScopes {
      grantPermissions = append(grantPermissions, aap.Scope{ Name:scope,})
    }

    var revokePermissions []aap.Scope
    for _, scope := range input.RevokedScopes {
      revokePermissions = append(revokePermissions, aap.Scope{ Name:scope,})
    }

    resourceOwner := aap.Identity{
      Id: input.Subject,
    }
    client := aap.Client{
      ClientId: input.ClientId,
    }

    var permissionList []aap.Scope
    if len(input.RequestedAudiences) > 0 {
      if ( len(input.RequestedAudiences) ) > 1 {
        log.WithFields(logrus.Fields{"requested_audiences":input.RequestedAudiences}).Debug("More than one audience not supported yet")
        c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
          "error": "More than one audience not supported yet Hint: Try only to use audience per token request one for now",
        })
        return
      }

      resourceServer, err := aap.FetchResourceServerByAudience(env.Driver, input.RequestedAudiences[0])
      if err != nil {
        log.WithFields(logrus.Fields{"aud":input.RequestedAudiences}).Debug("Resource server not found")
        c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
          "error": "Not found. Hint: Maybe audience does not exist.",
        })
        return
      }
      permissionList, err = aap.CreateConsentsToResourceServerForClientOnBehalfOfResourceOwner(env.Driver, resourceOwner, client, *resourceServer, grantPermissions, revokePermissions)
    } else {
      permissionList, err = aap.CreateConsentsForClientOnBehalfOfResourceOwner(env.Driver, resourceOwner, client, grantPermissions, revokePermissions)
    }
    if err != nil {
      log.Debug(err.Error())
    }

    if len(permissionList) > 0 {
      var grantedPermissions []string
      for _, permission := range permissionList {
        grantedPermissions = append(grantedPermissions, permission.Name)
      }

      c.AbortWithStatusJSON(http.StatusOK, grantedPermissions)
      return
    }

    // Deny by default
    c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
      "error": "Not found. Hint does the client exists?",
    })
  }
  return gin.HandlerFunc(fn)
}

func DeleteConsents(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "DeleteConsents",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{
      "error": "Missing implementation",
    })
  }
  return gin.HandlerFunc(fn)
}
