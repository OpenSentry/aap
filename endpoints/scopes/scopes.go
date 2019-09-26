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

    var handleRequest = func(iRequests []*utils.Request){
      createdByIdentity := aap.Identity{
        Id: c.MustGet("sub").(string),
      }

      for _, request := range iRequests {
        r := request.Request.(client.CreateScopesRequest)

        scope := aap.Scope{
          Name: r.Scope,
          Title: r.Title,
          Description: r.Description,
        }

        // TODO handle error
        rScope, rIdentity, err := aap.CreateScope(env.Driver, scope, createdByIdentity)

        if err != nil {
          request.Response = utils.NewInternalErrorResponse(request.Index)
          log.Debug(err.Error())
          continue
        }

        ok := client.Scope{
          Scope: rScope.Name,
          Title: rScope.Title,
          Description: rScope.Description,
          CreatedBy: rIdentity.Id,
        }

        response := client.CreateScopesResponse{Ok: ok}
        response.Index = request.Index
        response.Status = http.StatusOK
        request.Response = response
      }
    }

    responses := utils.HandleBulkRestRequest(requests, handleRequest, utils.HandleBulkRequestParams{})

    c.JSON(http.StatusOK, responses)
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

    var handleRequests = func(iRequests []*utils.Request){
      var scopes []aap.Scope

      for _, request := range iRequests {
        if request.Request != nil {
          var r client.ReadScopesRequest
          r = request.Request.(client.ReadScopesRequest)

          v := aap.Scope{
            Name: r.Scope,
          }
          scopes = append(scopes, v)
        }
      }

      dbScopes, _ := aap.FetchScopes(env.Driver, scopes)

      for _, request := range iRequests {
        var r client.ReadScopesRequest
        if request.Request != nil {
          r = request.Request.(client.ReadScopesRequest)
        }

        var ok []client.Scope
        for _, d := range dbScopes {
          if request.Request != nil && d.Name != r.Scope {
            continue
          }

          ok = append(ok, client.Scope{
            Scope:       d.Name,
            Title:       d.Title,
            Description: d.Description,
            CreatedBy:   d.CreatedBy.Id,
          })
        }

        var response client.ReadScopesResponse
        response.Index = request.Index
        response.Status = http.StatusOK
        response.Ok = ok
        request.Response = response
      }
    }

    responses := utils.HandleBulkRestRequest(requests, handleRequests, utils.HandleBulkRequestParams{EnableEmptyRequest: true})

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
