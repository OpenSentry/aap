package controller

import (
  "github.com/gin-gonic/gin"
  "net/http"
  "golang-cp-be/interfaces"
  "golang-cp-be/gateway/hydra"
  _ "os"
  "fmt"
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

  fmt.Println(hydraConsentResponse)

  if hydraConsentResponse.Skip {

    hydraConsentAcceptRequest := interfaces.HydraConsentAcceptRequest{
      GrantScope: hydraConsentResponse.RequestedScopes, // We can grant all scopes that have been requested - hydra already checked for us that no additional scopes are requested accidentally.
      Session: interfaces.HydraConsentAcceptSession {
      },
      GrantAccessTokenAudience: hydraConsentResponse.GrantAccessTokenAudience,
      Remember: true,
      RememberFor: 3600,
    }

    hydraConsentAcceptResponse, err := hydra.AcceptConsent(input.Challenge, hydraConsentAcceptRequest)
    if err != nil {
      fmt.Println(err)
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      c.Abort();
      return
    }

    c.JSON(http.StatusOK, gin.H{
      "id": hydraConsentResponse.Subject,
      "authorized": true,
      "redirect_to": hydraConsentAcceptResponse.RedirectTo,
    })
    c.Abort();
    return
  }

  hydraConsentAcceptRequest := interfaces.HydraConsentAcceptRequest{
    GrantScope: input.GrantScopes,
    Session: interfaces.HydraConsentAcceptSession {
    },
    GrantAccessTokenAudience: hydraConsentResponse.GrantAccessTokenAudience,
    Remember: true,
    RememberFor: 3600,
  }
  hydraConsentAcceptResponse, err := hydra.AcceptConsent(input.Challenge, hydraConsentAcceptRequest)
  if err != nil {
    fmt.Println(err)
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    c.Abort();
    return
  }

  c.JSON(http.StatusOK, gin.H{
    "id": hydraConsentResponse.Subject,
    "authorized": false,
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
