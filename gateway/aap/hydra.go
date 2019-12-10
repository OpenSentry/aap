package aap

import (
  hydra "github.com/charmixer/hydra/client"
  "github.com/neo4j/neo4j-go-driver/neo4j"
  "github.com/opensentry/aap/config"
  "github.com/opensentry/aap/utils"
  "strings"
)

func SyncScopesToHydra(tx neo4j.Transaction, iClient Identity) (err error) {
  // aap is not master with client data, only client scopes

  dbSubscriptions, err := FetchSubscriptions(tx, iClient, Identity{}, nil)

  if err != nil {
    return err
  }

  var scopes []string
  var audiences []string
  for _,s := range dbSubscriptions {
    if s.Subscriber.Id != iClient.Id {
      panic("Expected subscriptions for '" + iClient.Id + "' only, but found subscription for '" + s.Subscriber.Id + "' in response")
    }

    scopes = append(scopes, s.Scope.Name)

    if !utils.StringInSlice(s.Publisher.Id, audiences) {
      audiences = append(audiences, s.Publisher.Id)
    }
  }

  url := config.GetString("hydra.private.url") + config.GetString("hydra.private.endpoints.clients")

  client, err := hydra.ReadClient(url, iClient.Id)

  newClient := hydra.UpdateClientRequest(client)
  newClient.Scope = strings.Join(scopes, " ")
  newClient.Audience = audiences

  _, err = hydra.UpdateClient(url, iClient.Id, newClient)

  if err != nil {
    return err
  }

  return nil
}
