package client

import (
  bulky "github.com/charmixer/bulky/client"
)

// /scopes

type Publish struct {
  Publisher         string    `json:"publisher_id" validate:"required,uuid"`
  Scope             string    `json:"scope" validate:"required"`
  MayGrantScopes    []string  `json:"may_grant_scopes" validate:"omitempty"`
  Title             string    `json:"title"`
  Description       string    `json:"description"`
}

type CreatePublishesResponse Publish
type CreatePublishesRequest struct {
  Publisher   string `json:"publisher_id" validate:"required,uuid"`
  Scope       string `json:"scope" validate:"required"`
  Title       string `json:"title" validate:"required"`
  Description string `json:"description" validate:"required"`
}


type UpdatePublishesResponse Publish
type UpdatePublishesRequest struct {
  Publisher         string    `json:"publisher_id,omitempty" validate:"omitempty,uuid"`
}


type ReadPublishesResponse []Publish
type ReadPublishesRequest struct {
  Publisher         string    `json:"publisher_id,omitempty" validate:"omitempty,uuid"`
}


func ReadPublishes(client *AapClient, url string, requests []ReadPublishesRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "GET", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func CreatePublishes(client *AapClient, url string, requests []CreatePublishesRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "POST", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func UpdatePublishes(client *AapClient, url string, requests []UpdatePublishesRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "PUT", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}
