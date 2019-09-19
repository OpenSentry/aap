package scopes

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/aap/environment"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"
  "fmt"
)

func PostScopes(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostScopes",
    })

    var requests []client.CreateScopesRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    fmt.Println(c.Request.Header)
    var createdByIdentity aap.Identity
    createdByIdentity = aap.Identity{
      Id: c.MustGet("sub").(string),
    }

    var responses []client.CreateScopesResponse
    for _, request := range requests {
      scope := aap.Scope{
        Name:        request.Scope,
        Title:       request.Title,
        Description: request.Description,
      }

      rScope, rIdentity, err := aap.CreateScope(env.Driver, scope, createdByIdentity)
      fmt.Println(rScope, rIdentity, err)

      if err != nil {
        log.Println(err)
        c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
      }

      responses = append(responses, client.CreateScopesResponse{
        Scope:       rScope.Name,
        Title:       rScope.Title,
        Description: rScope.Description,
        CreatedBy:   rIdentity.Id,
      })
    }

    c.AbortWithStatusJSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}

func GetScopes(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetScopes",
    })

    var requests []client.ReadScopesRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var scopes []aap.Scope
    for _, e := range requests {
      v := aap.Scope{
        Name: e.Scope,
      }
      scopes = append(scopes, v)
    }

    dbScopes, err := aap.FetchScopes(env.Driver, scopes)

    if err != nil {
      log.Println(err)
    }

    var output []client.ReadScopesResponse
    for _, dbScope := range dbScopes {
      v := client.ReadScopesResponse{
        Scope: dbScope.Name,
        Title: dbScope.Title,
        Description: dbScope.Description,
      }
      output = append(output, v)
    }

    c.AbortWithStatusJSON(http.StatusOK, output)
  }
  return gin.HandlerFunc(fn)
}

func PutScopes(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostScopes",
    })

    var requests []client.UpdateScopesRequest
    err := c.BindJSON(&requests)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    fmt.Println(c.Request.Header)
    var createdByIdentity aap.Identity
    createdByIdentity = aap.Identity{
      Id: c.MustGet("sub").(string),
    }

    var responses []client.UpdateScopesResponse
    for _, request := range requests {
      scope := aap.Scope{
        Name:        request.Scope,
        Title:       request.Title,
        Description: request.Description,
      }

      rScope, rIdentity, err := aap.UpdateScope(env.Driver, scope, createdByIdentity)
      fmt.Println(rScope, rIdentity, err)

      if err != nil {
        log.Println(err)
        c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
      }

      responses = append(responses, client.UpdateScopesResponse{
        Scope:       rScope.Name,
        Title:       rScope.Title,
        Description: rScope.Description,
        CreatedBy:   rIdentity.Id,
      })
    }

    c.AbortWithStatusJSON(http.StatusOK, responses)
  }
  return gin.HandlerFunc(fn)
}
