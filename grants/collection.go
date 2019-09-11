package grants

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/environment"
  _ "github.com/charmixer/aap/gateway/aap"
  _ "github.com/charmixer/aap/client"
)

func GetGrants(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetConsents",
    })

    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}
