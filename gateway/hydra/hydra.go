package hydra

import (
  "golang-cp-be/config"
  "golang-cp-be/interfaces"
  "net/http"
  "bytes"
  "encoding/json"
  "io/ioutil"
)

func getDefaultHeaders() map[string][]string {
  return map[string][]string{
    "Content-Type": []string{"application/json"},
    "Accept": []string{"application/json"},
  }
}

func GetConsent(challenge string) interfaces.HydraConsentResponse {

  client := &http.Client{}

  request, _ := http.NewRequest("GET", config.Hydra.ConsentRequestUrl, nil)
  request.Header = getDefaultHeaders()

  query := request.URL.Query()
  query.Add("consent_challenge", challenge)
  request.URL.RawQuery = query.Encode()

  response, _ := client.Do(request)

  responseData, _ := ioutil.ReadAll(response.Body)

  var hydraConsentResponse interfaces.HydraConsentResponse
  json.Unmarshal(responseData, &hydraConsentResponse)

  return hydraConsentResponse
}

func AcceptConsent(challenge string, hydraConsentAcceptRequest interfaces.HydraConsentAcceptRequest) interfaces.HydraConsentAcceptResponse {
  // call hydra with accept login request

  client := &http.Client{}

  body, _ := json.Marshal(hydraConsentAcceptRequest)

  request, _ := http.NewRequest("PUT", config.Hydra.ConsentRequestAcceptUrl, bytes.NewBuffer(body))
  request.Header = getDefaultHeaders()

  query := request.URL.Query()
  query.Add("consent_challenge", challenge)
  request.URL.RawQuery = query.Encode()

  response, _ := client.Do(request)

  responseData, _ := ioutil.ReadAll(response.Body)

  var hydraConsentAcceptResponse interfaces.HydraConsentAcceptResponse
  json.Unmarshal(responseData, &hydraConsentAcceptResponse)

  return hydraConsentAcceptResponse
}
