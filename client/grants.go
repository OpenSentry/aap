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

type ReadGrantsResponse []Grant
type ReadGrantsRequest struct {
  IdentityId                string    `json:"identity_id,omitempty" binding:"required"`
  Scope                     string    `json:"scope,omitempty" binding:"required"`
  PublishedBy               string    `json:"published_by,omitempty" binding:"required"`
}


type CreateGrantsResponse Grant
type CreateGrantsRequest struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
  PublishedBy               string    `json:"published_by" binding:"required"`
}


type DeleteGrantsResponse struct {}
type DeleteGrantsRequest struct {
  IdentityId                string    `json:"identity_id" validate:"required,uuid"`
  Scope                     string    `json:"scope" validate:"required"`
  PublishedBy               string    `json:"published_by" binding:"required"`
}


func CreateGrants(client *AapClient, url string, requests []CreateGrantsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "POST", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func DeleteGrants(client *AapClient, url string, requests []DeleteGrantsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "DELETE", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func ReadGrants(client *AapClient, url string, requests []ReadGrantsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "GET", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}
