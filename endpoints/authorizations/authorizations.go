package authorizations

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/app"
  "github.com/charmixer/aap/client"
  "github.com/charmixer/aap/gateway/aap"
)

func GetAuthorizations(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetAuthorizations",
    })

    var input client.ConsentRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var requestedPermissions []aap.Scope
    for _, scope := range input.RequestedScopes {
      requestedPermissions = append(requestedPermissions, aap.Scope{ Name:scope})
    }

    resourceOwner := aap.Identity{
      Id: input.Subject,
    }
    client := aap.Client{
      ClientId: input.ClientId,
    }

    var resourceServer *aap.ResourceServer = nil
    if len(input.RequestedAudiences) > 0 {
      if ( len(input.RequestedAudiences) ) > 1 {
        log.WithFields(logrus.Fields{"requested_audiences":input.RequestedAudiences}).Debug("More than one audience not supported yet")
        c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
          "error": "More than one audience not supported yet Hint: Try only to use audience per token request one for now",
        })
        return
      }

      resourceServer, err = aap.FetchResourceServerByAudience(env.Driver, input.RequestedAudiences[0])
      if err != nil {
        log.WithFields(logrus.Fields{"aud":input.RequestedAudiences}).Debug("Resource server not found")
        c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
          "error": "Not found. Hint: Maybe audience does not exist.",
        })
        return
      }
    }

    var consentList []aap.Consent
    if resourceServer != nil {
      consentList, err = aap.FetchConsentsForResourceOwnerToClientAndResourceServer(env.Driver, resourceOwner, client, *resourceServer, requestedPermissions)
    } else {
      consentList, err = aap.FetchConsentsForResourceOwnerToClient(env.Driver, resourceOwner, client, requestedPermissions)
    }

    if err != nil {
      log.WithFields(logrus.Fields{"id":resourceOwner.Id, "client_id":client.ClientId, "scope":input.RequestedScopes}).Debug(err.Error())
      c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
        "error": "Unable to fetch consents",
      })
      return
    }

    //if len(consentList) > 0 {
      var consentedPermissions []string
      for _, consent := range consentList {
        consentedPermissions = append(consentedPermissions, consent.Scope.Name)
      }
      c.AbortWithStatusJSON(http.StatusOK, consentedPermissions)
      return
    //}

    // Deny by default
    /*c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
      "error": "Not found",
    })
    c.Abort()*/
  }
  return gin.HandlerFunc(fn)
}

func PostAuthorizations(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostAuthorizations",
    })

    var input client.ConsentRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    if len(input.RequestedScopes) <= 0 {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing granted_scopes"})
      return
    }

    var grantPermissions []aap.Scope
    for _, scope := range input.GrantedScopes {
      grantPermissions = append(grantPermissions, aap.Scope{ Name:scope,})
    }

    var revokePermissions []aap.Scope
    for _, scope := range input.RevokedScopes {
      revokePermissions = append(revokePermissions, aap.Scope{ Name:scope,})
    }

    resourceOwner := aap.Identity{
      Id: input.Subject,
    }
    client := aap.Client{
      ClientId: input.ClientId,
    }

    var permissionList []aap.Scope
    if len(input.RequestedAudiences) > 0 {
      if ( len(input.RequestedAudiences) ) > 1 {
        log.WithFields(logrus.Fields{"requested_audiences":input.RequestedAudiences}).Debug("More than one audience not supported yet")
        c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
          "error": "More than one audience not supported yet Hint: Try only to use audience per token request one for now",
        })
        return
      }

      resourceServer, err := aap.FetchResourceServerByAudience(env.Driver, input.RequestedAudiences[0])
      if err != nil {
        log.WithFields(logrus.Fields{"aud":input.RequestedAudiences}).Debug("Resource server not found")
        c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
          "error": "Not found. Hint: Maybe audience does not exist.",
        })
        return
      }
      permissionList, err = aap.CreateConsentsToResourceServerForClientOnBehalfOfResourceOwner(env.Driver, resourceOwner, client, *resourceServer, grantPermissions, revokePermissions)
    } else {
      permissionList, err = aap.CreateConsentsForClientOnBehalfOfResourceOwner(env.Driver, resourceOwner, client, grantPermissions, revokePermissions)
    }
    if err != nil {
      log.Debug(err.Error())
    }

    if len(permissionList) > 0 {
      var grantedPermissions []string
      for _, permission := range permissionList {
        grantedPermissions = append(grantedPermissions, permission.Name)
      }

      c.AbortWithStatusJSON(http.StatusOK, grantedPermissions)
      return
    }

    // Deny by default
    c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
      "error": "Not found. Hint does the client exists?",
    })
  }
  return gin.HandlerFunc(fn)
}

func PutAuthorizations(env *app.Environment) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PutAuthorizations",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  }
  return gin.HandlerFunc(fn)
}
