package entities

import (
  "strings"
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"
  "golang.org/x/oauth2"

  "github.com/charmixer/aap/app"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"

  hydra "github.com/charmixer/hydra/client"

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

      callerId := c.MustGet("sub").(string) // This is the requestor of the Judge call. Not the client we need to judge.

      hydraClient := hydra.NewHydraClient(env.OAuth2Delegator.Config)

      for _, request := range iRequests {
        r := request.Input.(client.ReadEntitiesJudgeRequest)

        tokenFromRequest := &oauth2.Token{
          AccessToken: r.AccessToken,
          TokenType: "bearer",
        }

        iCaller := aap.Identity{Id: callerId}
        iPublisher := aap.Identity{Id:r.Publisher}

        var scopes []string = strings.Split(r.Scope, " ")
        var iScopes []aap.Scope
        for _, scope := range scopes {
          iScopes = append(iScopes, aap.Scope{Name:scope})
        }

        var iOwners []aap.Identity
        for _, id := range r.Owners {
          iOwners = append(iOwners, aap.Identity{Id:id})
        }

        judgeVerdict, err := app.Judge(tx, tokenFromRequest, iPublisher, iScopes, iOwners, iCaller, hydraClient, env.OAuth2Delegator.IntrospectTokenUrl)
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

        var grantedScopes []string
        for _, s := range judgeVerdict.Verdict.GrantedScopes {
          grantedScopes = append(grantedScopes, s.Name)
        }

        var owners []string
        for _, o := range judgeVerdict.Verdict.Owners {
          owners = append(owners, o.Id)
        }

        request.Output = bulky.NewOkResponse(request.Index, client.ReadEntitiesJudgeResponse{
          Granted: judgeVerdict.Verdict.Granted,
          Identity: judgeVerdict.Verdict.Requestor.Id,
          Publisher: judgeVerdict.Verdict.Publisher.Id,
          Scope: strings.Join(grantedScopes, " "),
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
