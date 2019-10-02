package client

import (
  _ "net/http"
  _ "encoding/json"
  _ "io/ioutil"
  _ "bytes"
  _ "strings"
  _ "errors"
  _ "golang.org/x/net/context"
  _ "golang.org/x/oauth2/clientcredentials"
)

// /consents

type ReadConsentsRequest struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
}

type ReadConsentsResponse struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  ClientId                  string    `json:"client_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
}

type CreateConsentsRequest struct {
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type CreateConsentsResponse struct {
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type DeleteConsentsRequest struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type DeleteConsentsResponse struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  ResourceServerId          string    `json:"resource_server_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}
