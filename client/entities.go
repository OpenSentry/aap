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

type Verdict struct {
  Granted   bool     `json:"is_granted"`

  Identity  string   `json:"identity_id"  validate:"required,uuid"` // Subject access_token.sub
  Publisher string   `json:"publisher_id" validate:"required,uuid"` // Resource Server Audience
  Scope     string   `json:"scope"        validate:"required"`
  Owners    []string `json:"owners"       validate:"omitempty,dive,uuid"`// Resource Owners (often publisher or Subject)
}

type ReadEntitiesJudgeResponse Verdict
type ReadEntitiesJudgeRequest struct {
  Identity  string   `json:"identity_id"      validate:"required,uuid"` // Subject access_token.sub
  Publisher string   `json:"publisher_id"     validate:"required,uuid"` // Resource Server Audience
  Scope     string   `json:"scope"            validate:"required"`
  Owners    []string `json:"owners,omitempty" validate:"omitempty,dive,uuid"`// Resource Owners (often publisher or Subject)
}

func CreateEntities(client *AapClient, url string, requests []CreateEntitiesRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "POST", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func ReadEntitiesJudge(client *AapClient, url string, requests []ReadEntitiesJudgeRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "GET", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}