package entities

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/app"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"

  bulky "github.com/charmixer/bulky/server"
)

func GetEntitiesJudge(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetEntitiesJudge",
    })

    var requests []client.ReadEntitiesJudgeRequest
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

      // requestor := c.MustGet("sub").(string) // This is the requestor of the Judge call. Not the client we need to judge.

      for _, request := range iRequests {
        r := request.Input.(client.ReadEntitiesJudgeRequest)

        iPublisher := aap.Identity{Id:r.Publisher}
        iRequestor := aap.Identity{Id:r.Identity}
        iScope     := aap.Scope{Name:r.Scope}

        var iOwners []aap.Identity
        for _, id := range r.Owners {
          iOwners = append(iOwners, aap.Identity{Id:id})
        }

        verdict, err := aap.Judge(tx, iPublisher, iRequestor, iScope, iOwners)
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


        var owners []string
        for _, o := range verdict.Owners {
          owners = append(owners, o.Id)
        }

        request.Output = bulky.NewOkResponse(request.Index, client.ReadEntitiesJudgeResponse{
          Granted: verdict.Granted,
          Identity: verdict.Requestor.Id,
          Publisher: verdict.Publisher.Id,
          Scope: verdict.Scope.Name,
          Owners: owners,
        })
      }

      err = bulky.OutputValidateRequests(iRequests)
      if err == nil {
        tx.Commit()
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
