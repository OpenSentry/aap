package shadows

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/app"
  "github.com/charmixer/aap/client"
  "github.com/charmixer/aap/gateway/aap"

  bulky "github.com/charmixer/bulky/server"
)

func GetShadows(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetShadows",
    })

    var requests []client.ReadShadowsRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var handleRequest = func(iRequests []*bulky.Request){
      session, tx, err := aap.BeginReadTx(env.Driver)

      if err != nil {
        bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
        log.Debug(err.Error())
        return
      }

      defer tx.Close() // rolls back if not already committed/rolled back
      defer session.Close()

      for _, request := range iRequests {
        var r client.ReadShadowsRequest
        if request.Input != nil {
          r = request.Input.(client.ReadShadowsRequest)
        }

        var iIdentities []aap.Identity
        if r.Identity != "" {
          iIdentities = []aap.Identity{
            {Id: r.Identity},
          }
        }

        var iShadows []aap.Identity
        if r.Shadow != "" {
          iShadows = []aap.Identity{
            {Id: r.Shadow},
          }
        }

        // TODO handle error
        shadows, err := aap.FetchShadows(tx, iIdentities, iShadows)

        if err != nil {
          // fail all requests
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests)

          // specify error on this request
          request.Output = bulky.NewInternalErrorResponse(request.Index)
          log.Debug(err.Error())
          return
        }

        var ok = client.ReadShadowsResponse{}
        for _,shadow := range shadows {
          ok = append(ok, client.Shadow{
            Identity: shadow.Identity.Id,
            Shadow: shadow.Shadow.Id,
            NotBefore: shadow.GrantRule.NotBefore,
            Expire: shadow.GrantRule.Expire,
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

func PostShadows(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostShadows",
    })

    var requests []client.CreateShadowsRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var handleRequest = func(iRequests []*bulky.Request){
      session, tx, err := aap.BeginWriteTx(env.Driver)

      if err != nil {
        bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
        log.Debug(err.Error())
        return
      }

      defer tx.Close() // rolls back if not already committed/rolled back
      defer session.Close()

      for _, request := range iRequests {
        r := request.Input.(client.CreateShadowsRequest)

        iIdentity := aap.Identity{
          Id: r.Identity,
        }

        iShadow := aap.Identity{
          Id: r.Shadow,
        }

        // TODO handle error
        shadow, err := aap.CreateShadow(tx, iIdentity, iShadow, r.NotBefore, r.Expire)

        if err != nil {
          e := tx.Rollback()
          if e != nil {
            log.Debug(e.Error())
          }

          // fail all requests
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests)

          // specify error on this request
          request.Output = bulky.NewInternalErrorResponse(request.Index)
          log.Debug(err.Error())
          return
        }

        ok := client.Shadow{
          Identity: shadow.Identity.Id,
          Shadow: shadow.Shadow.Id,
          NotBefore: shadow.GrantRule.NotBefore,
          Expire: shadow.GrantRule.Expire,
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

func DeleteShadows(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "DeleteShadows",
    })

    var requests []client.DeleteShadowsRequest
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
        r := request.Input.(client.DeleteShadowsRequest)

        iIdentities := []aap.Identity{
          {Id: r.Identity},
        }

        iShadows := []aap.Identity{
          {Id: r.Shadow},
        }

        dbShadows, err := aap.FetchShadows(tx, iIdentities, iShadows)
        if err != nil {
          request.Output = bulky.NewInternalErrorResponse(request.Index)
          log.Debug(err.Error())
          return
        }

        if len(dbShadows) <= 0  {
          // not found translate into already deleted
          ok := client.DeleteGrantsResponse{}
          request.Output = bulky.NewOkResponse(request.Index, ok)
          continue;
        }
        shadowToDelete := dbShadows[0]

        if shadowToDelete.Identity.Id != "" {

          err := aap.DeleteShadow(tx, shadowToDelete)
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

          ok := client.DeleteShadowsResponse{}
          request.Output = bulky.NewOkResponse(request.Index, ok)
          continue
        }

        // Deny by default
        e := tx.Rollback()
        if e != nil {
          log.Debug(e.Error())
        }
        bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
        request.Output = bulky.NewInternalErrorResponse(request.Index)
        log.Debug("Delete shadow failed. Hint: Maybe input validation needs to be improved.")
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
