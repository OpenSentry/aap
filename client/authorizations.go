package client

import (
  "encoding/json"
  "bytes"
  _ "golang.org/x/net/context"
  _ "golang.org/x/oauth2/clientcredentials"
)

// authorizations

type ConsentRequest struct {
  Subject              string `form:"sub" json:"sub" binding:"required"`
  ClientId             string `form:"client_id,omitempty" json:"client_id,omitempty" binding:"required"`
  GrantedScopes      []string `form:"granted_scopes,omitempty" json:"granted_scopes,omitempty"`
  RevokedScopes      []string `form:"revoked_scopes,omitempty" json:"revoked_scopes,omitempty"`
  RequestedScopes    []string `form:"requested_scopes,omitempty" json:"requested_scopes,omitempty"`
  RequestedAudiences []string `form:"requested_audiences,omitempty" json:"requested_audiences,omitempty"` // hydra.requested_access_token_audience
}

type ConsentResponse struct {

}

// authorizations/authorize

type AuthorizeRequest struct {
  Challenge                   string            `json:"challenge" binding:"required"`
  GrantScopes                 []string          `json:"grant_scopes,omitempty"`
}

type AuthorizeResponse struct {
  Challenge                   string            `json:"challenge" binding:"required"`
  Authorized                  bool              `json:"authorized" binding:"required"`
  GrantScopes                 []string          `json:"grant_scopes,omitempty"`
  RequestedScopes             []string          `json:"requested_scopes,omitempty"`
  RedirectTo                  string            `json:"redirect_to,omitempty"`
  Subject                     string            `json:"subject,omitempty"`
  ClientId                    string            `json:"client_id,omitempty"`
  RequestedAudiences          []string          `json:"requested_audiences,omitempty"` // requested_access_token_audience
}

type RejectRequest struct {
  Challenge                   string            `json:"challenge" binding:"required"`
}

type RejectResponse struct {
  Authorized                  bool              `json:"authorized" binding:"required"`
  RedirectTo                  string            `json:"redirect_to" binding:"required"`
}

func CreateConsents(url string, client *AapClient, request ConsentRequest) (status int, response []string, err error) {
  body, err := json.Marshal(request)
  if err != nil {
    return 999, nil, err
  }

  status, responseData, err := callService(client, "POST", url, bytes.NewBuffer(body))
  if err != nil {
    return status, nil, err
  }

  err = json.Unmarshal(responseData, &response)
  if err != nil {
    return 999, nil, err
  }

  return status, response, nil
}

func FetchConsents(url string, client *AapClient, request ConsentRequest) (status int, response []string, err error) {
  body, err := json.Marshal(request)
  if err != nil {
    return 999, nil, err
  }

  status, responseData, err := callService(client, "GET", url, bytes.NewBuffer(body))
  if err != nil {
    return status, nil, err
  }

  err = json.Unmarshal(responseData, &response)
  if err != nil {
    return 666, nil, err
  }

  return status, response, nil
}

func Authorize(url string, client *AapClient, request AuthorizeRequest) (status int, response *AuthorizeResponse, err error) {
  body, err := json.Marshal(request)
  if err != nil {
    return 999, nil, err
  }

  status, responseData, err := callService(client, "POST", url, bytes.NewBuffer(body))
  if err != nil {
    return status, nil, err
  }

  err = json.Unmarshal(responseData, &response)
  if err != nil {
    return 666, nil, err
  }

  return status, response, nil
}

func Reject(url string, client *AapClient, request RejectRequest) (status int, response *RejectResponse, err error) {
  body, err := json.Marshal(request)
  if err != nil {
    return 999, nil, err
  }

  status, responseData, err := callService(client, "POST", url, bytes.NewBuffer(body))
  if err != nil {
    return status, nil, err
  }

  err = json.Unmarshal(responseData, &response)
  if err != nil {
    return 666, nil, err
  }

  return status, response, nil
}
