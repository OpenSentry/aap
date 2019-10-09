package client

import (
  bulky "github.com/charmixer/bulky/client"
)

// /scopes

type Scope struct {
  Scope       string    `json:"scope" validate:"required"`
}

type CreateScopesRequest struct {
  Scope                     string    `json:"scope" validate:"required"`
}

type CreateScopesResponse Scope

type UpdateScopesRequest struct {
  Scope                     string    `json:"scope" validate:"required"`
}

type UpdateScopesResponse Scope

type ReadScopesRequest struct {
  Scope                     string    `json:"scope" validate:"required"`
}

type ReadScopesResponse []Scope

func ReadScopes(url string, client *AapClient, request []ReadScopesRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, request, "GET", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func CreateScopes(url string, client *AapClient, request []CreateScopesRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, request, "POST", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func UpdateScopes(url string, client *AapClient, request []UpdateScopesRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, request, "PUT", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}
