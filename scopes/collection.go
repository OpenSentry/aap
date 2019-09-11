package scopes

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/environment"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"
)

func PostScopes(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetScopes",
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
      Title: input.Title,
      Description: input.Description,
    }

    _, err = aap.CreateScope(env.Driver, scope)

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

func GetScopes(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostScopes",
    })

    var input client.ReadScopesRequest
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

func PutScopes(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PutScopes",
    })

    c.JSON(http.StatusOK, gin.H{

    })
  }
  return gin.HandlerFunc(fn)
}
