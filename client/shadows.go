package client

import (
  bulky "github.com/charmixer/bulky/client"
)

type Shadow struct {
  Entity        string    `json:"entity_id" validate:"required,uuid"`
  Shadow        string    `json:"shadow_id" validate:"required,uuid"`
  NotBefore     int64     `json:"nbf" validate:"gte=0"`
  Expire        int64     `json:"exp" validate:"eq=0|gtefield=NotBefore"`
}

type CreateShadowsResponse Shadow
type CreateShadowsRequest struct {
  Entity        string    `json:"entity_id" validate:"required,uuid"`
  Shadow        string    `json:"shadow_id" validate:"required,uuid"`
  NotBefore     int64     `json:"nbf" validate:"gte=0"`
  Expire        int64     `json:"exp" validate:"eq=0|gtefield=NotBefore"`
}

type DeleteShadowsResponse Shadow
type DeleteShadowsRequest struct {
  Entity        string    `json:"entity_id" validate:"required,uuid"`
  Shadow        string    `json:"shadow_id" validate:"required,uuid"`
}

type ReadShadowsResponse []Shadow
type ReadShadowsRequest struct {
  Entity        string    `json:"entity_id" validate:"required,uuid"`
}


func CreateShadows(client *AapClient, url string, requests []CreateShadowsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "POST", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func ReadShadows(client *AapClient, url string, requests []ReadShadowsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "GET", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func DeleteShadows(client *AapClient, url string, requests []DeleteShadowsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "DELETE", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}
