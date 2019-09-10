package client

import (
  "net/http"
  "encoding/json"
  "io/ioutil"
  "bytes"
  _ "strings"
  "errors"
  _ "golang.org/x/net/context"
  _ "golang.org/x/oauth2/clientcredentials"
)

// /access

type CreateAccessRequest struct {
  Scope string `json:"scope" binding:"required"`
  Title string `json:"title"`
  Description string `json:"description"`
}

type CreateAccessResponse struct {
  Scope string `json:"scope" binding:"required"`
  Title string `json:"title"`
  Description string `json:"description"`
}

type UpdateAccessRequest struct {
  Scope string `json:"scope" binding:"required"`
  Title string `json:"title"`
  Description string `json:"description"`
}

type UpdateAccessResponse struct {
  Scope string `json:"scope" binding:"required"`
  Title string `json:"title"`
  Description string `json:"description"`
}

type ReadAccessRequest struct {
  Scope string `json:"scope" binding:"required"`
}

type ReadAccessResponse struct {
  Scope string `json:"scope" binding:"required"`
  Title string `json:"title"`
  Description string `json:"description"`
}

type DeleteAccessRequest struct {
  Scope string `json:"scope" binding:"required"`
}

type DeleteAccessResponse struct {

}

// /access/grant

type CreateAccessGrantRequest struct {
}

type CreateAccessGrantResponse struct {
}

type ReadAccessGrantRequest struct {
}

type ReadAccessGrantResponse struct {
}

type DeleteAccessGrantRequest struct {
}

type DeleteAccessGrantResponse struct {
}


func CreateAccess(accessUrl string, client *AapClient, createAccessRequest CreateAccessRequest) (*CreateAccessResponse, error) {

  body, err := json.Marshal(createAccessRequest)
  if err != nil {
    return nil, err
  }

  var data = bytes.NewBuffer(body)

  request, err := http.NewRequest("POST", accessUrl, data)
  if err != nil {
    return nil, err
  }

  response, err := client.Do(request)
  if err != nil {
     return nil, err
  }

  responseData, err := ioutil.ReadAll(response.Body)
  if err != nil {
    return nil, err
  }

  if response.StatusCode != 200 {
    return nil, errors.New("Failed to create access, status: " + string(response.StatusCode) + ", error="+string(responseData))
  }

  var createdAccessResponse CreateAccessResponse
  err = json.Unmarshal(responseData, &createdAccessResponse)
  if err != nil {
    return nil, err
  }

  return &createdAccessResponse, nil
}

func UpdateAccess(accessUrl string, client *AapClient, updateAccessRequest UpdateAccessRequest) (*UpdateAccessResponse, error) {

  body, err := json.Marshal(updateAccessRequest)
  if err != nil {
    return nil, err
  }

  var data = bytes.NewBuffer(body)

  request, err := http.NewRequest("PUT", accessUrl, data)
  if err != nil {
    return nil, err
  }

  response, err := client.Do(request)
  if err != nil {
     return nil, err
  }

  responseData, err := ioutil.ReadAll(response.Body)
  if err != nil {
    return nil, err
  }

  if response.StatusCode != 200 {
    return nil, errors.New("Failed to update access, status: " + string(response.StatusCode) + ", error="+string(responseData))
  }

  var updateAccessResponse UpdateAccessResponse
  err = json.Unmarshal(responseData, &updateAccessResponse)
  if err != nil {
    return nil, err
  }

  return &updateAccessResponse, nil
}
