package authorizations

import (
  "net/http"
  "strings"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"
  "golang-cp-be/environment"
  "golang-cp-be/gateway/aapapi"
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

type CreateConsentRequest struct {
  Subject                string   `json:"sub" binding:"required"`
  ClientId               string   `json:"client_id" binding:"required"`
  ResourceServerClientId string   `json:"client_id,omitempty" binding:"required"`
  GrantedScopes          []string `json:"granted_scopes,omitempty"`
  RevokedScopes          []string `json:"revoked_scopes,omitempty"`
  RequestedScopes        []string `json:"requested_scopes,omitempty"`
}

func GetCollection(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetCollection",
    })

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

    resourceOwner := aapapi.Identity{
      Subject: id,
    }
    client := aapapi.Client{
      ClientId: clientId,
    }
    permissionList, err := aapapi.FetchConsentsForResourceOwnerToClient(env.Driver, resourceOwner, client, permissions)
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
      "func": "PostCollection",
    })

    var input CreateConsentRequest
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

    resourceOwner := aapapi.Identity{
      Subject: input.Subject,
    }
    client := aapapi.Client{
      ClientId: input.ClientId,
    }
    resourceServer := aapapi.Client{
      ClientId: input.ResourceServerClientId,
    }
    permissionList, err := aapapi.CreateConsentsToResourceServerForClientOnBehalfOfResourceOwner(env.Driver, resourceOwner, client, resourceServer, grantPermissions, revokePermissions)
    if err != nil {
      log.Debug(err.Error())
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
      "func": "PutCollection",
    })

    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}
