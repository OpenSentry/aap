package grants

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/app"
  "github.com/charmixer/aap/client"
  "github.com/charmixer/aap/gateway/aap"

  bulky "github.com/charmixer/bulky/server"
  "fmt"
)

func GetGrants(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
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
        if r.Identity != "" {
          iGranted.Id = r.Identity
        }

        var iPublishers []aap.Identity
        if r.Publisher != "" {
          iPublishers = []aap.Identity{
            {Id: r.Publisher},
          }
        }

        var iOnBehalfOf []aap.Identity
        if r.OnBehalfOf != "" {
          iOnBehalfOf = []aap.Identity{
            {Id: r.OnBehalfOf},
          }
        }
        var iScopes []aap.Scope
        if r.Scope != "" {
          iScopes = []aap.Scope{
            {Name: r.Scope},
          }
        }

        // TODO handle error
        grants, err := aap.FetchGrants(tx, iGranted, iScopes, iPublishers, iOnBehalfOf)

        if err != nil {
          // fail all requests
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests)

          // specify error on this request
          request.Output = bulky.NewInternalErrorResponse(request.Index)
          log.Debug(err.Error())
          return
        }

        var ok = []client.Grant{}
        for _,grant := range grants {

          var mgscopes = []string{}
          for _,mgscope := range grant.MayGrantScopes {
            mgscopes = append(mgscopes, mgscope.Name)
          }

          fmt.Printf("%#v", grant)

          ok = append(ok, client.Grant{
            Identity: grant.Identity.Id,
            Scope: grant.Scope.Name,
            Publisher: grant.Publisher.Id,
            OnBehalfOf: grant.OnBehalfOf.Id,
            MayGrantScopes: mgscopes,
            NotBefore: grant.GrantRule.NotBefore,
            Expire: grant.GrantRule.Expire,
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

func PostGrants(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
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

        iReceive := aap.Identity{
          Id: r.Identity,
        }

        iScope := aap.Scope{
          Name: r.Scope,
        }

        iPublishedBy := aap.Identity{
          Id: r.Publisher,
        }

        iOnBehalfOf := aap.Identity{
          Id: r.OnBehalfOf,
        }

        // TODO handle error
        grant, err := aap.CreateGrant(tx, iReceive, iScope, iPublishedBy, iOnBehalfOf, r.NotBefore, r.Expire)

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

        ok := client.Grant{
          Identity: grant.Identity.Id,
          Scope: grant.Scope.Name,
          Publisher: grant.Publisher.Id,
          OnBehalfOf: grant.OnBehalfOf.Id,
          NotBefore: grant.GrantRule.NotBefore,
          Expire: grant.GrantRule.Expire,
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

func DeleteGrants(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "DeleteGrants",
    })

    var requests []client.DeleteGrantsRequest
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

      requestor := c.MustGet("sub").(string)

      for _, request := range iRequests {
        r := request.Input.(client.DeleteGrantsRequest)

        log = log.WithFields(logrus.Fields{"id": requestor})

        dbGrants, err := aap.FetchGrants(tx, aap.Identity{Id:r.Identity}, []aap.Scope{{Name: r.Scope}}, []aap.Identity{{Id:r.Publisher}}, []aap.Identity{{Id:r.OnBehalfOf}}  )
        if err != nil {
          request.Output = bulky.NewInternalErrorResponse(request.Index)
          log.Debug(err.Error())
          return
        }

        if len(dbGrants) <= 0  {
          // not found translate into already deleted
          ok := client.DeleteGrantsResponse{}
          request.Output = bulky.NewOkResponse(request.Index, ok)
          continue;
        }
        grantToDelete := dbGrants[0]

        if grantToDelete.Identity.Id != "" {

          err := aap.DeleteGrant(tx, grantToDelete)
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

          ok := client.DeleteGrantsResponse{}
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
        log.Debug("Delete grant failed. Hint: Maybe input validation needs to be improved.")
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
