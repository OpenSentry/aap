package authorizations

import (
  "net/http"
  "strings"
  "fmt"

  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"
  "golang-cp-be/environment"
  "golang-cp-be/gateway/aapapi"
  //"golang-cp-be/gateway/hydra"
)

type ConsentRequest struct {
  Subject string `json:"sub" binding:"required"`
  ClientId string `json:"client_id,omitempty"`
  GrantedScopes []string `json:"granted_scopes,omitempty"`
  RevokedScopes []string `json:"revoked_scopes,omitempty"`
  RequestedScopes []string `json:"requested_scopes,omitempty"`
}

type ConsentResponse struct {

}

func GetCollection(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "route.logid": route.LogId,
      "component": "authorizations",
      "func": "GetCollection",
    })

    log.Debug("Received authorizations request")

    id, _ := c.GetQuery("id")
    if id == "" {
      c.JSON(http.StatusNotFound, gin.H{
        "error": "Not found. Hint: Are you missing id in request?",
      })
      c.Abort()
      return;
    }

    clientId, _ := c.GetQuery("client_id")
    if clientId == "" {
      c.JSON(http.StatusNotFound, gin.H{
        "error": "Not found. Hint: Are you missing client_id in request?",
      })
      c.Abort()
      return;
    }

    var permissions []aapapi.Permission
    requestedScopes, _ := c.GetQuery("scope")
    if requestedScopes != "" {
      scopes := strings.Split(requestedScopes, ",")
      for _, scope := range scopes {
        permissions = append(permissions, aapapi.Permission{ Name:scope,})
      }
    }

    identity := aapapi.Identity{
      Subject: id,
    }
    applicationIdentity := aapapi.Identity{
      Subject: clientId,
    }
    permissionList, err := aapapi.FetchConsentsForIdentityToApplication(env.Driver, identity, applicationIdentity, permissions)
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

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "route.logid": route.LogId,
      "component": "authorizations",
      "func": "PostCollection",
    })

    log.Debug("Received authorizations request")

    var input ConsentRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      c.Abort()
      return
    }

    if len(input.RequestedScopes) <= 0 {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Missing granted_scopes"})
      c.Abort()
      return
    }

    var grantPermissions []aapapi.Permission
    for _, scope := range input.GrantedScopes {
      grantPermissions = append(grantPermissions, aapapi.Permission{ Name:scope,})
    }

    var revokePermissions []aapapi.Permission
    for _, scope := range input.RevokedScopes {
      revokePermissions = append(revokePermissions, aapapi.Permission{ Name:scope,})
    }

    identity := aapapi.Identity{
      Subject: input.Subject,
    }
    applicationIdentity := aapapi.Identity{
      Subject: input.ClientId,
    }
    permissionList, err := aapapi.CreateConsentsForIdentityToApplication(env.Driver, identity, applicationIdentity, grantPermissions, revokePermissions)
    if err != nil {
      fmt.Println(err)
    }
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

func PutCollection(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "route.logid": route.LogId,
      "component": "authorizations",
      "func": "PutCollection",
    })

    log.Debug("Received authorizations request")

    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}
