package client

import (
  bulky "github.com/charmixer/bulky/client"
)

type Entity struct {
  Reference        string    `json:"reference_id" validate:"required,uuid"`
  Creator          string    `json:"creator_id" validate:"required,uuid"`
}

type CreateEntitiesResponse Entity
type CreateEntitiesRequest struct {
  Reference        string    `json:"reference_id" validate:"required,uuid"`
  Creator          string    `json:"creator_id" validate:"required,uuid"`
}


func CreateEntities(client *AapClient, url string, requests []CreateEntitiesRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "POST", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}
