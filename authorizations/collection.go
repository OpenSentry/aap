package authorizations

import (
  "net/http"

  "golang.org/x/oauth2/clientcredentials"

  oidc "github.com/coreos/go-oidc"

  "github.com/gin-gonic/gin"
  "golang-cp-be/gateway/hydra"
)

type CpBeEnv struct {
  Provider *oidc.Provider
  HydraConfig *clientcredentials.Config
  HydraClient *hydra.HydraClient
}

func GetCollection(env *CpBeEnv) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}

func PostCollection(env *CpBeEnv) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}

func PutCollection(env *CpBeEnv) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}
