package grants

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/client"
  "github.com/charmixer/aap/environment"
  "github.com/charmixer/aap/gateway/aap"

  bulky "github.com/charmixer/bulky/server"
)

func GetGrants(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetGrants",
    })

    var requests []client.ReadGrantsRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var handleRequest = func(iRequests []*bulky.Request){
      iRequest := aap.Identity{
        Id: c.MustGet("sub").(string),
      }

      session, tx, err := aap.BeginReadTx(env.Driver)

      if err != nil {
        bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
        log.Debug(err.Error())
        return
      }

      defer tx.Close() // rolls back if not already committed/rolled back
      defer session.Close()

      for _, request := range iRequests {
        var r client.ReadGrantsRequest
        if request.Input != nil {
          r = request.Input.(client.ReadGrantsRequest)
        }

        iGranted := aap.Identity{
          Id: iRequest.Id,
        }
        // if identity id is given, use this instead
        if r.IdentityId != "" {
          iGranted.Id = r.IdentityId
        }

        var iPublisher []aap.Identity
        if r.PublishedBy != "" {
          iPublisher = []aap.Identity{
            {Id: r.PublishedBy},
          }
        }

        var iScopes []aap.Scope
        if r.Scope != "" {
          iScopes = []aap.Scope{
            {Name: r.Scope},
          }
        }

        // TODO handle error
        grants, err := aap.FetchGrants(tx, iGranted, iScopes, iPublisher)

        if err != nil {
          // fail all requests
          bulky.FailAllRequestsWithInternalErrorResponse(iRequests, "OPERATION_ABORTED")

          // specify error on this request
          request.Output = bulky.NewInternalErrorResponse(request.Index)
          log.Debug(err.Error())
          return
        }

        var ok = []client.Grant{}
        for _,e := range grants {
          ok = append(ok, client.Grant{
            IdentityId: e.Identity.Id,
            Scope: e.Scope.Name,
            PublishedBy: e.PublishedBy.Id,
            GrantedBy: e.GrantedBy.Id,
          })
        }

        request.Output = bulky.NewOkResponse(request.Index, ok)
      }

      tx.Commit()
    }

    responses := bulky.HandleRequest(requests, handleRequest, bulky.HandleRequestParams{EnableEmptyRequest: true})

    c.JSON(http.StatusOK, responses)
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

    var handleRequest = func(iRequests []*bulky.Request){
      iRequest := aap.Identity{
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
        r := request.Input.(client.CreateGrantsRequest)

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
          bulky.FailAllRequestsWithInternalErrorResponse(iRequests, "OPERATION_ABORTED")

          // specify error on this request
          request.Output = bulky.NewInternalErrorResponse(request.Index)
          log.Debug(err.Error())
          return
        }

        ok := client.Grant{
          IdentityId: granted.Id,
          Scope: scope.Name,
          PublishedBy: publisher.Id,
          GrantedBy: granter.Id,
        }

        request.Output = bulky.NewOkResponse(request.Index, ok)
      }

      err = bulky.OutputValidateRequests(iRequests)

      if err == nil {
        tx.Commit()
        return
      }

      // deny by default
      tx.Rollback()
    }

    responses := bulky.HandleRequest(requests, handleRequest, bulky.HandleRequestParams{})

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
