package authorizations

import (
  "net/http"
  //"strings"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"
  "golang-cp-be/environment"
  "golang-cp-be/gateway/aapapi"
)

type ConsentRequest struct {
  Subject              string `form:"sub" json:"sub" binding:"required"`
  ClientId             string `form:"client_id,omitempty" json:"client_id,omitempty" binding:"required"`
  GrantedScopes      []string `form:"granted_scopes,omitempty" json:"granted_scopes,omitempty"`
  RevokedScopes      []string `form:"revoked_scopes,omitempty" json:"revoked_scopes,omitempty"`
  RequestedScopes    []string `form:"requested_scopes,omitempty" json:"requested_scopes,omitempty"`
  RequestedAudiences []string `form:"requested_audiences,omitempty" json:"requested_audiences,omitempty"` // hydra.requested_access_token_audience
}

type ConsentResponse struct {

}

func GetCollection(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetCollection",
    })

    var input ConsentRequest
    err := c.Bind(&input)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      c.Abort()
      return
    }

    var requestedPermissions []aapapi.Permission
    for _, scope := range input.RequestedScopes {
      requestedPermissions = append(requestedPermissions, aapapi.Permission{ Name:scope})
    }

    resourceOwner := aapapi.Identity{
      Subject: input.Subject,
    }
    client := aapapi.Client{
      ClientId: input.ClientId,
    }

    var resourceServer *aapapi.ResourceServer = nil
    if len(input.RequestedAudiences) > 0 {
      if ( len(input.RequestedAudiences) ) > 1 {
        log.WithFields(logrus.Fields{"requested_audiences":input.RequestedAudiences}).Debug("More than one audience not supported yet")
        c.JSON(http.StatusNotFound, gin.H{
          "error": "More than one audience not supported yet Hint: Try only to use audience per token request one for now",
        })
        c.Abort()
        return
      }

      resourceServer, err = aapapi.FetchResourceServerByAudience(env.Driver, input.RequestedAudiences[0])
      if err != nil {
        log.WithFields(logrus.Fields{"aud":input.RequestedAudiences}).Debug("Resource server not found")
        c.JSON(http.StatusNotFound, gin.H{
          "error": "Not found. Hint: Maybe audience does not exist.",
        })
        c.Abort()
        return
      }
    }

    var consentList []aapapi.Consent
    if resourceServer != nil {
      consentList, err = aapapi.FetchConsentsForResourceOwnerToClientAndResourceServer(env.Driver, resourceOwner, client, *resourceServer, requestedPermissions)
    } else {
      consentList, err = aapapi.FetchConsentsForResourceOwnerToClient(env.Driver, resourceOwner, client, requestedPermissions)
    }

    if err != nil {
      log.WithFields(logrus.Fields{"id":resourceOwner.Subject, "client_id":client.ClientId, "scope":input.RequestedScopes}).Debug(err.Error())
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

    resourceOwner := aapapi.Identity{
      Subject: input.Subject,
    }
    client := aapapi.Client{
      ClientId: input.ClientId,
    }

    var permissionList []aapapi.Permission
    if len(input.RequestedAudiences) > 0 {
      if ( len(input.RequestedAudiences) ) > 1 {
        log.WithFields(logrus.Fields{"requested_audiences":input.RequestedAudiences}).Debug("More than one audience not supported yet")
        c.JSON(http.StatusNotFound, gin.H{
          "error": "More than one audience not supported yet Hint: Try only to use audience per token request one for now",
        })
        c.Abort()
        return
      }

      resourceServer, err := aapapi.FetchResourceServerByAudience(env.Driver, input.RequestedAudiences[0])
      if err != nil {
        log.WithFields(logrus.Fields{"aud":input.RequestedAudiences}).Debug("Resource server not found")
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
