package client

import (
  "github.com/charmixer/aap/models"
  "fmt"
)

type ReadAccessRequest struct {
  ReadAccessRequest *models.ReadAccessRequest
}

type ReadAccessResponse struct {
  ReadAccessResponse *models.ReadAccessResponse
}

func DoSomething(in ReadAccessRequest) (ReadAccessResponse, error) {

  var r ReadAccessResponse

  fmt.Println("Do Something")

  return r, nil
}
