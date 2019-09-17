package client

import (
  "net/http"
  "io/ioutil"
  "errors"
  "bytes"
  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/clientcredentials"
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

func callService(client *AapClient, method string, url string, data *bytes.Buffer) ([]byte, error) {
  req, err := http.NewRequest("POST", url, data)
  if err != nil {
    return nil, err
  }

  req.Header.Set("X-HTTP-Method-Override", method)

  res, err := client.Do(req)
  if err != nil {
    return nil, err
  }

  return parseResponse(res)
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
