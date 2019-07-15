package authorizations

import (
  "net/http"

  "github.com/gin-gonic/gin"
  "golang-cp-be/environment"
  "golang-cp-be/gateway/cpbe"
  //"golang-cp-be/gateway/hydra"
)

type AuthorizationsRequest struct {
  Id string `json:"id" binding:"required"`
}

type AuthorizationsResponse struct {
  Id string `json:"id" binding:"required"`
}

type GetAuthorizationsRequest struct {
  *AuthorizationsRequest
}

type GetAuthorizationsResponse struct {
  *AuthorizationsResponse
}

func GetCollection(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    requestId := c.MustGet(environment.RequestIdKey).(string)
    environment.DebugLog(route.LogId, "GetCollection", "", requestId)

    // FIXME: id should come from access token / identity token
    //accessToken := c.Get(environment.AccessTokenKey)
    //idToken := c.Get(environment.IdTokenKey)

    id, _ := c.GetQuery("id")
    if id == "" {
      c.JSON(http.StatusNotFound, gin.H{
        "error": "Not found. Hint: Are you missing id in request?",
      })
      c.Abort()
      return;
    }

    app, _ := c.GetQuery("app")
    if app == "" {
      c.JSON(http.StatusNotFound, gin.H{
        "error": "Not found. Hint: Are you missing app in request?",
      })
      c.Abort()
      return;
    }

    identity := cpbe.Identity{
      Subject: id,
    }
    application := cpbe.App{
      Name: app,
    }
    permissionList, err := cpbe.FetchPermissionsForIdentityForApplication(env.Driver, identity, application)
    if err == nil {

      var grantedPermissions []string
      for _, permission := range permissionList {
        grantedPermissions = append(grantedPermissions, permission.Name)
      }

      c.JSON(http.StatusOK, grantedPermissions)
      return
    }

    // Deny by default
    c.JSON(http.StatusNotFound, gin.H{
      "error": "Not found",
    })
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}

func PostCollection(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    requestId := c.MustGet(environment.RequestIdKey).(string)
    environment.DebugLog(route.LogId, "PostCollection", "", requestId)

    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}

func PutCollection(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    requestId := c.MustGet(environment.RequestIdKey).(string)
    environment.DebugLog(route.LogId, "PutCollection", "", requestId)

    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}
