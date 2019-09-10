package main

import (
  "strings"
  "net/http"
  "net/url"
  "os"
  "time"
  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/clientcredentials"
  "github.com/sirupsen/logrus"
  oidc "github.com/coreos/go-oidc"
  "github.com/gin-gonic/gin"
  "github.com/neo4j/neo4j-go-driver/neo4j"
  "github.com/pborman/getopt"
  "github.com/gofrs/uuid"

  "github.com/charmixer/aap/config"
  "github.com/charmixer/aap/environment"
  "github.com/charmixer/aap/authorizations"
  "github.com/charmixer/aap/access"
  "github.com/charmixer/aap/migration"
)

const app = "aap"

var (
  logDebug int // Set to 1 to enable debug
  logFormat string // Current only supports default and json

  log *logrus.Logger

  appFields logrus.Fields
)

func init() {
  log = logrus.New();

  err := config.InitConfigurations()
  if err != nil {
    log.Panic(err.Error())
    return
  }

  logDebug = config.GetInt("log.debug")
  logFormat = config.GetString("log.format")

  // We only have 2 log levels. Things developers care about (debug) and things the user of the app cares about (info)
  log = logrus.New();
  if logDebug == 1 {
    log.SetLevel(logrus.DebugLevel)
  } else {
    log.SetLevel(logrus.InfoLevel)
  }
  if logFormat == "json" {
    log.SetFormatter(&logrus.JSONFormatter{})
  }

  appFields = logrus.Fields{
    "appname": app,
    "log.debug": logDebug,
    "log.format": logFormat,
  }
}

func main() {

  optMigrate := getopt.BoolLong("migrate", 0, "Run migration")
  //optServe := getopt.BoolLong("serve", 0, "Serve application")
  optHelp := getopt.BoolLong("help", 0, "Help")
  getopt.Parse()

  if *optHelp {
    getopt.Usage()
    os.Exit(0)
  }

  // https://medium.com/neo4j/neo4j-go-driver-is-out-fbb4ba5b3a30
  // Each driver instance is thread-safe and holds a pool of connections that can be re-used over time. If you don’t have a good reason to do otherwise, a typical application should have a single driver instance throughout its lifetime.
  log.WithFields(appFields).Debug("Fixme Neo4j loggning should go trough logrus so it does not differ in output from rest of the app")
  driver, err := neo4j.NewDriver(config.GetString("neo4j.uri"), neo4j.BasicAuth(config.GetString("neo4j.username"), config.GetString("neo4j.password"), ""), func(config *neo4j.Config) {
    /*if logDebug == 1 {
      config.Log = neo4j.ConsoleLogger(neo4j.DEBUG)
    } else {
      config.Log = neo4j.ConsoleLogger(neo4j.INFO)
    }*/
  });
  if err != nil {
    logrus.WithFields(appFields).Panic("neo4j.NewDriver" + err.Error())
    return
  }
  defer driver.Close()

  // migrate then exit application
  if *optMigrate {
    migrate(driver)
    os.Exit(0)
    return
  }

  provider, err := oidc.NewProvider(context.Background(), config.GetString("hydra.public.url") + "/")
  if err != nil {
    logrus.WithFields(appFields).Panic("oidc.NewProvider" + err.Error())
    return
  }

  // Setup the hydra client aap is going to use (oauth2 client credentials flow)
  hydraConfig := &clientcredentials.Config{
    ClientID:     config.GetString("oauth2.client.id"),
    ClientSecret: config.GetString("oauth2.client.secret"),
    TokenURL:     provider.Endpoint().TokenURL,
    Scopes:       config.GetStringSlice("oauth2.scopes.required"),
    EndpointParams: url.Values{"audience": {"hydra"}},
    AuthStyle: 2, // https://godoc.org/golang.org/x/oauth2#AuthStyle
  }

  // Setup app state variables. Can be used in handler functions by doing closures see exchangeAuthorizationCodeCallback
  env := &environment.State{
    Provider: provider,
    HydraConfig: hydraConfig,
    Driver: driver,
  }

  //if *optServe {
    serve(env)
  /*} else {
    getopt.Usage()
    os.Exit(0)
  }*/

}

