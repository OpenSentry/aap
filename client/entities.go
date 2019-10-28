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

// Måske skal vi bare sige at judge siger ja og nej men at den smider debug med for alt hvad den har checket hvis man vil have det.
// Så hvis man spørger på mange owners så siger den kun ja hvis alle i listen er der og spørger man på ingen owners så kræver det at man er granted scopes direkte til rs

type Verdict struct {
  Granted bool
  // Hvis false så får man data på hvad der mangler.
}

/*
type Verdict struct {
  Owner string `json:"owner_id" validate:"required,uuid"`
  Granted bool `json:"granted"`
  GrantedScopes []Scope `json:"granted_scopes" validate:"omitempty,dive"`
  MissingScopes []Scope `json:"missing_scopes" validate:"omitempty,dive"`
}
*/

type ReadEntitiesJudgeResponse Verdict
type ReadEntitiesJudgeRequest struct {
  Publisher string  // Resource Server Audience
  Requestor string  // Subject (Either subject or client_id)
  Owners []string   // Resource Owners (often publisher or Subject)
  Scopes []string
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