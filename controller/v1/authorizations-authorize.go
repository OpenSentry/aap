package controller

import (
  "github.com/gin-gonic/gin"
  "net/http"
  "golang-cp-be/interfaces"
  "golang-cp-be/gateway/hydra"
  _ "os"
  _ "fmt"
)

func PostAuthorizationsAuthorize(c *gin.Context) {
  var input interfaces.PostAuthorizationsAuthorizeRequest

  err := c.BindJSON(&input)

  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  hydraConsentResponse, err := hydra.GetConsent(input.Challenge)

  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    return
  }

  hydraConsentAcceptRequest := interfaces.HydraConsentAcceptRequest{
    GrantScope: input.GrantScopes,
    Session: interfaces.HydraConsentAcceptSession {
    },
    GrantAccessTokenAudience: hydraConsentResponse.GrantAccessTokenAudience,
    Remember: false,
    RememberFor: 3600,
  }

  if hydraConsentResponse.Skip {
    hydraConsentAcceptRequest = interfaces.HydraConsentAcceptRequest{
      Subject: hydraConsentResponse.Subject,
      GrantScope: input.GrantScopes,
      Session: interfaces.HydraConsentAcceptSession {
      },
      GrantAccessTokenAudience: hydraConsentResponse.GrantAccessTokenAudience,
      Remember: false,
      RememberFor: 3600,
    }
  }

  hydraConsentAcceptResponse, _ := hydra.AcceptConsent(input.Challenge, hydraConsentAcceptRequest)

  c.JSON(http.StatusOK, gin.H{
    "authorized": true,
    "redirect_to": hydraConsentAcceptResponse.RedirectTo,
  })

  return
}

func GetAuthorizationsAuthorize(c *gin.Context) {
  challenge := c.Query("challenge")

  if challenge == "" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "GET param 'challenge' is missing"})
    return
  }

  hydraConsentResponse, err := hydra.GetConsent(challenge)

  if err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "requested_scopes": hydraConsentResponse.RequestedScopes,
  })

  return
}
