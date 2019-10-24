package aap

import (
  "github.com/neo4j/neo4j-go-driver/neo4j"

  "github.com/charmixer/aap/config"
)

func CreateEntity(tx neo4j.Transaction, iEntity Identity, iCreator Identity, iRequest Identity) (rEntity Identity, err error) {

  scopes := []string{
    "aap:read:grants",
    "aap:create:grants",
    "aap:delete:grants",
    "aap:read:publishes",
    "aap:create:publishes",
    "aap:delete:publishes",
    "aap:read:subscribes",
    "aap:create:subscribes",
    "aap:delete:subscribes",
    "aap:read:consents",
    "aap:create:consents",
    "aap:delete:consents",

    "mg:aap:read:grants",
    "mg:aap:create:grants",
    "mg:aap:delete:grants",
    "mg:aap:read:publishes",
    "mg:aap:create:publishes",
    "mg:aap:delete:publishes",
    "mg:aap:read:subscribes",
    "mg:aap:create:subscribes",
    "mg:aap:delete:subscribes",
    "mg:aap:read:consents",
    "mg:aap:create:consents",
    "mg:aap:delete:consents",

    "0:mg:aap:read:grants",
    "0:mg:aap:create:grants",
    "0:mg:aap:delete:grants",
    "0:mg:aap:read:publishes",
    "0:mg:aap:create:publishes",
    "0:mg:aap:delete:publishes",
    "0:mg:aap:read:subscribes",
    "0:mg:aap:create:subscribes",
    "0:mg:aap:delete:subscribes",
    "0:mg:aap:read:consents",
    "0:mg:aap:create:consents",
    "0:mg:aap:delete:consents",
  }

  var aapIdentity = Identity{
    Id: config.GetString("id"),
  }

  var grants []Grant
  for _,scope := range scopes {
    grant := Grant{
      Identity: iCreator,
      Scope: Scope{Name: scope},
      Publisher: aapIdentity,
      OnBehalfOf: iEntity,
    }

    grants = append(grants, grant)
  }

  _, err = CreateGrants(tx, grants)
  if err != nil {
    return rEntity, err
  }

  rEntity = iEntity

  return rEntity, nil
}
