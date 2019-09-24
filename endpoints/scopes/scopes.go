package scopes

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"


  "github.com/charmixer/aap/utils"
  "fmt"

  "github.com/charmixer/aap/environment"
  "github.com/charmixer/aap/gateway/aap"
  "github.com/charmixer/aap/client"
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
      log.Debug(err.Error())
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var handleData = func(iRequests interface{}) (interface{}) {
      requests := iRequests.([]client.ReadScopesRequest)

      var scopes []aap.Scope
      for _, request := range requests {
        v := aap.Scope{
          Name: request.Scope,
        }
        scopes = append(scopes, v)
      }

      dbScopes, _ := aap.FetchScopes(env.Driver, scopes)

      return dbScopes
    }

    var handleRequest = func(index int, iRequest interface{}, iData interface{}) (interface{}){
      var request client.ReadScopesRequest
      var response client.ReadScopesResponse

      if iRequest != nil {
        request = iRequest.(client.ReadScopesRequest)
      }

      scopes := iData.([]aap.Scope)

      var r []client.Scope
      for _, d := range scopes {
        if iRequest != nil && d.Name != request.Scope {
          continue
        }

        r = append(r, client.Scope{
          Scope: d.Name,
          Title: d.Title,
          Description: d.Description,
          CreatedBy: d.CreatedBy.Id,
        })
      }

      response.Index = index
      response.Status = http.StatusOK
      response.Ok = r

      return response
    }

    responses := utils.HandleBulkRestRequest(requests, true /* allow Ø */, handleData, handleRequest)

    c.JSON(http.StatusOK, responses)
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
