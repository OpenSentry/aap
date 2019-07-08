package config

import (
  "os"
)

type SelfConfig struct {
  Port          string
}

type HydraConfig struct {
  Url                         string
  AdminUrl                    string
  TokenUrl                    string
  ConsentRequestUrl           string
  ConsentRequestAcceptUrl     string
  ConsentRequestRejectUrl     string
}

type CpBeConfig struct {
  ClientId string
  ClientSecret string
  RequiredScopes []string
}

var Hydra HydraConfig
var CpBe CpBeConfig
var Self SelfConfig

func InitConfigurations() {
  Self.Port                   = getEnvStrict("PORT")

  Hydra.Url                     = getEnvStrict("HYDRA_URL")
  Hydra.AdminUrl                = getEnvStrict("HYDRA_ADMIN_URL")
  Hydra.TokenUrl                = Hydra.Url + "/oauth2/token"
  Hydra.ConsentRequestUrl       = Hydra.AdminUrl + "/oauth2/auth/requests/consent"
  Hydra.ConsentRequestAcceptUrl = Hydra.ConsentRequestUrl + "/accept"
  Hydra.ConsentRequestRejectUrl = Hydra.ConsentRequestUrl + "/reject"

  CpBe.ClientId       = getEnvStrict("CP_BACKEND_OAUTH2_CLIENT_ID")
  CpBe.ClientSecret   = getEnvStrict("CP_BACKEND_OAUTH2_CLIENT_SECRET")
  CpBe.RequiredScopes = []string{"hydra"}
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
