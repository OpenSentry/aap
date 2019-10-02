package publishes

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/environment"
  _ "github.com/charmixer/aap/gateway/aap"
  _ "github.com/charmixer/aap/client"
)

func PostPublishes(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostScopesExpose",
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
      "func": "DeleteScopesExpose",
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
      "func": "GetConsents",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}
