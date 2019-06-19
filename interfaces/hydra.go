package interfaces

type HydraConsentResponse struct {
  Skip                         bool                       `json:"skip"`
  RedirectTo                   string                     `json:"redirect_to"`
  Subject                      string                     `json:"subject"`
  GrantAccessTokenAudience     string                     `json:"grant_access_token_audience"`
}

type HydraConsentAcceptSession struct {
  AccessToken                  string                     `json:"access_token,omitempty"`
  IdToken                      string                     `json:"id_token,omitempty"`
}

type HydraConsentAcceptResponse struct {
  RedirectTo                   string                     `json:"redirect_to"`
}

type HydraConsentAcceptRequest struct {
  Subject                      string                     `json:"subject,omitempty"`
  GrantScope                   []string                   `json:"grant_scope"`
  Session                      HydraConsentAcceptSession  `json:"session" binding:"required"`
  GrantAccessTokenAudience     string                     `json:"grant_access_token_audience,omitempty" binding:"required"`
  Remember                     bool                       `json:"remember" binding:"required"`
  RememberFor                  int                        `json:"remember_for" binding:"required"`
}
