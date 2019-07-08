package main

import (
  "fmt"
  "strings"
  "net/http"
  "net/url"

  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/clientcredentials"

  oidc "github.com/coreos/go-oidc"
  "github.com/gin-gonic/gin"
  "github.com/atarantini/ginrequestid"

  "golang-cp-be/config"
  _ "golang-cp-be/gateway/hydra"
  "golang-cp-be/authorizations"
)

const app = "cpbe"
const accessTokenKey = "access_token"
const requestIdKey = "RequestId"

var (
  hydraClient *http.Client
)

func init() {
  config.InitConfigurations()
}

func main() {

  provider, err := oidc.NewProvider(context.Background(), config.Hydra.Url + "/")
  if err != nil {
    fmt.Println(err)
    return
  }

  // Setup the hydra client cpbe is going to use (oauth2 client credentials flow)
  hydraConfig := &clientcredentials.Config{
    ClientID:     config.CpBe.ClientId,
    ClientSecret: config.CpBe.ClientSecret,
    TokenURL:     config.Hydra.TokenUrl,
    Scopes:       config.CpBe.RequiredScopes,
    EndpointParams: url.Values{"audience": {"hydra"}},
    AuthStyle: 2, // https://godoc.org/golang.org/x/oauth2#AuthStyle
  }
  //hydraClient := hydra.NewHydraClient(hydraConfig)

  // Setup app state variables. Can be used in handler functions by doing closures see exchangeAuthorizationCodeCallback
  env := &authorizations.CpBeEnv{
    Provider: provider,
    HydraConfig: hydraConfig,
    // HydraClient: hydraClient, // Will this serialize the requests?
  }

  r := gin.Default()
  r.Use(ginrequestid.RequestId())

  // All requests need to be authenticated.
  r.Use(authenticationRequired())

  r.GET("/authorizations", authorizationRequired("cpbe.authorizations.get"), authorizations.GetCollection(env))
  r.POST("/authorizations", authorizationRequired("cpbe.authorizations.post"),authorizations.PostCollection(env))
  r.PUT("/authorizations", authorizationRequired("cpbe.authorizations.update"), authorizations.PutCollection(env))
  r.POST("/authorizations/authorize", authorizationRequired("cpbe.authorize"), authorizations.PostAuthorize(env))
  r.POST("/authorizations/reject", authorizationRequired("cpbe.reject"), authorizations.PostReject(env))

  r.RunTLS(":" + config.Self.Port, "/srv/certs/cpbe-cert.pem", "/srv/certs/cpbe-key.pem")
}

func authenticationRequired() gin.HandlerFunc {
  fn := func(c *gin.Context) {
    var requestId string = c.MustGet(requestIdKey).(string)
    debugLog(app, "authenticationRequired", "Checking Authorization: Bearer <token> in request", requestId)

    var token *oauth2.Token
    auth := c.Request.Header.Get("Authorization")
    split := strings.SplitN(auth, " ", 2)
    if len(split) == 2 || strings.EqualFold(split[0], "bearer") {
      debugLog(app, "authenticationRequired", "Authorization: Bearer <token> found for request.", requestId)
      token = &oauth2.Token{
        AccessToken: split[1],
        TokenType: split[0],
      }

      if token.Valid() == true {
        debugLog(app, "authenticationRequired", "Valid access token", requestId)
        c.Set(accessTokenKey, token)
        c.Next() // Authentication successful, continue.
        return;
      }

      // Deny by default
      debugLog(app, "authenticationRequired", "Invalid Access token", requestId)
      c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token."})
      c.Abort()
      return
    }

    // Deny by default
    debugLog(app, "authenticationRequired", "Missing access token", requestId)
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization: Bearer <token> not found in request."})
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}

func authorizationRequired(requiredScopes ...string) gin.HandlerFunc {
  fn := func(c *gin.Context) {
    var requestId string = c.MustGet(requestIdKey).(string)
    debugLog(app, "authorizationRequired", "Checking Authorization: Bearer <token> in request", requestId)

    accessToken, accessTokenExists := c.Get(accessTokenKey)
    if accessTokenExists == false {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "No access token found. Hint: Is bearer token missing?"})
      c.Abort()
      return
    }

    // Sanity check: Claims
    fmt.Println(accessToken)

    foundRequiredScopes := true
    if foundRequiredScopes {
      debugLog(app, "authorizationRequired", "Valid scopes. WE DID NOT CHECK IT - TODO!", requestId)
      c.Next() // Authentication successful, continue.
      return;
    }

    // Deny by default
    debugLog(app, "authorizationRequired", "Missing required scopes: ", requestId)
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing required scopes: "})
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}

func debugLog(app string, event string, msg string, requestId string) {
  if requestId == "" {
    fmt.Println(fmt.Sprintf("[app:%s][event:%s] %s", app, event, msg))
    return;
  }
  fmt.Println(fmt.Sprintf("[app:%s][request-id:%s][event:%s] %s", app, requestId, event, msg))
}
