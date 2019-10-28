package consents

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/app"
  _ "github.com/charmixer/aap/config"
  _ "github.com/charmixer/aap/gateway/aap"
  _ "github.com/charmixer/aap/client"
)

func GetConsents(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetConsents",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}

func PostConsents(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostScopesConsent",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{

    })
  }
  return gin.HandlerFunc(fn)
}

func DeleteConsents(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "DeleteScopesConsent",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{

    })
  }
  return gin.HandlerFunc(fn)
}
