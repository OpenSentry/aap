package main

import (
  "fmt"
  "strings"
  "net/http"
  "net/url"

  "golang.org/x/oauth2"
  "golang.org/x/oauth2/clientcredentials"

  "github.com/gin-gonic/gin"
  "github.com/atarantini/ginrequestid"

  "golang-cp-be/config"
  "golang-cp-be/controller"
)

var (
  hydraClient *http.Client
)

func init() {
  config.InitConfigurations()
}

func main() {

  // Initialize the cp-be http client with client credentials token for used to call hydra.
  var hydraClientCredentialsConfig *clientcredentials.Config = &clientcredentials.Config{
    ClientID:     config.CpBe.ClientId,
    ClientSecret: config.CpBe.ClientSecret,
    TokenURL:     config.Hydra.TokenUrl,
    Scopes:       config.CpBe.RequiredScopes,
    EndpointParams: url.Values{"audience": {"hydra"}},
    AuthStyle: 2, // https://godoc.org/golang.org/x/oauth2#AuthStyle
  }
  hydraToken, err := requestAccessTokenForHydra(hydraClientCredentialsConfig)
  if err != nil {
    fmt.Println("Unable to aquire hydra access token. Error: " + err.Error())
    return
  }
  fmt.Println("Logging access token to hydra. Do not do this in production")
  fmt.Println(hydraToken) // FIXME Do not log this!!
  hydraClient = hydraClientCredentialsConfig.Client(oauth2.NoContext)
  // FIXME: Will this serialize all calls to the api. As all ep calls must wait for 1 client resource to get ready. ?!

  r := gin.Default()
  r.Use(ginrequestid.RequestId())
  //r.Use(logRequest())
  r.Use(requireBearerAccessToken())
  r.Use(useHydraClient(hydraClient))

  r.GET( "/authorizations", controller.GetAuthorizations)
  r.POST("/authorizations", controller.PostAuthorizations)
  r.PUT( "/authorizations", controller.PutAuthorizations)

  r.POST( "/authorizations/authorize", controller.AuthorizationsAuthorize)
  r.POST( "/authorizations/reject", controller.AuthorizationsReject)

  r.RunTLS(":80", "/srv/certs/cpbe-cert.pem", "/srv/certs/cpbe-key.pem")
  // r.Run() // listen and serve on 0.0.0.0:8080
}

func requestAccessTokenForHydra(provider *clientcredentials.Config) (*oauth2.Token, error) {
  var token *oauth2.Token
  token, err := provider.Token(oauth2.NoContext)
  if err != nil {
    return token, err
  }
  return token, nil
}

//  = Middleware

func useHydraClient(client *http.Client) gin.HandlerFunc {
  return func(c *gin.Context) {
    c.Set("hydraClient", client)
    c.Next()
  }
}

func logRequest() gin.HandlerFunc {
  return func(c *gin.Context) {
    fmt.Println("Logging all requests. Do not do this in production it will leak tokens")
    fmt.Println(c.Request)
    c.Next()
  }
}

// Look for a bearer token and unmarshal it into the gin context for the request for later use.
func requireBearerAccessToken() gin.HandlerFunc {
  return func(c *gin.Context) {
    auth := c.Request.Header.Get("Authorization")
    split := strings.SplitN(auth, " ", 2)
    if len(split) == 2 && strings.EqualFold(split[0], "bearer") {
      token := &oauth2.Token{
        AccessToken: split[1],
        TokenType: split[0],
      }

      if token.Valid() {
        c.Set("bearer_token", token)
        c.Next()
        return
      }

      // Token invalid
      c.JSON(http.StatusForbidden, gin.H{"error": "Authorization bearer token is invalid"})
      c.Abort()
      return;
    }

    // Deny by default.
    c.JSON(http.StatusForbidden, gin.H{"error": "Authorization bearer token is missing"})
    c.Abort()
  }
}
