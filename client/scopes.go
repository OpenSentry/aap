package client

import (
	bulky "github.com/charmixer/bulky/client"
)

// /scopes

type Scope struct {
	Scope string `json:"scope" validate:"required"`
}

type CreateScopesResponse Scope
type CreateScopesRequest struct {
	Scope string `json:"scope" validate:"required"`
}

type UpdateScopesResponse Scope
type UpdateScopesRequest struct {
	Scope string `json:"scope" validate:"required"`
}

type ReadScopesResponse []Scope
type ReadScopesRequest struct {
	Scope string `json:"scope" validate:"required"`
}

func ReadScopes(client *AapClient, url string, requests []ReadScopesRequest) (status int, responses bulky.Responses, err error) {
	status, err = handleRequest(client, requests, "GET", url, &responses)

	if err != nil {
		return status, nil, err
	}

	return status, responses, nil
}

func CreateScopes(client *AapClient, url string, requests []CreateScopesRequest) (status int, responses bulky.Responses, err error) {
	status, err = handleRequest(client, requests, "POST", url, &responses)

	if err != nil {
		return status, nil, err
	}

	return status, responses, nil
}

func UpdateScopes(client *AapClient, url string, requests []UpdateScopesRequest) (status int, responses bulky.Responses, err error) {
	status, err = handleRequest(client, requests, "PUT", url, &responses)

	if err != nil {
		return status, nil, err
	}

	return status, responses, nil
}
