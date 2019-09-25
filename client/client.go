package client

import (
  "net/http"
  "io/ioutil"
  "errors"
  "bytes"
  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/clientcredentials"
  "fmt"
  "encoding/json"
  "reflect"
)

type AapClient struct {
  *http.Client
}

func NewAapClient(config *clientcredentials.Config) *AapClient {
  ctx := context.Background()
  client := config.Client(ctx)
  return &AapClient{client}
}

func NewAapClientWithUserAccessToken(config *oauth2.Config, token *oauth2.Token) *AapClient {
  ctx := context.Background()
  client := config.Client(ctx, token)
  return &AapClient{client}
}

type ErrorResponse struct {
  Code  int    `json:"code" binding:"required"`
  Error string `json:"error" binding:"required"`
}

type BulkResponse struct {
  Index  int             `json:"index" binding:"required"`
  Status int             `json:"status" binding:"required"`
  Errors []ErrorResponse `json:"errors,omitempty"`
}

func callService(client *AapClient, method string, url string, data *bytes.Buffer) ([]byte, error) {
  // for logging only
  reqData := (*data).Bytes()

  req, err := http.NewRequest("POST", url, data)
  if err != nil {
    return nil, err
  }

  req.Header.Set("X-HTTP-Method-Override", method)

  res, err := client.Do(req)

  defer res.Body.Close()

  if err != nil {
    return nil, err
  }

  response, err := parseResponse(res)

  logRequestResponse(method, url, reqData, res.Status, response, err)

  return response, nil
}

func logRequestResponse(method string, url string, reqData []byte, resStatus string, resData []byte, err error) {
  var prettyJsonRequest bytes.Buffer
  e := json.Indent(&prettyJsonRequest, reqData, "", "  ")

  if e != nil {
    fmt.Println(e.Error())
  }

  var response string
  if err == nil {
    var prettyJsonResponse bytes.Buffer
    json.Indent(&prettyJsonResponse, resData, "", "  ")
    response = string(prettyJsonResponse.Bytes())
  } else {
    response = "Error: " + err.Error()
  }

  request := string(prettyJsonRequest.Bytes())

  fmt.Println("\n============== REST DEBUGGING ===============\n" + method + " " + url + " " + request + " -> [" + resStatus + "] " + response + "\n\n")
}

func parseResponse(res *http.Response) ([]byte, error) {

  resData, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return nil, err
  }

  switch (res.StatusCode) {
  case 200:
    return resData, nil
  case 400:
    return nil, errors.New("Bad Request: " + string(resData))
  case 401:
    return nil, errors.New("Unauthorized: " + string(resData))
  case 403:
    return nil, errors.New("Forbidden: " + string(resData))
  case 404:
    return nil, errors.New("Not Found: " + string(resData))
  case 500:
    return nil, errors.New("Internal Server Error")
  default:
    return nil, errors.New("Unhandled error")
  }
}


func UnmarshalResponse(iIndex int, iResponses interface{}) (rStatus int, rOk interface{}, rErr []ErrorResponse) {
  responses := reflect.ValueOf(iResponses)
  for index := 0; index < responses.Len(); index++ {
    response := responses.Index(index)

    i := response.FieldByName("Index").Interface().(int)

    if i == iIndex {
      // found response with given index

      rStatus := response.FieldByName("Status").Interface().(int)
      err    := response.FieldByName("Errors")
      ok     := response.FieldByName("Ok")

      if ok.CanInterface() {
        rOk = ok.Interface()
      }

      if err.CanInterface() {
        rErr = err.Interface().([]ErrorResponse)
      }

      return rStatus, rOk, rErr
    }

  }

  panic("Given index not found")
}
