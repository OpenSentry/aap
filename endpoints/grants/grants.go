package grants

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/client"
  "github.com/charmixer/aap/environment"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/utils"
  E "github.com/charmixer/aap/client/errors"
)

func GetGrants(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetGrants",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}

func PostGrants(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostGrants",
    })

    var requests []client.CreateGrantsRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var handleRequest = func(iRequests []*utils.Request){
      iRequest := aap.Identity{
        Id: c.MustGet("sub").(string),
      }

      session, tx, err := aap.BeginWriteTx(env.Driver)

      if err != nil {
        utils.FailAllRequestsWithInternalErrorResponse(iRequests)
        log.Debug(err.Error())
        return
      }

      defer tx.Close() // rolls back if not already committed/rolled back
      defer session.Close()

      for _, request := range iRequests {
        r := request.Request.(client.CreateGrantsRequest)

        iGrant := aap.Identity{
          Id: iRequest.Id,
        }

        // no identity id provided, so use whoever requested it
        if r.IdentityId != "" {
          iGrant.Id = r.IdentityId
        }

        iPublish := aap.Identity{
          Id: r.PublishedBy,
        }

        iScope := aap.Scope{
          Name: r.Scope,
        }

        // TODO handle error
        scope, publisher, granted, granter, err := aap.CreateGrant(tx, iGrant, iScope, iPublish, iRequest)

        if err != nil {
          e := tx.Rollback()
          if e != nil {
            log.Debug(e.Error())
          }

          // fail all requests
          utils.FailAllRequestsWithInternalErrorResponse(iRequests, E.OPERATION_ABORTED)

          // specify error on this request
          request.Response = utils.NewInternalErrorResponse(request.Index)
          log.Debug(err.Error())
          return
        }

        ok := client.Grant{
          IdentityId: granted.Id,
          Scope: scope.Name,
          PublishedBy: publisher.Id,
          GrantedBy: granter.Id,
        }

        response := client.CreateGrantsResponse{Ok: ok}
        response.Index = request.Index
        response.Status = http.StatusOK
        request.Response = response
      }

      tx.Commit()
    }

    responses := utils.HandleBulkRestRequest(requests, handleRequest, utils.HandleBulkRequestParams{})

    c.JSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}

func DeleteGrants(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "DeleteGrants",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}
