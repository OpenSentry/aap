package environment

import (
  "fmt"
  "golang.org/x/oauth2/clientcredentials"
  oidc "github.com/coreos/go-oidc"
)

const (
  RequestIdKey string = "RequestId"
  AccessTokenKey = "access_token"
  IdTokenKey = "id_token"
)

type State struct {
  Provider *oidc.Provider
  HydraConfig *clientcredentials.Config
}

type Route struct {
  URL string
  LogId string
}

func DebugLog(app string, event string, msg string, requestId string) {
  if requestId == "" {
    fmt.Println(fmt.Sprintf("[app:%s][event:%s] %s", app, event, msg))
    return;
  }
  fmt.Println(fmt.Sprintf("[app:%s][request-id:%s][event:%s] %s", app, requestId, event, msg))
}
