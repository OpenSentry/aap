package client

import (
  bulky "github.com/charmixer/bulky/client"
)

// /grants

type Grant struct {
  IdentityId                string    `json:"identity_id" validate:"required,uuid"`
  Scope                     string    `json:"scope" validate:"required"`
  Publisher                 string    `json:"published_by" validate:"required,uuid"`
}

type ReadGrantsRequest struct {
  IdentityId                string    `json:"identity_id,omitempty" binding:"required"`
  Scope                     string    `json:"scope,omitempty" binding:"required"`
  PublishedBy               string    `json:"published_by,omitempty" binding:"required"`
}

type ReadGrantsResponse []Grant

type CreateGrantsRequest struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
  PublishedBy               string    `json:"published_by" binding:"required"`
}

type CreateGrantsResponse Grant

type DeleteGrantsRequest struct {
  IdentityId                string    `json:"identity_id" validate:"required,uuid"`
  Scope                     string    `json:"scope" validate:"required"`
  PublishedBy               string    `json:"published_by" binding:"required"`
}

type DeleteGrantsResponse struct {}

func CreateGrants(url string, client *AapClient, request []CreateGrantsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, request, "POST", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func DeleteGrants(url string, client *AapClient, request []DeleteGrantsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, request, "DELETE", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func ReadGrants(url string, client *AapClient, request []ReadGrantsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, request, "GET", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}
