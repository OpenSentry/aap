package client

import (
  "net/http"
  "encoding/json"
  "io/ioutil"
  "bytes"
  _ "strings"
  "errors"
  _ "golang.org/x/net/context"
  _ "golang.org/x/oauth2/clientcredentials"
)

// /scopes

type Scope struct {
  Scope       string    `json:"scope" validate:"required"`
  Title       string    `json:"title" validate:"required"`
  Description string    `json:"description" validate:"required"`
  CreatedBy   string    `json:"created_by" validate:"required,uuid"`
}

type CreateScopesRequest struct {
  Scope                     string    `json:"scope" validate:"required"`
  Title                     string    `json:"title" validate:"required"`
  Description               string    `json:"description" validate:"required"`
}

type CreateScopesResponse struct {
  BulkResponse
  Ok Scope `json:"ok,omitempty" validate:"dive"`
}

type UpdateScopesRequest struct {
  Scope                     string    `json:"scope" validate:"required"`
  Title                     string    `json:"title" validate:"required"`
  Description               string    `json:"description" validate:"required"`
}

type UpdateScopesResponse struct {
  Scope       string    `json:"scope" validate:"required"`
  Title       string    `json:"title" validate:"required"`
  Description string    `json:"description" validate:"required"`
  CreatedBy   string    `json:"created_by" validate:"required"`
}

type ReadScopesRequest struct {
  Scope                     string    `json:"scope" validate:"required"`
}

type ReadScopesResponse struct {
  BulkResponse
  Ok []Scope `json:"ok,omitempty" validate:"dive"`
}

// /scopes/grant

type CreateScopesGrantRequest struct {
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type CreateScopesGrantResponse struct {
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type DeleteScopesGrantRequest struct {
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type DeleteScopesGrantResponse struct {
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

// /scopes/expose

type CreateScopesExposeRequest struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type CreateScopesExposeResponse struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type DeleteScopesExposeRequest struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type DeleteScopesExposeResponse struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

// /scopes/consent

type CreateScopesConsentRequest struct {
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type CreateScopesConsentResponse struct {
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type DeleteScopesConsentRequest struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type DeleteScopesConsentResponse struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

func ReadScopes(url string, client *AapClient, requests []ReadScopesRequest) ([]ReadScopesResponse, error) {
  var response []ReadScopesResponse

  body, err := json.Marshal(requests)
  if err != nil {
    return nil, err
  }

  responseData, err := callService(client, "GET", url, bytes.NewBuffer(body))
  if err != nil {
    return nil, err
  }

  err = json.Unmarshal(responseData, &response)
  if err != nil {
    return nil, err
  }

  return response, nil
}

func CreateScopes(url string, client *AapClient, requests []CreateScopesRequest) ([]CreateScopesResponse, error) {
  var response []CreateScopesResponse

  body, err := json.Marshal(requests)
  if err != nil {
    return nil, err
  }

  responseData, err := callService(client, "POST", url, bytes.NewBuffer(body))
  if err != nil {
    return nil, err
  }

  err = json.Unmarshal(responseData, &response)
  if err != nil {
    return nil, err
  }

  return response, nil
}

func UpdateScopes(scopesUrl string, client *AapClient, updateScopesRequest UpdateScopesRequest) (*UpdateScopesResponse, error) {

  body, err := json.Marshal(updateScopesRequest)
  if err != nil {
    return nil, err
  }

  var data = bytes.NewBuffer(body)

  request, err := http.NewRequest("POST", scopesUrl, data)
  if err != nil {
    return nil, err
  }

  request.Header.Set("X-HTTP-Method-Override", "PUT")

  response, err := client.Do(request)
  if err != nil {
     return nil, err
  }

  responseData, err := ioutil.ReadAll(response.Body)
  if err != nil {
    return nil, err
  }

  if response.StatusCode != 200 {
    return nil, errors.New("Failed to update scopes, status: " + string(response.StatusCode) + ", error="+string(responseData))
  }

  var updateScopesResponse UpdateScopesResponse
  err = json.Unmarshal(responseData, &updateScopesResponse)
  if err != nil {
    return nil, err
  }

  return &updateScopesResponse, nil
}
