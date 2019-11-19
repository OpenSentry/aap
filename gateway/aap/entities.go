package aap

import (
  "strings"
  "fmt"
  "github.com/neo4j/neo4j-go-driver/neo4j"

  "github.com/charmixer/aap/config"
)

func CreateEntity(tx neo4j.Transaction, iEntity Identity, iCreator Identity, iRequest Identity) (rEntity Identity, err error) {

  scopes := []string{
    // "aap:read:entities:judge", // Only AAP should ever have mg:aap:read:entities:judge, 0:mg:aap:read:entities:judge
    "aap:read:grants",
    "aap:create:grants",
    "aap:delete:grants",
    "aap:read:publishes",
    "aap:create:publishes",
    "aap:delete:publishes",
    "aap:read:subscriptions",
    "aap:create:subscriptions",
    "aap:delete:subscriptions",
    "aap:read:consents",
    "aap:create:consents",
    "aap:delete:consents",

    "mg:aap:read:grants",
    "mg:aap:create:grants",
    "mg:aap:delete:grants",
    "mg:aap:read:publishes",
    "mg:aap:create:publishes",
    "mg:aap:delete:publishes",
    "mg:aap:read:subscriptions",
    "mg:aap:create:subscriptions",
    "mg:aap:delete:subscriptions",
    "mg:aap:read:consents",
    "mg:aap:create:consents",
    "mg:aap:delete:consents",

    "0:mg:aap:read:grants",
    "0:mg:aap:create:grants",
    "0:mg:aap:delete:grants",
    "0:mg:aap:read:publishes",
    "0:mg:aap:create:publishes",
    "0:mg:aap:delete:publishes",
    "0:mg:aap:read:subscriptions",
    "0:mg:aap:create:subscriptions",
    "0:mg:aap:delete:subscriptions",
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

func FetchEntities(tx neo4j.Transaction, iEntities []Identity) (entities []Identity, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  cypEntities := ""
  if len(iEntities) > 0 {
    var ids []string
    for _, entity := range iEntities {
      ids = append(ids, entity.Id)
    }
    cypEntities = ` WHERE i.id in split($ids, ",") `
    params["ids"] = strings.Join(ids, ",")
  }

  cypher = fmt.Sprintf(`
    MATCH (i:Identity) %s RETURN i
  `, cypEntities)

  logCypher(cypher, params)

  if result, err = tx.Run(cypher, params); err != nil {
    return nil, err
  }

  for result.Next() {
    record          := result.Record()
    identityNode    := record.GetByIndex(0)

    if identityNode != nil {
      i := marshalNodeToIdentity(identityNode.(neo4j.Node))

      entities = append(entities, i)
    }
  }

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return nil, err
  }

  return entities, nil
}