func migrate(driver neo4j.Driver) {
  migration.Migrate(driver)
}

func serve(env *environment.State) {
  // Setup routes to use, this defines log for debug log
  routes := map[string]environment.Route{
    "/authorizations":           environment.Route{URL: "/authorizations",           LogId: "aap://authorizations"},
    "/authorizations/authorize": environment.Route{URL: "/authorizations/authorize", LogId: "aap://authorizations/authorize"},
    "/authorizations/reject":    environment.Route{URL: "/authorizations/reject",    LogId: "aap://authorizations/reject"},
    "/access":                   environment.Route{URL: "/access",                   LogId: "aap://access"},
    "/access/grant":             environment.Route{URL: "/access/grant",             LogId: "aap://access/grant"},
    "/access/revoke":            environment.Route{URL: "/access/revoke",            LogId: "aap://access/revoke"},
  }

  r := gin.New() // Clean gin to take control with logging.
  r.Use(gin.Recovery())

  r.Use(requestId())
  r.Use(RequestLogger(env))

  // ## QTNA - Questions that need answering before granting access to a protected resource
  // 1. Is the user or client authenticated? Answered by the process of obtaining an access token.
  // 2. Is the access token expired?
  // 3. Is the access token granted the required scopes?
  // 4. Is the user or client giving the grants in the access token authorized to operate the scopes granted?
  // 5. Is the access token revoked?

  // All requests need to be authenticated.
  r.Use(authenticationRequired())

  r.GET(routes["/authorizations"].URL, authorizationRequired(routes["/authorizations"], "aap.authorizations.get"), authorizations.GetCollection(env, routes["/authorizations"]))
  r.POST(routes["/authorizations"].URL, authorizationRequired(routes["/authorizations"], "aap.authorizations.post"), authorizations.PostCollection(env, routes["/authorizations"]))
  r.PUT(routes["/authorizations"].URL, authorizationRequired(routes["/authorizations"], "aap.authorizations.update"), authorizations.PutCollection(env, routes["/authorizations"]))

  r.GET(routes["/access"].URL, authorizationRequired(routes["/access"], "aap:read:access"), access.GetCollection(env, routes["/access"]))
  r.POST(routes["/access"].URL, authorizationRequired(routes["/access"], "aap:create:access"), access.PostCollection(env, routes["/access"]))
  r.PUT(routes["/access"].URL, authorizationRequired(routes["/access"], "aap:update:access"), access.PutCollection(env, routes["/access"]))

  r.PUT(routes["/access/grant"].URL, authorizationRequired(routes["/access/grant"], "aap:update:access:grant"), access.PutGrant(env, routes["/access/grant"]))

  r.POST(routes["/authorizations/authorize"].URL, authorizationRequired(routes["/authorizations/authorize"], "authorize:identity"), authorizations.PostAuthorize(env, routes["/authorizations/authorize"]))
  r.POST(routes["/authorizations/reject"].URL, authorizationRequired(routes["/authorizations/reject"], "aap.reject"), authorizations.PostReject(env, routes["/authorizations/reject"]))

  r.RunTLS(":" + config.GetString("serve.public.port"), config.GetString("serve.tls.cert.path"), config.GetString("serve.tls.key.path"))
}

func RequestLogger(env *environment.State) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    // Start timer
    start := time.Now()
    path := c.Request.URL.Path
    raw := c.Request.URL.RawQuery

    var requestId string = c.MustGet(environment.RequestIdKey).(string)
    requestLog := log.WithFields(appFields).WithFields(logrus.Fields{
      "request.id": requestId,
    })
    c.Set(environment.LogKey, requestLog)

    c.Next() // Give control to the controllers

    // Stop timer
    stop := time.Now()
    latency := stop.Sub(start)

    ipData, err := getRequestIpData(c.Request)
    if err != nil {
      log.WithFields(appFields).WithFields(logrus.Fields{
        "func": "RequestLogger",
      }).Debug(err.Error())
    }

    forwardedForIpData, err := getForwardedForIpData(c.Request)
    if err != nil {
      log.WithFields(appFields).WithFields(logrus.Fields{
        "func": "RequestLogger",
      }).Debug(err.Error())
    }

    method := c.Request.Method
    statusCode := c.Writer.Status()
    errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

    bodySize := c.Writer.Size()

    var fullpath string = path
    if raw != "" {
      fullpath = path + "?" + raw
    }

    log.WithFields(appFields).WithFields(logrus.Fields{
      "latency": latency,
      "forwarded_for.ip": forwardedForIpData.Ip,
      "forwarded_for.port": forwardedForIpData.Port,
      "ip": ipData.Ip,
      "port": ipData.Port,
      "method": method,
      "status": statusCode,
      "error": errorMessage,
      "body_size": bodySize,
      "path": fullpath,
      "request.id": requestId,
    }).Info("")
  }
  return gin.HandlerFunc(fn)
}

