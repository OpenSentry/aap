package client

import (
  bulky "github.com/charmixer/bulky/client"
)

type Consent struct {
  Reference  string `json:"reference_id"  validate:"required,uuid"` // OAuth2:Subject
  Subscriber string `json:"subscriber_id" validate:"required,uuid"` // OAuth2:Client
  Publisher  string `json:"publisher_id"  validate:"required,uuid"` // OAuth2:Resource Server
  Scope      string `json:"scope"         validate:"required"`      // OAuth2:Scope, published by the resource server
}

type CreateConsentsResponse Consent
type CreateConsentsRequest struct {
  Reference  string `json:"reference_id"  validate:"required,uuid"` // OAuth2:Subject
  Subscriber string `json:"subscriber_id" validate:"required,uuid"` // OAuth2:Client
  Publisher  string `json:"publisher_id"  validate:"required,uuid"` // OAuth2:Resource Server
  Scope      string `json:"scope"         validate:"required"`      // OAuth2:Scope, published by the resource server
}

type ReadConsentsResponse []Consent
type ReadConsentsRequest struct {
  Reference  string `json:"reference_id"            validate:"required,uuid"`
  Subscriber string `json:"subscriber_id,omitempty" validate:"omitempty,uuid"` // OAuth2:Client
  Publisher  string `json:"publisher_id,omitempty"  validate:"omitempty,uuid"` // OAuth2:Resource Server
  Scope      string `json:"scope,omitempty"         validate:"omitempty"`      // OAuth2:Scope, published by the resource server
}

type DeleteConsentsResponse Consent
type DeleteConsentsRequest struct {
  Reference  string `json:"reference_id"  validate:"required,uuid"` // OAuth2:Subject
  Subscriber string `json:"subscriber_id" validate:"required,uuid"` // OAuth2:Client
  Publisher  string `json:"publisher_id"  validate:"required,uuid"` // OAuth2:Resource Server
  Scope      string `json:"scope"         validate:"required"`      // OAuth2:Scope, published by the resource server
}

type Authorization struct {
  Challenge  string `json:"challenge" validate:"required"`
  Authorized bool   `json:"authorized"`
  RedirectTo string `json:"redirect_to" validate:"omitempty,uri"`

  ClientId string `json:"client_id,omitempty" validate:"omitempty,uuid"`
  ClientName string `json:"client_name,omitempty"`

  Subject string `json:"subject,omitempty" validate:"omitempty,uuid"`
  SubjectName string `json:"subject_name,omitempty"`
  SubjectEmail string `json:"subject_email,omitempty" validate:"omitempty,email"`

  RequestedScopes []string `json:"requested_scopes,omitempty"`
  GrantedScopes   []string `json:"grant_scopes,omitempty"`

  RequestedAudiences []string `json:"requested_audiences,omitempty"` // requested_access_token_audience
}

type CreateConsentsAuthorizeResponse Authorization
type CreateConsentsAuthorizeRequest struct {
  Challenge   string   `json:"challenge" validate:"required"`
  GrantScopes []string `json:"grant_scopes,omitempty"`
}

type CreateConsentsRejectResponse Authorization
type CreateConsentsRejectRequest struct {
  Challenge   string   `json:"challenge" validate:"required"`
}

func CreateConsents(client *AapClient, url string, requests []CreateConsentsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "POST", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func ReadConsents(client *AapClient, url string, requests []ReadConsentsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "GET", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func DeleteConsents(client *AapClient, url string, requests []DeleteConsentsRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "DELETE", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}

func CreateConsentsAuthorize(client *AapClient, url string, requests []CreateConsentsAuthorizeRequest) (status int, responses bulky.Responses, err error) {
  status, err = handleRequest(client, requests, "POST", url, &responses)

  if err != nil {
    return status, nil, err
  }

  return status, responses, nil
}
