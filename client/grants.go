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

// /grants

type ReadGrantsRequest struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
}

type ReadGrantsResponse struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
}
