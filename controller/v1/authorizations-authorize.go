package controller

import (
  "github.com/gin-gonic/gin"
  "net/http"
  "golang-cp-be/interfaces"
  _ "os"
  "fmt"
  "io/ioutil"
  "encoding/json"
  "bytes"
)

func PostAuthorizationsAuthorize(c *gin.Context) {

  var input interfaces.PostAuthorizationsAuthorizeRequest

  err := c.BindJSON(&input)

  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }

  client := &http.Client{}

  headers := map[string][]string{
    "Content-Type": []string{"application/json"},
    "Accept": []string{"application/json"},
  }

  req, err := http.NewRequest("GET", "http://hydra:4445/oauth2/auth/requests/consent", nil)
  req.Header = headers

  q := req.URL.Query()
  q.Add("consent_challenge", input.Challenge)
  req.URL.RawQuery = q.Encode()

  response, err := client.Do(req)

  responseData, err := ioutil.ReadAll(response.Body)

  var hydraConsentRequestResponse interfaces.HydraConsentRequestResponse
  json.Unmarshal(responseData, &hydraConsentRequestResponse)

  if hydraConsentRequestResponse.Skip {
    body, _ := json.Marshal(map[string]string{
      "subject": hydraConsentRequestResponse.Subject,
    })

    req1, _ := http.NewRequest("PUT", "http://hydra:4445/oauth2/auth/requests/consent/accept", bytes.NewBuffer(body))
    req1.Header = headers

    q1 := req1.URL.Query()
    q1.Add("consent_challenge", input.Challenge)
    req1.URL.RawQuery = q1.Encode()

    response1, _ := client.Do(req1)

    responseData1, _ := ioutil.ReadAll(response1.Body)

    var hydraConsentRequestAcceptResponse interfaces.HydraConsentRequestAcceptResponse
    json.Unmarshal(responseData1, &hydraConsentRequestAcceptResponse)

    c.JSON(http.StatusOK, gin.H{
      "authorized": true,
      "redirect_to": hydraConsentRequestAcceptResponse.RedirectTo,
    })

    return
  }

  fmt.Println(hydraConsentRequestResponse.GrantAccessTokenAudience)

  hydraConsentAcceptRequest := &interfaces.HydraConsentAcceptRequest{
    GrantScope: input.GrantScopes,
    Session: interfaces.HydraConsentAcceptSession {
    },
    GrantAccessTokenAudience: hydraConsentRequestResponse.GrantAccessTokenAudience,
    Remember: false,
    RememberFor: 3600,
  }

  // call hydra with accept login request
  body, _ := json.Marshal(hydraConsentAcceptRequest)

  fmt.Println(string(body))

  req2, _ := http.NewRequest("PUT", "http://hydra:4445/oauth2/auth/requests/consent/accept", bytes.NewBuffer(body))
  req2.Header = headers

  q2 := req2.URL.Query()
  q2.Add("consent_challenge", input.Challenge)
  req2.URL.RawQuery = q2.Encode()

  response2, _ := client.Do(req2)

  responseData2, _ := ioutil.ReadAll(response2.Body)

  var hydraConsentRequestAcceptResponse interfaces.HydraConsentRequestAcceptResponse
  json.Unmarshal(responseData2, &hydraConsentRequestAcceptResponse)

  c.JSON(http.StatusOK, gin.H{
    "authorized": true,
    "redirect_to": hydraConsentRequestAcceptResponse.RedirectTo,
  })

  return
}
