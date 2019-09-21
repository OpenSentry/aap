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

type ErrorResponse struct {
  Code  int    `json:"code" binding:"required"`
  Error string `json:"error" binding:"required"`
}

type BulkResponse struct {
  Index  int             `json:"index"`
  Status int             `json:"status"`
  Errors []ErrorResponse `json:"errors"`
}

// /scopes

type CreateScopesRequest struct {
  Scope                     string    `json:"scope" binding:"required"`
  Title                     string    `json:"title" binding:"required"`
  Description               string    `json:"description" binding:"required"`
}

type CreateScopesResponse struct {
  Scope                     string    `json:"scope" binding:"required"`
  Title                     string    `json:"title" binding:"required"`
  Description               string    `json:"description" binding:"required"`
  CreatedBy                 string    `json:"created_by" binding:"required"`
}
type CreateScopesBulkResponse struct {
  *BulkResponse
  Ok CreateScopesResponse `json:"ok"`
}

type UpdateScopesRequest struct {
  Scope                     string    `json:"scope" binding:"required"`
  Title                     string    `json:"title" binding:"required"`
  Description               string    `json:"description" binding:"required"`
}

type UpdateScopesResponse struct {
  Scope                     string    `json:"scope" binding:"required"`
  Title                     string    `json:"title" binding:"required"`
  Description               string    `json:"description" binding:"required"`
  CreatedBy                 string    `json:"created_by" binding:"required"`
}

type ReadScopesRequest struct {
  Scope                     string    `json:"scope" binding:"required"`
}

type ReadScopesResponse struct {
  Scope                     string    `json:"scope" binding:"required"`
  Title                     string    `json:"title" binding:"required"`
  Description               string    `json:"description"`
  CreatedBy                 string    `json:"created_by" binding:"required"`
}
type ReadScopesBulkResponse struct {
  *BulkResponse
  Ok []ReadScopesResponse `json:"ok"`
}

// /scopes/grant

type CreateScopesGrantRequest struct {
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type CreateScopesGrantResponse struct {
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesGrantRequest struct {
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesGrantResponse struct {
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

// /scopes/expose

type CreateScopesExposeRequest struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type CreateScopesExposeResponse struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesExposeRequest struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesExposeResponse struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

// /scopes/consent

type CreateScopesConsentRequest struct {
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type CreateScopesConsentResponse struct {
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesConsentRequest struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesConsentResponse struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

func ReadScopes(url string, client *AapClient, requests []ReadScopesRequest) ([]ReadScopesBulkResponse, error) {
  var response []ReadScopesBulkResponse // []ReadScopesResponse

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
