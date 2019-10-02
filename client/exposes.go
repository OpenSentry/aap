package client

// /publishes

type ReadPublishesRequest struct {
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
}

type ReadPublishesResponse struct {
  ResourceServerId          string    `json:"resource_server_id" binding:"required"`
  CreatedByIdentityId       string    `json:"created_by_identity_id" binding:"required"`
  Scope                     string    `json:"scope" binding:"required"`
}

type CreatePublishesRequest struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type CreatePublishesResponse struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type DeletePublishesRequest struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}

type DeletePublishesResponse struct {
  IdentityId                string    `json:"identity_id" validate:"required"`
  Scopes                    []string  `json:"scopes" validate:"required"`
}
