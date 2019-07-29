package main

import (
  //"fmt"
  "strings"
  "net/http"
  "net/url"

  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/clientcredentials"

  oidc "github.com/coreos/go-oidc"
  "github.com/gin-gonic/gin"
  "github.com/atarantini/ginrequestid"

  "github.com/neo4j/neo4j-go-driver/neo4j"

  "golang-cp-be/config"
  "golang-cp-be/environment"
  //"golang-cp-be/gateway/hydra"
  "golang-cp-be/authorizations"
)

const app = "cpbe"

func init() {
  config.InitConfigurations()
}

func main() {

  // https://medium.com/neo4j/neo4j-go-driver-is-out-fbb4ba5b3a30
  // Each driver instance is thread-safe and holds a pool of connections that can be re-used over time. If you don’t have a good reason to do otherwise, a typical application should have a single driver instance throughout its lifetime.
  driver, err := neo4j.NewDriver(config.App.Neo4j.Uri, neo4j.BasicAuth(config.App.Neo4j.Username, config.App.Neo4j.Password, ""), func(config *neo4j.Config) {
    config.Log = neo4j.ConsoleLogger(neo4j.DEBUG)
  });
  if err != nil {
    environment.DebugLog(app, "main", "[database:Neo4j] " + err.Error(), "")
    return
  }
  defer driver.Close()

  provider, err := oidc.NewProvider(context.Background(), config.Discovery.Hydra.Public.Url + "/")
  if err != nil {
    environment.DebugLog(app, "main", "[provider:hydra] " + err.Error(), "")
    return
  }

  // Setup the hydra client cpbe is going to use (oauth2 client credentials flow)
  hydraConfig := &clientcredentials.Config{
    ClientID:     config.App.Oauth2.Client.Id,
    ClientSecret: config.App.Oauth2.Client.Secret,
    TokenURL:     config.Discovery.Hydra.Public.Endpoints.Oauth2Token,
    Scopes:       config.App.Oauth2.Scopes.Required,
    EndpointParams: url.Values{"audience": {"hydra"}},
    AuthStyle: 2, // https://godoc.org/golang.org/x/oauth2#AuthStyle
  }

  // Setup app state variables. Can be used in handler functions by doing closures see exchangeAuthorizationCodeCallback
  env := &environment.State{
    Provider: provider,
    HydraConfig: hydraConfig,
    Driver: driver,
  }

  // Setup routes to use, this defines log for debug log
  routes := map[string]environment.Route{
    "/authorizations": environment.Route{
       URL: "/authorizations",
       LogId: "cpbe://authorizations",
    },
    "/authorizations/authorize": environment.Route{
      URL: "/authorizations/authorize",
      LogId: "cpfe://authorizations/authorize",
    },
    "/authorizations/reject": environment.Route{
      URL: "/authorizations/reject",
      LogId: "cpfe://authorizations/reject",
    },
  }

  r := gin.Default()
  r.Use(ginrequestid.RequestId())

  // ## QTNA - Questions that need answering before granting access to a protected resource
  // 1. Is the user or client authenticated? Answered by the process of obtaining an access token.
  // 2. Is the access token expired?
  // 3. Is the access token granted the required scopes?
  // 4. Is the user or client giving the grants in the access token authorized to operate the scopes granted?
  // 5. Is the access token revoked?

  // All requests need to be authenticated.
  r.Use(authenticationRequired())

  r.GET(routes["/authorizations"].URL, authorizationRequired(routes["/authorizations"], "cpbe.authorizations.get"), authorizations.GetCollection(env, routes["/authorizations"]))
  r.POST(routes["/authorizations"].URL, authorizationRequired(routes["/authorizations"], "cpbe.authorizations.post"), authorizations.PostCollection(env, routes["/authorizations"]))
  r.PUT(routes["/authorizations"].URL, authorizationRequired(routes["/authorizations"], "cpbe.authorizations.update"), authorizations.PutCollection(env, routes["/authorizations"]))

  r.POST(routes["/authorizations/authorize"].URL, authorizationRequired(routes["/authorizations/authorize"], "cpbe.authorize"), authorizations.PostAuthorize(env, routes["/authorizations/authorize"]))
  r.POST(routes["/authorizations/reject"].URL, authorizationRequired(routes["/authorizations/reject"], "cpbe.reject"), authorizations.PostReject(env, routes["/authorizations/reject"]))

  r.RunTLS(":" + config.App.Serve.Public.Port, config.App.Serve.Tls.Cert.Path, config.App.Serve.Tls.Key.Path)
}

func authenticationRequired() gin.HandlerFunc {
  fn := func(c *gin.Context) {
    requestId := c.MustGet(environment.RequestIdKey).(string)
    environment.DebugLog(app, "authenticationRequired", "Checking Authorization: Bearer <token> in request", requestId)

    var token *oauth2.Token
    auth := c.Request.Header.Get("Authorization")
    split := strings.SplitN(auth, " ", 2)
    if len(split) == 2 || strings.EqualFold(split[0], "bearer") {
      environment.DebugLog(app, "authenticationRequired", "Authorization: Bearer <token> found for request.", requestId)
      token = &oauth2.Token{
        AccessToken: split[1],
        TokenType: split[0],
      }

      if token.Valid() == true {
        environment.DebugLog(app, "authenticationRequired", "Valid access token", requestId)
        c.Set(environment.AccessTokenKey, token)
        c.Next() // Authentication successful, continue.
        return;
      }

      // Deny by default
      environment.DebugLog(app, "authenticationRequired", "Invalid Access token", requestId)
      c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token."})
      c.Abort()
      return
    }

    // Deny by default
    environment.DebugLog(app, "authenticationRequired", "Missing access token", requestId)
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization: Bearer <token> not found in request."})
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}

func authorizationRequired(route environment.Route, requiredScopes ...string) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    requestId := c.MustGet(environment.RequestIdKey).(string)
    environment.DebugLog(app, "authorizationRequired", "Checking Authorization: Bearer <token> in request", requestId)

    _ /*(accessToken)*/, accessTokenExists := c.Get(environment.AccessTokenKey)
    if accessTokenExists == false {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "No access token found. Hint: Is bearer token missing?"})
      c.Abort()
      return
    }

    // Sanity check: Claims
    //fmt.Println(accessToken)

    foundRequiredScopes := true
    if foundRequiredScopes {
      environment.DebugLog(app, "authorizationRequired", "Valid scopes. WE DID NOT CHECK IT - TODO!", requestId)
      c.Next() // Authentication successful, continue.
      return;
    }

    // Deny by default
    environment.DebugLog(app, "authorizationRequired", "Missing required scopes: ", requestId)
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing required scopes: "})
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}
