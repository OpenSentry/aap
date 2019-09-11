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

// /exposes

type ReadExposesRequest struct {
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
}

type ReadExposesResponse struct {
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
}
