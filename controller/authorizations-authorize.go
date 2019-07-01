package controller

import (
  _ "os"
  "net/http"
  _ "encoding/json"
  "fmt"

  "github.com/gin-gonic/gin"

  "golang-cp-be/interfaces"
  "golang-cp-be/gateway/hydra"
)

type AuthorizeRequest struct {
  Challenge                   string            `json:"challenge" binding:"required"`
  GrantScopes                 []string          `json:"grant_scopes,omitempty"`
}

type AuthorizeResponse struct {
  Challenge                   string            `json:"challenge" binding:"required"`
  Authorized                  bool              `json:"authorized" binding:"required"`
  GrantScopes                 []string          `json:"grant_scopes,omitempty"`
  RequestedScopes             []string          `json:"requested_scopes,omitempty"`
  RedirectTo                  string            `json:"redirect_to,omitempty`
}

type RejectRequest struct {
  Challenge                   string            `json:"challenge" binding:"required"`
}

type RejectResponse struct {
  Authorized                  bool              `json:"authorized" binding:"required"`
  RedirectTo                  string            `json:"redirect_to" binding:"required"`
}

func authorize(client *http.Client, authorizeRequest AuthorizeRequest) (AuthorizeResponse, error) {
  var authorizeResponse AuthorizeResponse

  hydraConsentResponse, err := hydra.GetConsent(authorizeRequest.Challenge, client)
  if err != nil {
    return authorizeResponse, err
  }

  if hydraConsentResponse.Skip {
    hydraConsentAcceptRequest := interfaces.HydraConsentAcceptRequest{
      GrantScope: hydraConsentResponse.RequestedScopes, // We can grant all scopes that have been requested - hydra already checked for us that no additional scopes are requested accidentally.
      Session: interfaces.HydraConsentAcceptSession {
      },
      GrantAccessTokenAudience: hydraConsentResponse.GrantAccessTokenAudience,
      Remember: true,
      RememberFor: 3600,
    }
    hydraConsentAcceptResponse, err := hydra.AcceptConsent(authorizeRequest.Challenge, client, hydraConsentAcceptRequest)
    if err != nil {
      return authorizeResponse, err
    }

    authorizeResponse = AuthorizeResponse{
      Challenge: authorizeRequest.Challenge,
      Authorized: true,
      GrantScopes: hydraConsentResponse.RequestedScopes,
      RequestedScopes: authorizeRequest.GrantScopes,
      RedirectTo: hydraConsentAcceptResponse.RedirectTo,
    }
    return authorizeResponse, nil
  }

  // Require atleast one scope to grant or this is just a masked read.
  if len(authorizeRequest.GrantScopes) <= 0 {
    authorizeResponse = AuthorizeResponse{
      Challenge: authorizeRequest.Challenge,
      Authorized: false,
      RequestedScopes: hydraConsentResponse.RequestedScopes,
    }
    return authorizeResponse, nil
  }

  hydraConsentAcceptRequest := interfaces.HydraConsentAcceptRequest{
    GrantScope: authorizeRequest.GrantScopes,
    Session: interfaces.HydraConsentAcceptSession {
    },
    GrantAccessTokenAudience: hydraConsentResponse.GrantAccessTokenAudience,
    Remember: true,
    RememberFor: 3600,
  }
  hydraConsentAcceptResponse, err := hydra.AcceptConsent(authorizeRequest.Challenge, client, hydraConsentAcceptRequest)
  if err != nil {
    return authorizeResponse, err
  }

  authorizeResponse = AuthorizeResponse{
    Challenge: authorizeRequest.Challenge,
    Authorized: true,
    GrantScopes: authorizeRequest.GrantScopes,
    RequestedScopes: authorizeRequest.GrantScopes,
    RedirectTo: hydraConsentAcceptResponse.RedirectTo,
  }
  return authorizeResponse, nil
}

func AuthorizationsAuthorize(c *gin.Context) {
  fmt.Println(fmt.Sprintf("[request-id:%s][event:AuthorizationsAuthorize]", c.MustGet("RequestId")))

  hydraClient, _ := c.Get("hydraClient")

  var input AuthorizeRequest
  err := c.BindJSON(&input)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    c.Abort()
    return
  }

  authorizeResponse, err := authorize(hydraClient.(*http.Client), input)
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    c.Abort()
    return
  }

  fmt.Println(fmt.Sprintf("CpBe.AuthorizationsAuthorize, authorized:%s redirect_to:%s", authorizeResponse.Authorized, authorizeResponse.RedirectTo))
  c.JSON(http.StatusOK, authorizeResponse)
}

func AuthorizationsReject(c *gin.Context) {
  fmt.Println(fmt.Sprintf("[request-id:%s][event:PostAuthorizationsReject]", c.MustGet("RequestId")))

  hydraClient, _ := c.Get("hydraClient")

  var input RejectRequest
  err := c.BindJSON(&input)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    c.Abort()
    return
  }

  hydraConsentRejectRequest := interfaces.HydraConsentRejectRequest{
    Error: "",
    ErrorDebug: "",
    ErrorDescription: "",
    ErrorHint: "",
    StatusCode: 403,
  }
  hydraConsentRejectResponse, err := hydra.RejectConsent(input.Challenge, hydraClient.(*http.Client), hydraConsentRejectRequest)
  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    c.Abort()
    return
  }

  fmt.Println("CpBe.PostAuthorizationsReject, authorized:false redirect_to:" + hydraConsentRejectResponse.RedirectTo)
  c.JSON(http.StatusOK, gin.H{
    "authorized": false,
    "redirect_to": hydraConsentRejectResponse.RedirectTo,
  })
  c.Abort()
}
