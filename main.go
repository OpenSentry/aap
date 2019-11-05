package main

import (
  "net/url"
  "os"
  "golang.org/x/net/context"
  "golang.org/x/oauth2/clientcredentials"
  "github.com/sirupsen/logrus"
  oidc "github.com/coreos/go-oidc"
  "github.com/gin-gonic/gin"
  "github.com/neo4j/neo4j-go-driver/neo4j"
  "github.com/pborman/getopt"

  nats "github.com/nats-io/nats.go"

  "github.com/charmixer/aap/app"
  "github.com/charmixer/aap/config"

  "github.com/charmixer/aap/endpoints/entities"
  "github.com/charmixer/aap/endpoints/scopes"
  "github.com/charmixer/aap/endpoints/grants"
  "github.com/charmixer/aap/endpoints/publishings"
  "github.com/charmixer/aap/endpoints/consents"
  "github.com/charmixer/aap/endpoints/subscriptions"
  "github.com/charmixer/aap/migration"
)

const (
  appName = "aap"

  RequestIdKey string = "RequestId"
  LogKey string = "log"

  AccessTokenKey string = "access_token"
  IdTokenKey string = "id_token"
)

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
    "appname": appName,
    "log.debug": logDebug,
    "log.format": logFormat,
  }
}

func main() {

  optMigrate := getopt.BoolLong("migrate", 0, "Run migration")
  optServe := getopt.BoolLong("serve", 0, "Serve application")
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

  hydraIntrospectUrl := config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.introspect")
  if hydraIntrospectUrl == "" {
    logrus.WithFields(appFields).Panic("Missing hydra introspect url")
    return
  }

  natsConnection, err := nats.Connect(config.GetString("nats.url"))
  if err != nil {
    log.WithFields(appFields).Panic(err.Error())
    return
  }
  defer natsConnection.Close()

  // Setup app state variables. Can be used in handler functions by doing closures see exchangeAuthorizationCodeCallback
  env := &app.Environment{
    Driver: driver, // Database
    Provider: provider,
    OAuth2Delegator: &app.EnvironmentOauth2Delegator{
      Config: hydraConfig,
      IntrospectTokenUrl: hydraIntrospectUrl,
    },
    Constants: &app.EnvironmentConstants{
      LogKey: LogKey,
      AccessTokenKey: AccessTokenKey,
      IdTokenKey: IdTokenKey,
      RequestIdKey: RequestIdKey,
    },
    Nats: natsConnection,
  }

  if *optServe {
    serve(env)
  } else {
    getopt.Usage()
    os.Exit(0)
  }

}

func migrate(driver neo4j.Driver) {
  migration.Migrate(driver)
}

func serve(env *app.Environment) {
  r := gin.New() // Clean gin to take control with logging.
  r.Use(app.ProcessMethodOverride(r))
  r.Use(gin.Recovery())

  r.Use(app.RequestId())
  r.Use(app.RequestLogger(env.Constants.LogKey, env.Constants.RequestIdKey, log, appFields))

  // ## QTNA - Questions that need answering before granting access to a protected resource
  // 1. Is the user or client authenticated? Answered by the process of obtaining an access token.
  // 2. Is the access token expired?
  // 3. Is the access token granted the required scopes?
  // 4. Is the user or client giving the grants in the access token authorized to operate the scopes granted?
  // 5. Is the access token revoked?

  // All requests need to be authenticated.
  r.Use(app.AuthenticationRequired(env.Constants.LogKey, env.Constants.AccessTokenKey))

  r.POST("/entities",                 app.AuthorizationRequired(env, "aap:create:entities"),       entities.PostEntities(env))
  r.GET( "/entities/judge",           app.AuthorizationRequired(env, "aap:read:entities:judge"),   entities.GetEntitiesJudge(env))

  r.GET("/scopes",                    app.AuthorizationRequired(env, "aap:read:scopes"),           scopes.GetScopes(env))
  r.POST("/scopes",                   app.AuthorizationRequired(env, "aap:create:scopes"),         scopes.PostScopes(env))
  r.PUT("/scopes",                    app.AuthorizationRequired(env, "aap:update:scopes"),         scopes.PutScopes(env))

  r.POST("/grants",                   app.AuthorizationRequired(env, "aap:create:grants"),         grants.PostGrants(env))
  r.GET("/grants",                    app.AuthorizationRequired(env, "aap:read:grants"),           grants.GetGrants(env))
  r.DELETE("/grants",                 app.AuthorizationRequired(env, "aap:delete:grants"),         grants.DeleteGrants(env))

  r.POST("/consents",                 app.AuthorizationRequired(env, "aap:create:consents"),           consents.PostConsents(env))
  r.GET("/consents",                  app.AuthorizationRequired(env, "aap:read:consents"),             consents.GetConsents(env))
  r.DELETE("/consents",               app.AuthorizationRequired(env, "aap:delete:consents"),           consents.DeleteConsents(env))
  r.POST("/consents/authorize",       app.AuthorizationRequired(env, "aap:create:consents:authorize"), consents.PostAuthorize(env))
  r.POST("/consents/reject",          app.AuthorizationRequired(env, "aap:create:consents:reject"),    consents.PostReject(env))

  r.POST("/publishes",                app.AuthorizationRequired(env, "aap:create:publishes"),      publishings.PostPublishes(env))
  r.GET("/publishes",                 app.AuthorizationRequired(env, "aap:read:publishes"),        publishings.GetPublishes(env))
  r.DELETE("/publishes",              app.AuthorizationRequired(env, "aap:delete:publishes"),      publishings.DeletePublishes(env))

  r.POST("/subscriptions",            app.AuthorizationRequired(env, "aap:create:subscriptions"),  subscriptions.PostSubscriptions(env))
  r.GET("/subscriptions",             app.AuthorizationRequired(env, "aap:read:subscriptions"),    subscriptions.GetSubscriptions(env))
  r.DELETE("/subscriptions",          app.AuthorizationRequired(env, "aap:delete:subscriptions"),  subscriptions.DeleteSubscriptions(env))

  r.RunTLS(":" + config.GetString("serve.public.port"), config.GetString("serve.tls.cert.path"), config.GetString("serve.tls.key.path"))
}
