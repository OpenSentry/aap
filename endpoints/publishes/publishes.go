package publishes

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/environment"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"

  bulky "github.com/charmixer/bulky/server"

  "fmt"
)

func PostPublishes(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostPublishesExpose",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{

    })
  }
  return gin.HandlerFunc(fn)
}

func DeletePublishes(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "DeletePublishesExpose",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{

    })
  }
  return gin.HandlerFunc(fn)
}

func GetPublishes(env *environment.State) gin.HandlerFunc {
fn := func(c *gin.Context) {
  log := c.MustGet(environment.LogKey).(*logrus.Entry)
  log = log.WithFields(logrus.Fields{
    "func": "GetPublishes",
  })

  var requests []client.ReadPublishesRequest
  err := c.BindJSON(&requests)
  if err != nil {
    log.Debug(err.Error())
    c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  var handleRequests = func(iRequests []*bulky.Request){
    var identities []aap.Identity

    for _, request := range iRequests {
      if request.Input != nil {
        var r client.ReadPublishesRequest
        r = request.Input.(client.ReadPublishesRequest)

        v := aap.Identity{
          Id: r.Publisher,
        }
        identities = append(identities, v)
      }
    }

    session, tx, err := aap.BeginReadTx(env.Driver)

    if err != nil {
      bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
      log.Debug(err.Error())
      return
    }

    defer tx.Close() // rolls back if not already committed/rolled back
    defer session.Close()

    dbPublishes, _ := aap.FetchPublishes(tx, identities)

    for _, request := range iRequests {
      var r client.ReadPublishesRequest
      if request.Input != nil {
        r = request.Input.(client.ReadPublishesRequest)
      }

      fmt.Println(r)

      var ok client.ReadPublishesResponse
      for _, _ = range dbPublishes {
        //if request.Input != nil && db.Id != r.Publisher {
          //continue
        //}

        //ok = append(ok, client.Publish{
          //Scope:       db.Name,
        //})
      }

      request.Output = bulky.NewOkResponse(request.Index, ok)
    }
  }

  responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{EnableEmptyRequest: true})

  c.JSON(http.StatusOK, responses)
}
return gin.HandlerFunc(fn)
}
