package interfaces

type HydraConsentResponse struct {
  Subject                      string                     `json:"subject"`
  Skip                         bool                       `json:"skip"`
  RedirectTo                   string                     `json:"redirect_to"`
  GrantAccessTokenAudience     string                     `json:"grant_access_token_audience"`
  RequestUrl                   string                     `json:"request_url"`
  RequestedAccessTokenAudience []string                   `json:"requested_access_token_audience"`
  RequestedScopes              []string                   `json:"requested_scope"`
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

type HydraConsentRejectRequest struct {
  Error            string `json:"error"`
  ErrorDebug       string `json:"error_debug"`
  ErrorDescription string `json:"error_description"`
  ErrorHint        string `json:"error_hint"`
  StatusCode       int    `json:"status_code"`
}

type HydraConsentRejectResponse struct {
  RedirectTo                   string                     `json:"redirect_to"`
}
