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

func GetConsent(challenge string, client *http.Client) (interfaces.HydraConsentResponse, error) {
  var hydraConsentResponse interfaces.HydraConsentResponse
  var err error

  request, _ := http.NewRequest("GET", config.Hydra.ConsentRequestUrl, nil)

  query := request.URL.Query()
  query.Add("consent_challenge", challenge)
  request.URL.RawQuery = query.Encode()

  response, _ := client.Do(request)

  statusCode := response.StatusCode

  if statusCode == 200 {
    responseData, _ := ioutil.ReadAll(response.Body)
    json.Unmarshal(responseData, &hydraConsentResponse)
    return hydraConsentResponse, nil
  }

  // Deny by default
  if ( statusCode == 404 ) {
    err = fmt.Errorf("Consent request not found for challenge %s", challenge)
  } else {
    err = fmt.Errorf("Consent request failed with status code %s for challenge %s", statusCode, challenge)
  }
  return hydraConsentResponse, err
}

func AcceptConsent(challenge string, client *http.Client, hydraConsentAcceptRequest interfaces.HydraConsentAcceptRequest) (interfaces.HydraConsentAcceptResponse, error) {
  var hydraConsentAcceptResponse interfaces.HydraConsentAcceptResponse

  body, _ := json.Marshal(hydraConsentAcceptRequest)

  request, _ := http.NewRequest("PUT", config.Hydra.ConsentRequestAcceptUrl, bytes.NewBuffer(body))

  query := request.URL.Query()
  query.Add("consent_challenge", challenge)
  request.URL.RawQuery = query.Encode()

  response, _ := client.Do(request)

  responseData, _ := ioutil.ReadAll(response.Body)

  json.Unmarshal(responseData, &hydraConsentAcceptResponse)

  return hydraConsentAcceptResponse, nil
}

func RejectConsent(challenge string, client *http.Client, hydraConsentRejectRequest interfaces.HydraConsentRejectRequest) (interfaces.HydraConsentRejectResponse, error) {
  var hydraConsentRejectResponse interfaces.HydraConsentRejectResponse

  body, _ := json.Marshal(hydraConsentRejectRequest)

  request, _ := http.NewRequest("PUT", config.Hydra.ConsentRequestAcceptUrl, bytes.NewBuffer(body))

  query := request.URL.Query()
  query.Add("consent_challenge", challenge)
  request.URL.RawQuery = query.Encode()

  response, _ := client.Do(request)

  responseData, _ := ioutil.ReadAll(response.Body)

  json.Unmarshal(responseData, &hydraConsentRejectResponse)

  return hydraConsentRejectResponse, nil
}
