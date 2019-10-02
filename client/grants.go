package client

// /grants

type Grant struct {
  IdentityId                string    `json:"identity_id" validate:"required,uuid"`
  Scope                     string    `json:"scope" validate:"required"`
  PublishedBy               string    `json:"published_by" validate:"required,uuid"`
  GrantedBy                 string    `json:"granted_by" validate:"required,uuid"`
}

type ReadGrantsRequest struct {
  IdentityId                string    `json:"identity_id,omitempty" binding:"required"`
  Scope                     string    `json:"scope,omitempty" binding:"required"`
  PublishedBy               string    `json:"published_by,omitempty" binding:"required"`
}

type ReadGrantsResponse struct {
  BulkResponse
  Ok                        []Grant   `json:"ok,omitempty" validate:"dive"`
}

type CreateGrantsRequest struct {
  IdentityId                string    `json:"identity_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
  PublishedBy               string    `json:"published_by" binding:"required"`
}

type CreateGrantsResponse struct {
  BulkResponse
  Ok                        Grant     `json:"ok,omitempty" validate:"dive"`
}

type DeleteGrantsRequest struct {
  IdentityId                string    `json:"identity_id" validate:"required,uuid"`
  Scope                     string    `json:"scope" validate:"required"`
  PublishedBy               string    `json:"published_by" binding:"required"`
}

type DeleteGrantsResponse struct {
  BulkResponse // 200 OK, nothing more fancy than that
}

func CreateGrants(url string, client *AapClient, request []CreateGrantsRequest) (status int, response []CreateGrantsResponse, err error) {
  status, err = handleRequest(client, request, "POST", url, &response)

  if err != nil {
    return status, nil, err
  }

  return status, response, nil
}

func DeleteGrants(url string, client *AapClient, request []DeleteGrantsRequest) (status int, response []DeleteGrantsResponse, err error) {
  status, err = handleRequest(client, request, "DELETE", url, &response)

  if err != nil {
    return status, nil, err
  }

  return status, response, nil
}

func ReadGrants(url string, client *AapClient, request []ReadGrantsRequest) (status int, response []ReadGrantsResponse, err error) {
  status, err = handleRequest(client, request, "GET", url, &response)

  if err != nil {
    return status, nil, err
  }

  return status, response, nil
}
