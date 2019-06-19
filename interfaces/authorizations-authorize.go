package interfaces

type CpSession struct {
  AccessToken                 string            `json:"access_token"`
  IdToken                     string            `json:"id_token"`
}

type PostAuthorizationsAuthorizeRequest struct {
  GrantScopes                 []string          `json:"grant_scopes" binding:"required"`
  Challenge                   string            `json:"challenge" binding:"required"`
  Session                     struct {
    AccessToken                 string            `json:"access_token"`
    IdToken                     string            `json:"id_token"`
  } `json:"session" binding:"required"`
}

type PostAuthorizationsAuthorizeResponse struct {
  GrantScopes                 []string          `json:"grant_scopes" binding:"required"`
  RequestedScopes             []string          `json:"requested_scopes" binding:"required"`
  Authorized                  bool              `json:"authorized" binding:"required"`
  RedirectTo                  string            `json:"redirect_to" binding:"required"`
}
