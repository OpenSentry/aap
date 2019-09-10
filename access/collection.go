package access

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/environment"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"
)

func PostCollection(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetCollection",
    })

    var input client.CreateAccessRequest
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
    })
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}

func GetCollection(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostCollection",
    })

    c.JSON(http.StatusOK, gin.H{
      "scope": "scope:name",
    })
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}

func PutCollection(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PutCollection",
    })

    c.JSON(http.StatusOK, gin.H{
      "scope": "scope:name",
    })
  }
  return gin.HandlerFunc(fn)
}
