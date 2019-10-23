package client

import (
  bulky "github.com/charmixer/bulky/client"
)

// /grants

type Grant struct {
  Identity                  string    `json:"identity_id" validate:"required,uuid"`
  Scope                     string    `json:"scope" validate:"required"`
  Publisher                 string    `json:"publisher_id" validate:"required,uuid"`
  OnBehalfOf                string    `json:"on_behalf_of_id" validate:"required,uuid"`
}


type ReadGrantsResponse []Grant
type ReadGrantsRequest struct {
  Identity                  string    `json:"identity_id,omitempty" validate:"omitempty,uuid"`
  Scope                     string    `json:"scope,omitempty" validate:"omitempty"`
  Publisher                 string    `json:"publisher_id,omitempty" validate:"omitempty,uuid"`
  OnBehalfOf                string    `json:"on_behalf_of_id,omitempty" validate:"omitempty,uuid"`
}


type CreateGrantsResponse Grant
type CreateGrantsRequest struct {
  Identity                  string    `json:"identity_id" validate:"required,uuid"`
  Scope                     string    `json:"scope" validate:"required"`
  Publisher                 string    `json:"publisher_id" validate:"required,uuid"`
  OnBehalfOf                string    `json:"on_behalf_of_id" validate:"required,uuid"`
}


type DeleteGrantsResponse struct {}
type DeleteGrantsRequest struct {
  Identity                  string    `json:"identity_id" validate:"required,uuid"`
  Scope                     string    `json:"scope" validate:"required"`
  Publisher                 string    `json:"publisher_id" validate:"required,uuid"`
  OnBehalfOf                string    `json:"on_behalf_of_id" validate:"required,uuid"`
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
