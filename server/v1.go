package server

import (
  "github.com/gin-gonic/gin"
  v1 "golang-cp-be/controller/v1"
)

func V1Routes(r *gin.RouterGroup) {
  r.POST( "/authorizations/authorize", v1.PostAuthorizationsAuthorize)
  r.GET( "/authorizations/authorize", v1.GetAuthorizationsAuthorize)

  r.GET( "/authorizations", v1.GetAuthorizations)
  r.POST("/authorizations", v1.PostAuthorizations)
  r.PUT( "/authorizations", v1.PutAuthorizations)
}
