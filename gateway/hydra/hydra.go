package hydra

import (
  "golang-cp-be/config"
  "golang-cp-be/interfaces"
  "net/http"
  "bytes"
  "encoding/json"
  "io/ioutil"
  "fmt"
)

func getDefaultHeaders() map[string][]string {
  return map[string][]string{
    "Content-Type": []string{"application/json"},
    "Accept": []string{"application/json"},
  }
}

func GetConsent(challenge string) (interfaces.HydraConsentResponse, error) {
  var hydraConsentResponse interfaces.HydraConsentResponse

  client := &http.Client{}

  request, _ := http.NewRequest("GET", config.Hydra.ConsentRequestUrl, nil)
  request.Header = getDefaultHeaders()

  query := request.URL.Query()
  query.Add("consent_challenge", challenge)
  request.URL.RawQuery = query.Encode()

  response, _ := client.Do(request)

  if response.StatusCode == 404 {
    return hydraConsentResponse, fmt.Errorf("hydra: consent request not found from challenge %s", challenge)
  }

  responseData, _ := ioutil.ReadAll(response.Body)

  fmt.Println(string(responseData))

  json.Unmarshal(responseData, &hydraConsentResponse)

  return hydraConsentResponse, nil
}

func AcceptConsent(challenge string, hydraConsentAcceptRequest interfaces.HydraConsentAcceptRequest) (interfaces.HydraConsentAcceptResponse, error) {
  var hydraConsentAcceptResponse interfaces.HydraConsentAcceptResponse

  client := &http.Client{}

  body, _ := json.Marshal(hydraConsentAcceptRequest)

  request, _ := http.NewRequest("PUT", config.Hydra.ConsentRequestAcceptUrl, bytes.NewBuffer(body))
  request.Header = getDefaultHeaders()

  query := request.URL.Query()
  query.Add("consent_challenge", challenge)
  request.URL.RawQuery = query.Encode()

  response, _ := client.Do(request)

  responseData, _ := ioutil.ReadAll(response.Body)

  json.Unmarshal(responseData, &hydraConsentAcceptResponse)

  return hydraConsentAcceptResponse, nil
}
