package subscriptions

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/opensentry/aap/app"
  "github.com/opensentry/aap/gateway/aap"
  "github.com/opensentry/aap/client"

  bulky "github.com/charmixer/bulky/server"
)

func PostSubscriptions(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostSubscriptions",
    })

    var requests []client.CreateSubscriptionsRequest
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

      var clients []string

      for _, request := range iRequests {
        r := request.Input.(client.CreateSubscriptionsRequest)

        iSubscription := aap.Subscription{
          Subscriber: aap.Identity{Id:r.Subscriber},
          Publisher: aap.Identity{Id:r.Publisher},
          Scope: aap.Scope{Name:r.Scope},
        }
        rSubscription, err := aap.CreateSubscription(tx, iSubscription, aap.Identity{Id:requestor})
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

        if rSubscription.Subscriber.Id != "" {
          ok := client.CreateSubscriptionsResponse{
            Subscriber: rSubscription.Subscriber.Id,
            Publisher: rSubscription.Publisher.Id,
            Scope: rSubscription.Scope.Name,
          }
          request.Output = bulky.NewOkResponse(request.Index, ok)

          clients = append(clients, rSubscription.Subscriber.Id)
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

        readSession, readTx, err := aap.BeginReadTx(env.Driver)
        if err != nil {
          log.Debug(err.Error())
          return
        }
        defer readTx.Close() // rolls back if not already committed/rolled back
        defer readSession.Close()

        for _,id := range clients {
          aap.SyncScopesToHydra(readTx, aap.Identity{Id:id}) // fire and forget to hydra
        }

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

func DeleteSubscriptions(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "DeleteSubscriptions",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{

    })
  }
  return gin.HandlerFunc(fn)
}

func GetSubscriptions(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetGrants",
    })

    var requests []client.ReadSubscriptionsRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var handleRequest = func(iRequests []*bulky.Request){
      //iRequest := aap.Identity{
        //Id: c.MustGet("sub").(string),
      //}

      session, tx, err := aap.BeginReadTx(env.Driver)

      if err != nil {
        bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
        log.Debug(err.Error())
        return
      }

      defer tx.Close() // rolls back if not already committed/rolled back
      defer session.Close()

      for _, request := range iRequests {
        var r client.ReadSubscriptionsRequest
        if request.Input != nil {
          r = request.Input.(client.ReadSubscriptionsRequest)
        }

        var iFilterSubscriber aap.Identity
        if r.Subscriber != "" {
          iFilterSubscriber = aap.Identity{Id: r.Subscriber}
        }

        var iFilterPublisher aap.Identity
        if r.Publisher != "" {
          iFilterPublisher = aap.Identity{Id: r.Publisher}
        }

        var iFilterScopes []aap.Scope
        if r.Scopes != nil {
          for _,scopeName := range r.Scopes {
            iFilterScopes = append(iFilterScopes, aap.Scope{Name: scopeName})
          }
        }

        // TODO handle error
        subscriptions, err := aap.FetchSubscriptions(tx, iFilterSubscriber, iFilterPublisher, iFilterScopes)

        if err != nil {
          // fail all requests
          bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests)

          // specify error on this request
          request.Output = bulky.NewInternalErrorResponse(request.Index)
          log.Debug(err.Error())
          return
        }

        var ok = client.ReadSubscriptionsResponse{}
        for _,subscription := range subscriptions {
          ok = append(ok, client.Subscription{
            Subscriber: subscription.Subscriber.Id,
            Scope: subscription.Scope.Name,
            Publisher: subscription.Publisher.Id,
          })
        }

        request.Output = bulky.NewOkResponse(request.Index, ok)
      }
    }

    responses := bulky.HandleRequest(requests, handleRequest, bulky.HandleRequestParams{EnableEmptyRequest: true})

    c.JSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}
