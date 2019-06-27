package config

import (
  "os"
)

type HydraConfig struct {
  Url                         string
  AdminUrl                    string
  ConsentRequestUrl           string
  ConsentRequestAcceptUrl     string
  ConsentRequestRejectUrl     string
}

type OAuth2ClientConfig struct {
  ClientId        string
  ClientSecret    string
  Scopes          []string
  RedirectURL     string
  Endpoint        string
}

type CpFeConfig struct {
  CsrfAuthKey     string
  CpBackendUrl    string
}

var Hydra HydraConfig
var CpFe CpFeConfig

func InitConfigurations() {
  Hydra.Url                     = getEnvStrict("HYDRA_URL")
  Hydra.AdminUrl                = getEnvStrict("HYDRA_ADMIN_URL")
  Hydra.ConsentRequestUrl       = Hydra.AdminUrl + "/oauth2/auth/requests/consent"
  Hydra.ConsentRequestAcceptUrl = Hydra.ConsentRequestUrl + "/accept"
  Hydra.ConsentRequestRejectUrl = Hydra.ConsentRequestUrl + "/reject"

  CpFe.CsrfAuthKey              = getEnv("CSRF_AUTH_KEY") // 32 byte long auth key. When you change this user session will break.
  CpFe.CpBackendUrl             = getEnv("CP_BACKEND_URL")
}

func getEnv(name string) string {
  return os.Getenv(name)
}

func getEnvStrict(name string) string {
  r := getEnv(name)

  if r == "" {
    panic("Missing environment variable: " + name)
  }

  return r
}
