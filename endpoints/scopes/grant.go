package scopes

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/client"
  "github.com/charmixer/aap/environment"
  "github.com/charmixer/aap/gateway/aap"
)

func PostScopesGrant(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetScopes",
    })

    var input []client.ReadScopesRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var scopes []aap.Scope
    for _, e := range input {
      v := aap.Scope{
        Name: e.Scope,
      }
      scopes = append(scopes, v)
    }

    dbScopes, err := aap.ReadScopes(env.Driver, scopes)

    if err != nil {
      log.Println(err)
    }

    var output []client.ReadScopesResponse
    for _, dbScope := range dbScopes {
      v := client.ReadScopesResponse{
        Scope: dbScope.Name,
      }
      output = append(output, v)
    }

    c.AbortWithStatusJSON(http.StatusOK, output)
  }
  return gin.HandlerFunc(fn)
}

func DeleteScopesGrant(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PutGrant",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}
