package app

import (
  "time"
  "strings"
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"
  "github.com/gofrs/uuid"

  "golang.org/x/oauth2/clientcredentials"
  oidc "github.com/coreos/go-oidc"
  "github.com/neo4j/neo4j-go-driver/neo4j"

  nats "github.com/nats-io/nats.go"

  "github.com/charmixer/aap/utils"
)

type EnvironmentOauth2Delegator struct {
  Config *clientcredentials.Config
  IntrospectTokenUrl string
}

type EnvironmentConstants struct {
  RequestIdKey   string
  LogKey         string
  AccessTokenKey string
  IdTokenKey     string
}

type Environment struct {
  Provider        *oidc.Provider
  OAuth2Delegator *EnvironmentOauth2Delegator
  Driver          neo4j.Driver
  Constants       *EnvironmentConstants
  Nats            *nats.Conn
}

func ProcessMethodOverride(r *gin.Engine) gin.HandlerFunc {
  return func(c *gin.Context) {

    // Only need to check POST method
    if c.Request.Method != "POST" {
      return
    }

    method := c.Request.Header.Get("X-HTTP-Method-Override")
    method = strings.ToLower(method)
    method = strings.TrimSpace(method)

    // Require using method override
    if method == "" {
      c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or empty X-HTTP-Method-Override header"})
      c.Abort()
      return
    }


    if method == "post" {
      // if HandleContext is called you will make an infinite loop
      //c.Next()
      return
    }

    if method == "get" {
      c.Request.Method = "GET"
      r.HandleContext(c)
      c.Abort()
      return
    }

    if method == "put" {
      c.Request.Method = "PUT"
      r.HandleContext(c)
      c.Abort()
      return
    }

    if method == "delete" {
      c.Request.Method = "DELETE"
      r.HandleContext(c)
      c.Abort()
      return
    }

    c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported method"})
    c.Abort()
    return
  }
}

func RequestId() gin.HandlerFunc {
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

func RequestLogger(logKey string, requestIdKey string, log *logrus.Logger, appFields logrus.Fields) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    // Start timer
    start := time.Now()
    path := c.Request.URL.Path
    raw := c.Request.URL.RawQuery

    var requestId string = c.MustGet(requestIdKey).(string)
    requestLog := log.WithFields(appFields).WithFields(logrus.Fields{
      "request.id": requestId,
    })
    c.Set(logKey, requestLog)

    c.Next()

    // Stop timer
    stop := time.Now()
    latency := stop.Sub(start)

    ipData, err := utils.GetRequestIpData(c.Request)
    if err != nil {
      log.WithFields(appFields).WithFields(logrus.Fields{
        "func": "RequestLogger",
      }).Debug(err.Error())
    }

    forwardedForIpData, err := utils.GetForwardedForIpData(c.Request)
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