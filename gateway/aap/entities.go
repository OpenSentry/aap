package aap

import (
  //"strings"
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

/*

  // AAP.Id a73b547b-f26d-487b-9e3a-2574fe3403fe
  // Marc (60e1ce36-6e2b-448c-b8e1-dfd5e8af0042) vil gerne give rettigheden login til Cactus (1cd78bcd-ccac-433d-970f-8ef6b12ecd84) til Sune
  // Dette betyder at marc skal give scope 'openid' til Sune

  // For at give scope 'openid' til Sune skal Marc have lov til at give rettigheder til Cactus
  // Dette betyder at Marc skal være granted en regel som på vegne af Cactus giver 'openid'

  // Funktionen 'at give rettigheder' bliver published af AAP og er givet ved scope 'aap:create:grants'

  // Dette vil sige at følgende sti i grafen skal findes for Marc
  MATCH (aap:Identity {id:"a73b547b-f26d-487b-9e3a-2574fe3403fe"})-[:IS_PUBLISHING]->(pr:Publish:Rule)-[:PUBLISH]->(s:Scope {name:"aap:create:grants"})
  MATCH (i:Identity {id:"60e1ce36-6e2b-448c-b8e1-dfd5e8af0042"})-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr)
  MATCH (i)-[:IS_GRANTED]->(gr)-[:ON_BEHALF_OF]->(obo:Identity {id:"1cd78bcd-ccac-433d-970f-8ef6b12ecd84"})
  return aap, pr, s, i, gr, obo

*/

  cypher = fmt.Sprintf(`
    MATCH (publisher:Identity {id:$publisher})-[:IS_PUBLISHING]->(pr:Publish:Rule)-[:PUBLISH]->(scope:Scope {name:$scope})
    MATCH (requestor:Identity {id:$requestor})-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr)
    MATCH (requestor)-[:IS_GRANTED]->(gr)-[:ON_BEHALF_OF]->(owner:Identity {id:$owner})
    RETURN publisher, requestor, owner, scope, gr, pr
  `)

  params["publisher"] = iPublisher.Id
  params["requestor"] = iRequestor.Id
  params["owner"] = iOwner.Id
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