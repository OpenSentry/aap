package scopes

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/app"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"

  bulky "github.com/charmixer/bulky/server"
)

func PostScopes(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostScopes",
    })

    var requests []client.CreateScopesRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var handleRequest = func(iRequests []*bulky.Request){
      createdByIdentity := aap.Identity{
        Id: c.MustGet("sub").(string),
      }

      session, tx, err := aap.BeginWriteTx(env.Driver)
      if err != nil {
        bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
        log.Debug(err.Error())
        return
      }
      defer tx.Close() // rolls back if not already committed/rolled back
      defer session.Close()

      for _, request := range iRequests {
        r := request.Input.(client.CreateScopesRequest)

        scope := aap.Scope{
          Name: r.Scope,
        }

        // TODO handle error
        rScope, err := aap.CreateScope(tx, scope, createdByIdentity)

        if err != nil {
          tx.Rollback()

          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests)
          request.Output = bulky.NewInternalErrorResponse(request.Index)

          log.Debug(err.Error())
          return
        }

        ok := client.CreateScopesResponse{
          Scope: rScope.Name,
        }

        request.Output = bulky.NewOkResponse(request.Index, ok)
      }

      // should be deny by default
      tx.Commit()
    }

    responses := bulky.HandleRequest(requests, handleRequest, bulky.HandleRequestParams{})

    c.JSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}

func GetScopes(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetScopes",
    })

    var requests []client.ReadScopesRequest
    err := c.BindJSON(&requests)
    if err != nil {
      log.Debug(err.Error())
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var handleRequests = func(iRequests []*bulky.Request){
      var scopes []aap.Scope

      for _, request := range iRequests {
        if request.Input != nil {
          var r client.ReadScopesRequest
          r = request.Input.(client.ReadScopesRequest)

          v := aap.Scope{
            Name: r.Scope,
          }
          scopes = append(scopes, v)
        }
      }

      dbScopes, _ := aap.FetchScopes(env.Driver, scopes)

      for _, request := range iRequests {
        var r client.ReadScopesRequest
        if request.Input != nil {
          r = request.Input.(client.ReadScopesRequest)
        }

        var ok client.ReadScopesResponse
        for _, d := range dbScopes {
          if request.Input != nil && d.Name != r.Scope {
            continue
          }

          ok = append(ok, client.Scope{
            Scope:       d.Name,
          })
        }

        request.Output = bulky.NewOkResponse(request.Index, ok)
      }
    }

    responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{EnableEmptyRequest: true})

    c.JSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}

func PutScopes(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostScopes",
    })

    var requests []client.UpdateScopesRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    c.AbortWithStatus(http.StatusNotFound)
  }
  return gin.HandlerFunc(fn)
}
