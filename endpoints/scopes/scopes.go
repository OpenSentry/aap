package scopes

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

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

    var input client.CreateScopesRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var scope aap.Scope
    scope = aap.Scope{
      Name: input.Scope,
      Title: input.Title,
      Description: input.Description,
    }

    var createdByIdentity aap.Identity
    createdByIdentity = aap.Identity{
      Id: "root", // TODO FIXME
    }

    scope, identity, err := aap.CreateScope(env.Driver, scope, createdByIdentity)

    log.Println(scope, identity)

    if err != nil {
      log.Println(err)
    }

    var output client.CreateScopesResponse
    output = client.CreateScopesResponse{
      Scope: scope.Name,
      Title: scope.Title,
      Description: scope.Description,
    }

    c.AbortWithStatusJSON(http.StatusOK, output)
  }
  return gin.HandlerFunc(fn)
}

func GetScopes(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "GetScopes",
    })

    var input []client.ReadScopesRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      return
    }

    var scopes []aap.Scope
    for _, e := range input {
      v := aap.Scope{
        Name: e.Scope,
      }
      scopes = append(scopes, v)
    }

    dbScopes, err := aap.ReadScopes(env.Driver, scopes)

    if err != nil {
      log.Println(err)
    }

    var output []client.ReadScopesResponse
    for _, dbScope := range dbScopes {
      v := client.ReadScopesResponse{
        Scope: dbScope.Name,
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
      "func": "PutScopes",
    })

    c.AbortWithStatusJSON(http.StatusOK, gin.H{

    })
  }
  return gin.HandlerFunc(fn)
}
