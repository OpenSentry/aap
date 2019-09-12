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

type CreateScopesRequest struct {
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
  Title                     string    `json:"title" binding:"required"`
  Description               string    `json:"description" binding:"required"`
}

type CreateScopesResponse struct {
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
  Title                     string    `json:"title" binding:"required"`
  Description               string    `json:"description" binding:"required"`
}

type UpdateScopesRequest struct {
  UpdatedByIdentityId       string    `json:"updated_by_identity_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
  Title                     string    `json:"title" binding:"required"`
  Description               string    `json:"description" binding:"required"`
}

type UpdateScopesResponse struct {
  UpdatedByIdentityId       string    `json:"updated_by_identity_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
  Title                     string    `json:"title" binding:"required"`
  Description               string    `json:"description" binding:"required"`
}

type ReadScopesRequest struct {
  CreatedByIdentityId       string    `json:"created_by_identity_id"`
  Scope                     string    `json:"scope"`
}

type ReadScopesResponse struct {
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
  Title                     string    `json:"title" binding:"required"`
  Description               string    `json:"description"`
}

// /scopes/grant

type CreateScopesGrantRequest struct {
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type CreateScopesGrantResponse struct {
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesGrantRequest struct {
  DeletedByIdentityId       string    `json:"deleted_by_identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesGrantResponse struct {
  DeletedByIdentityId       string    `json:"deleted_by_identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

// /scopes/expose

type CreateScopesExposeRequest struct {
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type CreateScopesExposeResponse struct {
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesExposeRequest struct {
  DeletedByIdentityId       string    `json:"deleted_by_identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesExposeResponse struct {
  DeletedByIdentityId       string    `json:"deleted_by_identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

// /scopes/consent

type CreateScopesConsentRequest struct {
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type CreateScopesConsentResponse struct {
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesConsentRequest struct {
  DeletedByIdentityId       string    `json:"deleted_by_identity_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}

type DeleteScopesConsentResponse struct {
  DeletedByIdentityId       string    `json:"deleted_by_identity_id" binding:"required"`
  IdentityId                string    `json:"identity_id" binding:"required"`
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  Scopes                    []string  `json:"scopes" binding:"required"`
}


func CreateScopes(scopesUrl string, client *AapClient, createScopesRequest CreateScopesRequest) (*CreateScopesResponse, error) {

  body, err := json.Marshal(createScopesRequest)
  if err != nil {
    return nil, err
  }

  var data = bytes.NewBuffer(body)

  request, err := http.NewRequest("POST", scopesUrl, data)
  if err != nil {
    return nil, err
  }

  response, err := client.Do(request)
  if err != nil {
     return nil, err
  }

  responseData, err := ioutil.ReadAll(response.Body)
  if err != nil {
    return nil, err
  }

  if response.StatusCode != 200 {
    return nil, errors.New("Failed to create scopes, status: " + string(response.StatusCode) + ", error="+string(responseData))
  }

  var createdScopesResponse CreateScopesResponse
  err = json.Unmarshal(responseData, &createdScopesResponse)
  if err != nil {
    return nil, err
  }

  return &createdScopesResponse, nil
}

func UpdateScopes(scopesUrl string, client *AapClient, updateScopesRequest UpdateScopesRequest) (*UpdateScopesResponse, error) {

  body, err := json.Marshal(updateScopesRequest)
  if err != nil {
    return nil, err
  }

  var data = bytes.NewBuffer(body)

  request, err := http.NewRequest("PUT", scopesUrl, data)
  if err != nil {
    return nil, err
  }

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