func authenticationRequired() gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "authenticationRequired",
    })

    log = log.WithFields(logrus.Fields{"authorization": "bearer"})
    log.Debug("Looking for access token")
    var token *oauth2.Token
    auth := c.Request.Header.Get("Authorization")
    split := strings.SplitN(auth, " ", 2)
    if len(split) == 2 || strings.EqualFold(split[0], "bearer") {

      log.Debug("Found access token")

      token = &oauth2.Token{
        AccessToken: split[1],
        TokenType: split[0],
      }

      // See #2 of QTNA
      // https://godoc.org/golang.org/x/oauth2#Token.Valid
      if token.Valid() == true {
        log.Debug("Valid access token")

        // See #5 of QTNA
        log.WithFields(logrus.Fields{"fixme": 1, "qtna": 5}).Debug("Missing check against token-revoked-list to check if token is revoked")

        c.Set(environment.AccessTokenKey, token)
        c.Next() // Authentication successful, continue.
        return;
      }

      // Deny by default
      log.Debug("Invalid access token")
      c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token."})
      c.Abort()
      return
    }

    // Deny by default
    log.Debug("Missing access token")
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization: Bearer <token> not found in request."})
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}

func authorizationRequired(route environment.Route, requiredScopes ...string) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{"func": "authorizationRequired"})

    // This is required to be here but should be garantueed by the authenticationRequired function.
    _ /*accessToken*/, accessTokenExists := c.Get(environment.AccessTokenKey)
    if accessTokenExists == false {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "No access token found. Hint: Is bearer token missing?"})
      c.Abort()
      return
    }

    strRequiredScopes := strings.Join(requiredScopes, ",")
    log.WithFields(logrus.Fields{"scopes": strRequiredScopes}).Debug("Checking required scopes");

    // See #3 of QTNA
    log.WithFields(logrus.Fields{"fixme": 1, "qtna": 3}).Debug("Missing check if access token is granted the required scopes")

    // See #4 of QTNA
    log.WithFields(logrus.Fields{"fixme": 1, "qtna": 4}).Debug("Missing check if the user or client giving the grants in the access token authorized to use the scopes granted")

    foundRequiredScopes := true
    if foundRequiredScopes {
      log.WithFields(logrus.Fields{"scopes": strRequiredScopes}).Debug("Found required scopes")
      c.Next() // Authentication successful, continue.
      return;
    }

    // Deny by default
    log.WithFields(logrus.Fields{"fixme": 1}).Debug("Calculate missing scopes and only log those");
    log.WithFields(logrus.Fields{"scopes": strRequiredScopes}).Debug("Missing required scopes. Hint: Some required scopes are missing, invalid or not granted")
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing required scopes. Hint: Some required scopes are missing, invalid or not granted"})
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}

func requestId() gin.HandlerFunc {
  return func(c *gin.Context) {
    // Check for incoming header, use it if exists
    requestID := c.Request.Header.Get("X-Request-Id")

    // Create request id with UUID4
    if requestID == "" {
      uuid4, _ := uuid.NewV4()
      requestID = uuid4.String()
    }

    // Expose it for use in the application
    c.Set("RequestId", requestID)

    // Set X-Request-Id header
    c.Writer.Header().Set("X-Request-Id", requestID)
    c.Next()
  }
}
