package controller

import (
  "github.com/gin-gonic/gin"
  "net/http"
  "golang-cp-be/interfaces"
  "golang-cp-be/gateway/hydra"
  _ "os"
)

func PostAuthorizationsAuthorize(c *gin.Context) {

  var input interfaces.PostAuthorizationsAuthorizeRequest

  err := c.BindJSON(&input)

  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  hydraConsentResponse := hydra.GetConsent(input.Challenge)

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

  hydraConsentAcceptResponse := hydra.AcceptConsent(input.Challenge, hydraConsentAcceptRequest)

  c.JSON(http.StatusOK, gin.H{
    "authorized": true,
    "redirect_to": hydraConsentAcceptResponse.RedirectTo,
  })

  return
}
