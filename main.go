package main

import (
  "github.com/gin-gonic/gin"
  "golang-cp-be/config"
  "golang-cp-be/controller"
)

/*

/authorizations

*/

func init() {
  config.InitConfigurations()
}

func main() {

  r := gin.Default()

  r.POST( "/authorizations/authorize", controller.PostAuthorizationsAuthorize)
  r.GET( "/authorizations/authorize", controller.GetAuthorizationsAuthorize)

  r.GET( "/authorizations", controller.GetAuthorizations)
  r.POST("/authorizations", controller.PostAuthorizations)
  r.PUT( "/authorizations", controller.PutAuthorizations)

  r.Run() // listen and serve on 0.0.0.0:8080
}
