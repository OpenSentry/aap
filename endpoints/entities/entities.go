package entities

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/app"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"

  bulky "github.com/charmixer/bulky/server"

  //"fmt"
)

func PostEntities(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostEntities",
    })

    var requests []client.CreateEntitiesRequest
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

      validScopes := []string{
        "aap:read:grants",
        "aap:create:grants",
        "aap:delete:grants",
        "aap:read:publishes",
        "aap:create:publishes",
        "aap:delete:publishes",
        "aap:read:subscriptions",
        "aap:create:subscriptions",
        "aap:delete:subscriptions",
        "aap:read:consents",
        "aap:create:consents",
        "aap:delete:consents",
        "aap:read:shadows",
        "aap:create:shadows",
        "aap:delete:shadows",

        "mg:aap:read:grants",
        "mg:aap:create:grants",
        "mg:aap:delete:grants",
        "mg:aap:read:publishes",
        "mg:aap:create:publishes",
        "mg:aap:delete:publishes",
        "mg:aap:read:subscriptions",
        "mg:aap:create:subscriptions",
        "mg:aap:delete:subscriptions",
        "mg:aap:read:consents",
        "mg:aap:create:consents",
        "mg:aap:delete:consents",
        "mg:aap:read:shadows",
        "mg:aap:create:shadows",
        "mg:aap:delete:shadows",

        "0:mg:aap:read:grants",
        "0:mg:aap:create:grants",
        "0:mg:aap:delete:grants",
        "0:mg:aap:read:publishes",
        "0:mg:aap:create:publishes",
        "0:mg:aap:delete:publishes",
        "0:mg:aap:read:subscriptions",
        "0:mg:aap:create:subscriptions",
        "0:mg:aap:delete:subscriptions",
        "0:mg:aap:read:consents",
        "0:mg:aap:create:consents",
        "0:mg:aap:delete:consents",
        "0:mg:aap:read:shadows",
        "0:mg:aap:create:shadows",
        "0:mg:aap:delete:shadows",
      }

      for _, request := range iRequests {
        r := request.Input.(client.CreateEntitiesRequest)

        for _,s := range r.Scopes {
          var foundScope = false
          for _,vs := range validScopes {
            if s == vs {
              foundScope = true
            }
          }

          if !foundScope {
            bulky.FailAllRequestsWithClientOperationAbortedResponse(iRequests) // Fail all with abort
            request.Output = bulky.NewClientErrorResponse(request.Index)

            return
          }
        }

        entity, err := aap.CreateEntity(tx, aap.Identity{Id: r.Reference}, aap.Identity{Id:r.Creator}, r.Scopes)
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

        if entity.Id != "" {
          ok := client.CreateEntitiesResponse{
            Reference: entity.Id,
            Creator: r.Creator,
            Scopes: r.Scopes,
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
        request.Output = bulky.NewInternalErrorResponse(request.Index) // Specify error on failed one
        return
      }

      err = bulky.OutputValidateRequests(iRequests)
      if err == nil {
        tx.Commit()
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

func DeleteEntities(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "DeleteEntities",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{

    })
  }
  return gin.HandlerFunc(fn)
}

func GetEntities(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetEntities",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{

    })
  }
  return gin.HandlerFunc(fn)
}
