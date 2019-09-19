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

  "github.com/charmixer/aap/config"
  "github.com/charmixer/aap/environment"

  "github.com/charmixer/aap/endpoints/authorizations"
  "github.com/charmixer/aap/endpoints/scopes"
  "github.com/charmixer/aap/endpoints/grants"
  "github.com/charmixer/aap/endpoints/exposes"
  "github.com/charmixer/aap/endpoints/consents"
  "github.com/charmixer/aap/migration"
  "github.com/charmixer/aap/utils"
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
  r := gin.New() // Clean gin to take control with logging.
  r.Use(utils.ProcessMethodOverride(r))
  r.Use(gin.Recovery())

  r.Use(utils.RequestId())
  r.Use(utils.RequestLogger(environment.LogKey, environment.RequestIdKey, log, appFields))

  // ## QTNA - Questions that need answering before granting access to a protected resource
  // 1. Is the user or client authenticated? Answered by the process of obtaining an access token.
  // 2. Is the access token expired?
  // 3. Is the access token granted the required scopes?
  // 4. Is the user or client giving the grants in the access token authorized to operate the scopes granted?
  // 5. Is the access token revoked?

  // All requests need to be authenticated.
  r.Use(utils.AuthenticationRequired(environment.LogKey, environment.AccessTokenKey))

  hydraIntrospectUrl := config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.introspect")

  aconf := utils.AuthorizationConfig{
    LogKey:             environment.LogKey,
    AccessTokenKey:     environment.AccessTokenKey,
    HydraConfig:        env.HydraConfig,
    HydraIntrospectUrl: hydraIntrospectUrl,
  }

  r.GET("/authorizations",            utils.AuthorizationRequired(aconf, "aap:authorizations:get"),    authorizations.GetAuthorizations(env))
  r.POST("/authorizations",           utils.AuthorizationRequired(aconf, "aap:authorizations:post"),   authorizations.PostAuthorizations(env))
  r.PUT("/authorizations",            utils.AuthorizationRequired(aconf, "aap:authorizations:update"), authorizations.PutAuthorizations(env))

  r.GET("/scopes",                    utils.AuthorizationRequired(aconf, "aap:read:scopes"),           scopes.GetScopes(env))
  r.POST("/scopes",                   utils.AuthorizationRequired(aconf, "aap:create:scopes"),         scopes.PostScopes(env))
  r.PUT("/scopes",                    utils.AuthorizationRequired(aconf, "aap:update:scopes"),         scopes.PutScopes(env))

  r.POST("/scopes/grant",             utils.AuthorizationRequired(aconf, "aap:create:scopes:grant"),   scopes.PostScopesGrant(env))
  r.DELETE("/scopes/grant",           utils.AuthorizationRequired(aconf, "aap:delete:scopes:grant"),   scopes.DeleteScopesGrant(env))

  r.POST("/scopes/consent",           utils.AuthorizationRequired(aconf, "aap:create:scopes:consent"), scopes.PostScopesConsent(env))
  r.DELETE("/scopes/consent",         utils.AuthorizationRequired(aconf, "aap:delete:scopes:consent"), scopes.DeleteScopesConsent(env))

  r.POST("/scopes/expose",            utils.AuthorizationRequired(aconf, "aap:create:scopes:expose"),  scopes.PostScopesExpose(env))
  r.DELETE("/scopes/expose",          utils.AuthorizationRequired(aconf, "aap:delete:scopes:expose"),  scopes.DeleteScopesExpose(env))

  r.GET("/exposes",                   utils.AuthorizationRequired(aconf, "aap:read:exposes"),          exposes.GetExposes(env))
  r.GET("/consents",                  utils.AuthorizationRequired(aconf, "aap:read:consents"),         consents.GetConsents(env))
  r.GET("/grants",                    utils.AuthorizationRequired(aconf, "aap:read:grants"),           grants.GetGrants(env))

  r.POST("/authorizations/authorize", utils.AuthorizationRequired(aconf, "aap:authorize:identity"),    authorizations.PostAuthorize(env))
  r.POST("/authorizations/reject",    utils.AuthorizationRequired(aconf, "aap:reject:identity"),       authorizations.PostReject(env))

  // r.POST("/scopes", utils.AuthorizationRequired(), Route(GetScopes(), input, output))
  // r.POST("/scopes", utils.AuthorizationRequired(), bindInput(definition), handler(), bindOutput(defintion))

  r.RunTLS(":" + config.GetString("serve.public.port"), config.GetString("serve.tls.cert.path"), config.GetString("serve.tls.key.path"))
}
