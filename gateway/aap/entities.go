package aap

import (
  "strings"
  "fmt"
  "github.com/neo4j/neo4j-go-driver/neo4j"

  "github.com/charmixer/aap/config"
)

type EntityVerdict struct {
  Publisher Identity
  Requestor Identity
  Owner Identity
  Scope Scope

  Granted bool
}

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

// F_judge(I_rs, I_requestor, [I_owner], [S]) => [ I_owner, T/F, [S_granted], [S_missing] ] where I_requestor := [I_client, I_human]
func JudgeEntity(tx neo4j.Transaction, iPublisher Identity, iRequestor Identity, iOwner Identity, iScope Scope) (verdict EntityVerdict, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

/*
  var cypOwners string
  var cypScopes string

  if len(iScopes) > 0 {
    var scopes []string
    for _, scope := range iScopes {
      scopes = append(scopes, scope.Name)
    }
    cypScopes = `and scope.name in split($scopes, ",")`
    params["scopes"] = strings.Join(scopes, ",")
  }

  if len(iOwners) > 0 {
    var owners []string
    for _, e := range iOwners {
      owners = append(owners, e.Id)
    }
    cypOwners = `and owner.id in split($owners, ",")`
    params["owners"] = strings.Join(owners, ",")
  }*/

  cypher = fmt.Sprintf(`
    // Judge

    MATCH (publisher:Identity {id:$publisher})
    MATCH (requestor:Identity {id:$requestor})
    MATCH (owner:Identity) where owner.id in split($owner, " ") // Requestor must be granted a rule on behalf of publisher (root grant) or owner (subject grant)
    MATCH (scope:Scope {name:$scope})

    // Require publisher must publish scope
    MATCH (publisher)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(scope)

    // Require requestor must be granted the publish rule on behalf of owner
    MATCH (requestor)-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr)
    MATCH (gr)-[:ON_BEHALF_OF]->(owner)

    // Conclude. This only returns anything if everything match!
    RETURN publisher, requestor, owner, scope
  `)

  params["publisher"] = iPublisher.Id
  params["requestor"] = iRequestor.Id
  params["owner"] = strings.Join([]string{iPublisher.Id,iOwner.Id}, " ")
  params["scope"] = iScope.Name

  if result, err = tx.Run(cypher, params); err != nil {
    return verdict, err
  }

  for result.Next() {
    record          := result.Record()
    publisherNode   := record.GetByIndex(0)
    requestorNode   := record.GetByIndex(1)
    ownerNode       := record.GetByIndex(2)
    scopeNode       := record.GetByIndex(3)

    if publisherNode != nil && requestorNode != nil && ownerNode != nil && scopeNode != nil {
      p := marshalNodeToIdentity(publisherNode.(neo4j.Node))
      r := marshalNodeToIdentity(requestorNode.(neo4j.Node))
      o := marshalNodeToIdentity(ownerNode.(neo4j.Node))
      s := marshalNodeToScope(scopeNode.(neo4j.Node))

      verdict = EntityVerdict{
        Publisher: p,
        Requestor: r,
        Owner: o,
        Scope: s,
        Granted: true,
      }
    } else {
      verdict = EntityVerdict{
        Granted: false,
      }
    }
  }

  logCypher(cypher, params)

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return verdict, err
  }

  return verdict, nil
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

  logCypher(cypher, params)

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return nil, err
  }

  return entities, nil
}
