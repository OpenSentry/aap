package hydra

import (
  "net/http"
  "bytes"
  "encoding/json"
  "io/ioutil"
  "fmt"

  "golang.org/x/net/context"
  "golang.org/x/oauth2/clientcredentials"
)

type HydraConsentResponse struct {
  Subject                      string                     `json:"subject"`
  Skip                         bool                       `json:"skip"`
  RedirectTo                   string                     `json:"redirect_to"`
  GrantAccessTokenAudience     string                     `json:"grant_access_token_audience"`
  RequestUrl                   string                     `json:"request_url"`
  RequestedAccessTokenAudience []string                   `json:"requested_access_token_audience"`
  RequestedScopes              []string                   `json:"requested_scope"`
}

type HydraConsentAcceptSession struct {
  AccessToken                  string                     `json:"access_token,omitempty"`
  IdToken                      string                     `json:"id_token,omitempty"`
}

type HydraConsentAcceptResponse struct {
  RedirectTo                   string                     `json:"redirect_to"`
}

type HydraConsentAcceptRequest struct {
  Subject                      string                     `json:"subject,omitempty"`
  GrantScope                   []string                   `json:"grant_scope"`
  Session                      HydraConsentAcceptSession  `json:"session" binding:"required"`
  GrantAccessTokenAudience     string                     `json:"grant_access_token_audience,omitempty" binding:"required"`
  Remember                     bool                       `json:"remember" binding:"required"`
  RememberFor                  int                        `json:"remember_for" binding:"required"`
}

type HydraConsentRejectRequest struct {
  Error            string `json:"error"`
  ErrorDebug       string `json:"error_debug"`
  ErrorDescription string `json:"error_description"`
  ErrorHint        string `json:"error_hint"`
  StatusCode       int    `json:"status_code"`
}

type HydraConsentRejectResponse struct {
  RedirectTo                   string                     `json:"redirect_to"`
}

type HydraIntrospectRequest struct {
  Token string `json:"token"`
  Scope string `json:"scope"`
}

type HydraIntrospectResponse struct {
  Active string `json:"active"`
  Aud string `json:"aud"`
  ClientId string `json:"client_id"`
  Exp string `json:"exp"`
  Iat string `json:"iat"`
  Iss string `json:"iss"`
  Scope string `json:"scope"`
  Sub string `json:"sub"`
  TokenType string `json:"token_type"`
}

type HydraClient struct {
  *http.Client
}

func NewHydraClient(config *clientcredentials.Config) *HydraClient {
  ctx := context.Background()
  client := config.Client(ctx)
  return &HydraClient{client}
}

func IntrospectToken(url string, client *HydraClient, introspectRequest HydraIntrospectRequest) (HydraIntrospectResponse, error) {
  var introspectResponse HydraIntrospectResponse

  body, _ := json.Marshal(introspectRequest)

  request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
  if err != nil {
    return introspectResponse, err
  }

  response, err := client.Do(request)
  if err != nil {
    return introspectResponse, err
  }

  responseData, err := ioutil.ReadAll(response.Body)
  if err != nil {
    return introspectResponse, err
  }
  json.Unmarshal(responseData, &introspectResponse)

  return introspectResponse, nil
}

// config.Hydra.ConsentRequestUrl
func GetConsent(url string, client *HydraClient, challenge string) (HydraConsentResponse, error) {
  var hydraConsentResponse HydraConsentResponse
  var err error

  request, _ := http.NewRequest("GET", url, nil)

  query := request.URL.Query()
  query.Add("consent_challenge", challenge)
  request.URL.RawQuery = query.Encode()

  response, _ := client.Do(request)

  statusCode := response.StatusCode

  if statusCode == 200 {
    responseData, _ := ioutil.ReadAll(response.Body)
    json.Unmarshal(responseData, &hydraConsentResponse)
    return hydraConsentResponse, nil
  }

  // Deny by default
  if ( statusCode == 404 ) {
    err = fmt.Errorf("Consent request not found for challenge %s", challenge)
  } else {
    err = fmt.Errorf("Consent request failed with status code %s for challenge %s", statusCode, challenge)
  }
  return hydraConsentResponse, err
}

// config.Hydra.ConsentRequestAcceptUrl
func AcceptConsent(url string, client *HydraClient, challenge string, hydraConsentAcceptRequest HydraConsentAcceptRequest) (HydraConsentAcceptResponse, error) {
  var hydraConsentAcceptResponse HydraConsentAcceptResponse

  body, _ := json.Marshal(hydraConsentAcceptRequest)

  request, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))

  query := request.URL.Query()
  query.Add("consent_challenge", challenge)
  request.URL.RawQuery = query.Encode()

  response, _ := client.Do(request)

  responseData, _ := ioutil.ReadAll(response.Body)

  json.Unmarshal(responseData, &hydraConsentAcceptResponse)

  return hydraConsentAcceptResponse, nil
}

// config.Hydra.ConsentRequestAcceptUrl
func RejectConsent(url string, client *HydraClient, challenge string, hydraConsentRejectRequest HydraConsentRejectRequest) (HydraConsentRejectResponse, error) {
  var hydraConsentRejectResponse HydraConsentRejectResponse

  body, _ := json.Marshal(hydraConsentRejectRequest)

  request, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))

  query := request.URL.Query()
  query.Add("consent_challenge", challenge)
  request.URL.RawQuery = query.Encode()

  response, _ := client.Do(request)

  responseData, _ := ioutil.ReadAll(response.Body)

  json.Unmarshal(responseData, &hydraConsentRejectResponse)

  return hydraConsentRejectResponse, nil
}
