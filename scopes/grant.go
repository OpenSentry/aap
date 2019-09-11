package scopes

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/client"
  "github.com/charmixer/aap/environment"
  "github.com/charmixer/aap/gateway/aap"
)

func PostScopesGrant(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PutGrant",
    })

    var input client.CreateScopesRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      c.Abort()
      return
    }

    var scope aap.Scope
    scope = aap.Scope{
      Name: input.Scope,
    }

    _, err = aap.ReadScope(env.Driver, scope)

    if err != nil {
      log.Println(err)
    }

    c.JSON(http.StatusOK, gin.H{
      "scope": scope.Name,
      "title": scope.Title,
      "description": scope.Description,
    })
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}

func DeleteScopesGrant(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PutGrant",
    })

    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}
