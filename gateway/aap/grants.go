package aap

import (
  "strings"
  "github.com/neo4j/neo4j-go-driver/neo4j"
  "fmt"
)

func CreateGrants(tx neo4j.Transaction, iGrants []Grant) (rGrants []Grant, err error) {

  for _,g := range iGrants {
    grant, err := CreateGrant(tx, g.Identity, g.Scope, g.Publisher, g.OnBehalfOf)

    if err != nil {
      return nil, err
    }

    rGrants = append(rGrants, grant)
  }

  return rGrants, nil
}

func CreateGrant(tx neo4j.Transaction, iReceive Identity, iScope Scope, iPublishedBy Identity, iOnBehalfOf Identity) (rGrant Grant, err error) {
  var result neo4j.Result
  var cypher string
  var params map[string]interface{}

  cypher = `
    // FIXME ensure users exists and return errors
    MATCH (receiver:Identity {id: $receiverId})
    MATCH (publisher:Identity {id: $publisherId})
    MATCH (obo:Identity {id: $onBehalfOfId})
    MATCH (scope:Scope {name: $scopeName})
    MATCH (publisher)-[:IS_PUBLISHING]->(publishRule:Publish:Rule)-[:PUBLISH]->(scope)

    // ensure unique rules
    CREATE (grantRule:Grant:Rule)

    // create scope and match it to the identity who created it
    MERGE (receiver)-[:IS_GRANTED]->(grantRule)-[:GRANTS]->(publishRule)
    MERGE (grantRule)-[:ON_BEHALF_OF]->(obo)

    // Conclude
    return scope, publisher, receiver, obo
  `

  params = map[string]interface{}{
    "receiverId":  iReceive.Id,
    "scopeName":   iScope.Name,
    "publisherId": iPublishedBy.Id,
    "onBehalfOfId": iOnBehalfOf.Id,
  }

  if result, err = tx.Run(cypher, params); err != nil {
    return Grant{}, err
  }

  var rScope Scope
  var rPublisher Identity
  var rReceiver Identity
  var rOnBehalfOf Identity

  if result.Next() {
    record := result.Record()

    scopeNode     := record.GetByIndex(0)
    publisherNode := record.GetByIndex(1)
    receiverNode  := record.GetByIndex(2)
    onBehalfOfNode:= record.GetByIndex(3)

    if scopeNode != nil {
      rScope = marshalNodeToScope(scopeNode.(neo4j.Node))
    }

    if publisherNode != nil {
      rPublisher = marshalNodeToIdentity(publisherNode.(neo4j.Node))
    }

    if receiverNode != nil {
      rReceiver = marshalNodeToIdentity(receiverNode.(neo4j.Node))
    }

    if onBehalfOfNode != nil {
      rOnBehalfOf = marshalNodeToIdentity(onBehalfOfNode.(neo4j.Node))
    }

    rGrant = Grant{
      Identity: rReceiver,
      Scope: rScope,
      Publisher: rPublisher,
      OnBehalfOf: rOnBehalfOf,
    }

  }

  logCypher(cypher, params)

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return Grant{}, err
  }

  if err != nil {
    return Grant{}, err
  }

  return rGrant, nil
}

func FetchGrants(tx neo4j.Transaction, iGranted Identity, iFilterScopes []Scope, iFilterPublishers []Identity, iFilterOnBehalfOf []Identity) (grants []Grant, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  var where1 string
  var where2 string
  var where3 string

  if len(iFilterScopes) > 0 {
    var filterScopes []string
    for _,e := range iFilterScopes {
      filterScopes = append(filterScopes, e.Name)
    }

    where1 = "and scope.name in split($filterScopes, \",\")"
    params["filterScopes"] = strings.Join(filterScopes, ",")
  }

  if len(iFilterPublishers) > 0 {
    var filterPublishers []string
    for _,e := range iFilterPublishers {
      filterPublishers = append(filterPublishers, e.Id)
    }

    where2 = "and publisher.id in split($filterPublishers, \",\")"
    params["filterPublishers"] = strings.Join(filterPublishers, ",")
  }

  if len(iFilterOnBehalfOf) > 0 {
    var filterOnBehalfOf []string
    for _,e := range iFilterOnBehalfOf {
      filterOnBehalfOf = append(filterOnBehalfOf, e.Id)
    }

    where3 = "and onbehalfof.id in split($filterOnBehalfOf, \",\")"
    params["filterOnBehalfOf"] = strings.Join(filterOnBehalfOf, ",")
  }

  cypher = fmt.Sprintf(`
    match (identity:Identity {id:$id})-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr:Publish:Rule)-[:PUBLISH]->(scope:Scope)
    where 1=1 %s
    match (publisher:Identity)-[:IS_PUBLISHING]->(pr)
    where 1=1 %s
    match (gr)-[:ON_BEHALF_OF]->(obo:Identity)
    where 1=1 %s

    optional match (pr)-[:MAY_GRANT]->(mgpr:Publish:Rule)-[:PUBLISH]->(mgs:Scope)

    return identity, scope, publisher, obo, collect(mgs)
  `, where1, where2, where3)

  params["id"] = iGranted.Id

  if result, err = tx.Run(cypher, params); err != nil {
    return nil, err
  }

  for result.Next() {
    record          := result.Record()
    identityNode    := record.GetByIndex(0)
    scopeNode       := record.GetByIndex(1)
    publishedByNode := record.GetByIndex(2)
    onBehalfOfNode  := record.GetByIndex(3)
    mgsNodes        := record.GetByIndex(4)

    if identityNode != nil && scopeNode != nil && publishedByNode != nil && onBehalfOfNode != nil {
      i := marshalNodeToIdentity(identityNode.(neo4j.Node))
      s := marshalNodeToScope(scopeNode.(neo4j.Node))
      p := marshalNodeToIdentity(publishedByNode.(neo4j.Node))
      o := marshalNodeToIdentity(onBehalfOfNode.(neo4j.Node))

      var mgs []Scope
      if mgsNodes != nil {
        for _, n := range mgsNodes.([]interface{}) {
          mgs = append(mgs, marshalNodeToScope(n.(neo4j.Node)))
        }
      }

      grants = append(grants, Grant{
        Identity: i,
        Scope: s,
        Publisher: p,
        OnBehalfOf: o,
        MayGrantScopes: mgs,
      })
    }
  }

  logCypher(cypher, params)

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return nil, err
  }

  return grants, nil
}
