package client

import (
  bulky "github.com/charmixer/bulky/client"
)

// /scopes

type Scope struct {
  Scope       string    `json:"scope" validate:"required"`
  Title       string    `json:"title" validate:"required"`
  Description string    `json:"description" validate:""`
  CreatedBy   string    `json:"created_by" validate:"required,uuid"`
}

type CreateScopesRequest struct {
  Scope                     string    `json:"scope" validate:"required"`
  Title                     string    `json:"title" validate:"required"`
  Description               string    `json:"description" validate:""`
}

type CreateScopesResponse []struct {
  Scope `json:"ok" validate:"dive"`
}

type UpdateScopesRequest struct {
  Scope                     string    `json:"scope" validate:"required"`
  Title                     string    `json:"title" validate:"required"`
  Description               string    `json:"description" validate:"required"`
}

type UpdateScopesResponse struct {
  Scope       string    `json:"scope" validate:"required"`
  Title       string    `json:"title" validate:"required"`
  Description string    `json:"description" validate:"required"`
  CreatedBy   string    `json:"created_by" validate:"required"`
}

type ReadScopesRequest struct {
  Scope                     string    `json:"scope" validate:"required"`
}

type ReadScopesResponse []Scope

func ReadScopes(url string, client *AapClient, request []ReadScopesRequest) (status int, response []bulky.Response, err error) {
  status, err = handleRequest(client, request, "GET", url, &response)

  if err != nil {
    return status, nil, err
  }

  return status, response, nil
}

func CreateScopes(url string, client *AapClient, request []CreateScopesRequest) (status int, response []CreateScopesResponse, err error) {
  status, err = handleRequest(client, request, "POST", url, &response)

  if err != nil {
    return status, nil, err
  }

  return status, response, nil
}

func UpdateScopes(url string, client *AapClient, request []UpdateScopesRequest) (status int, response []UpdateScopesResponse, err error) {
  status, err = handleRequest(client, request, "PUT", url, &response)

  if err != nil {
    return status, nil, err
  }

  return status, response, nil
}
