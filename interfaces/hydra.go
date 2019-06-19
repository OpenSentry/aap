package interfaces

type HydraConsentRequestResponse struct {
  Skip                         bool        `json:"skip"`
  RedirectTo                   string      `json:"redirect_to"`
  Subject                      string      `json:"subject"`
  GrantAccessTokenAudience     string      `json:"grant_access_token_audience"`
}

type HydraConsentRequestAcceptResponse struct {
  RedirectTo  string      `json:"redirect_to"`
}

type HydraConsentAcceptSession struct {
  AccessToken                  string            `json:"access_token,omitempty"`
  IdToken                      string            `json:"id_token,omitempty"`
}

type HydraConsentAcceptRequest struct {
  GrantScope                   []string                   `json:"grant_scope"`
  Session                      HydraConsentAcceptSession  `json:"session" binding:"required"`
  GrantAccessTokenAudience     string            `json:"grant_access_token_audience,omitempty" binding:"required"`
  Remember                     bool              `json:"remember" binding:"required"`
  RememberFor                  int               `json:"remember_for" binding:"required"`
}
