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
  RequestedAudiences []string `json:"requested_audiences,omitempty"`
}

type ConsentResponse struct {

}

type CreateConsentRequest struct {
  Subject                string   `json:"sub" binding:"required"`
  ClientId               string   `json:"client_id" binding:"required"`
  Audience               string   `json:"aud" binding:"required"`
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
    resourceServer := aapapi.ResourceServer{
      Name: "idpapi",
    }
    consentList, err := aapapi.FetchConsentsForResourceOwnerToClientAndResourceServer(env.Driver, resourceOwner, client, resourceServer, permissions)
    if err != nil {
      log.WithFields(logrus.Fields{"id":resourceOwner.Subject, "client_id":client.ClientId, "scope":requestedScopes}).Debug(err.Error())
      c.JSON(http.StatusInternalServerError, gin.H{
        "error": "Unable to fetch consents",
      })
      c.Abort()
      return
    }

    //if len(consentList) > 0 {
      var consentedPermissions []string
      for _, consent := range consentList {
        consentedPermissions = append(consentedPermissions, consent.Permission.Name)
      }
      c.JSON(http.StatusOK, consentedPermissions)
      return
    //}

    // Deny by default
    /*c.JSON(http.StatusNotFound, gin.H{
      "error": "Not found",
    })
    c.Abort()*/
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

    var permissionList []aapapi.Permission
    if input.Audience != "" {
      resourceServer, err := aapapi.FetchResourceServerByAudience(env.Driver, input.Audience)
      if err != nil {
        log.WithFields(logrus.Fields{"aud":input.Audience}).Debug("Resource server not found")
        c.JSON(http.StatusNotFound, gin.H{
          "error": "Not found. Hint: Maybe audience does not exist.",
        })
        c.Abort()
        return
      }
      permissionList, err = aapapi.CreateConsentsToResourceServerForClientOnBehalfOfResourceOwner(env.Driver, resourceOwner, client, *resourceServer, grantPermissions, revokePermissions)
    } else {
      permissionList, err = aapapi.CreateConsentsForClientOnBehalfOfResourceOwner(env.Driver, resourceOwner, client, grantPermissions, revokePermissions)
    }
    if err != nil {
      log.Debug(err.Error())
    }

    if len(permissionList) > 0 {
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
