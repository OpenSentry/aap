package entities

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/environment"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"

  bulky "github.com/charmixer/bulky/server"
)

func GetEntitiesJudge(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
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
        iRequestor := aap.Identity{Id:r.Requestor}

        grantedOwnersCount := 0

        for _, id := range r.Owners {
          iOwner := aap.Identity{Id:id}

          var grantedScopes []aap.Scope
          var missingScopes []aap.Scope

          for _, scope := range r.Scopes {
            iScope := aap.Scope{Name:scope}

            verdict, err := aap.JudgeEntity(tx, iPublisher, iRequestor, iOwner, iScope)
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

            if verdict.Granted == true {
              grantedScopes = append(grantedScopes, verdict.Scope)
            } else {
              missingScopes = append(missingScopes, verdict.Scope)
            }

          }

          if len(missingScopes) == 0 && len(grantedScopes) == len(r.Scopes) {
            // Granted
            grantedOwnersCount += 1
          }

        }

        if grantedOwnersCount == len(r.Owners) {
          request.Output = bulky.NewOkResponse(request.Index, client.ReadEntitiesJudgeResponse{
            Granted: true,
          })
          continue
        }

        // Deny by default
        request.Output = bulky.NewOkResponse(request.Index, client.ReadEntitiesJudgeResponse{
          Granted: false,
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
