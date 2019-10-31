package aap

import (
  hydra "github.com/charmixer/hydra/client"
  "github.com/charmixer/aap/config"
  "strings"
  "fmt"
)

func SyncClientToHydra(id string, scopes []string) (err error) {

  // call Hydra to get current state
  // aap is not master with client data, only client scopes
  // hydra.FetchClient()

  // FetchScopes()
  url := config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.clients") + "/" + id
  r, err := hydra.UpdateClient(url, hydra.UpdateClientRequest{
    Scope: strings.Join(scopes, " "),
  })

  if err != nil {
    return err
  }

  fmt.Println("SYNC CLIENT TO HYDRA")
  fmt.Println(r)

  return nil
}
