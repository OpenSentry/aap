package aap

import (
  "strings"
  "fmt"
  "errors"
  "github.com/neo4j/neo4j-go-driver/neo4j"
)

func Judge(tx neo4j.Transaction, iPublisher Identity, iRequestor Identity, iScope Scope, iFilterOwners []Identity) (verdict Verdict, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  if iPublisher.Id == "" {
    return Verdict{}, errors.New("Missing iPublisher.Id")
  }
  params["publisher"] = iPublisher.Id

  if iRequestor.Id == "" {
    return Verdict{}, errors.New("Missing iRequestor.Id")
  }
  params["requestor"] = iRequestor.Id

  if iScope.Name == "" {
    return Verdict{}, errors.New("Missing iScope.Name")
  }
  params["scope"] = iScope.Name

  // NOTE: Let cypher do the distinction of the owners instead of go.

  // Always look for publisher owner grant
  iFilterOwners = append(iFilterOwners, iPublisher)

  cypFilterOwners := ""
  if len(iFilterOwners) > 0 {
    var filterOwners []string
    for _,o := range iFilterOwners {
      filterOwners = append(filterOwners, o.Id)
    }
    cypFilterOwners = ` and owner.id in split($filterOwners, " ")`
    params["filterOwners"] = strings.Join(filterOwners, " ")
  }

  cypher = fmt.Sprintf(`
    // Judge

    MATCH (publisher:Identity {id:$publisher})
    MATCH (requestor:Identity {id:$requestor})

    // Collect all publishings for scopes by publisher
    MATCH (scope:Scope {name:$scope})
    MATCH (publisher)-[:PUBLISH]->(publishing:Publish:Rule)-[:PUBLISH]->(scope)

    // Collet all granted owners for requested publishings
    MATCH (requestor)-[:IS_GRANTED]->(grant:Grant:Rule)-[:GRANTS]->(publishing), (grant)-[:ON_BEHALF_OF]->(owner:Identity)
    WHERE grant.nbf <= datetime().epochSeconds AND (grant.exp >= datetime().epochSeconds OR grant.exp = 0) %s

    // Conclude. This only returns anything if everything match!
    RETURN publisher, requestor, scope, collect(owner) as owner
  `, cypFilterOwners)

  logCypher(cypher, params)

  if result, err = tx.Run(cypher, params); err != nil {
    return verdict, err
  }

  // Deny by default
  verdict = Verdict{
    Publisher: iPublisher,
    Requestor: iRequestor,
    Scope: iScope,
    Owners: iFilterOwners,
    Granted: false,
  }

  if result.Next() {
    record          := result.Record()
    publisherNode   := record.GetByIndex(0)
    requestorNode   := record.GetByIndex(1)
    scopeNode       := record.GetByIndex(2)
    ownerNodes      := record.GetByIndex(3)

    if publisherNode != nil && requestorNode != nil && scopeNode != nil && ownerNodes != nil {
      p := marshalNodeToIdentity(publisherNode.(neo4j.Node))
      r := marshalNodeToIdentity(requestorNode.(neo4j.Node))
      s := marshalNodeToScope(scopeNode.(neo4j.Node))

      var owners []Identity
      if ownerNodes != nil {
        for _, n := range ownerNodes.([]interface{}) {
          owners = append(owners, marshalNodeToIdentity(n.(neo4j.Node)))
        }
      }

      verdict = Verdict{
        Publisher: p,
        Requestor: r,
        Scope: s,
        Owners: owners,
        Granted: true,
      }
    }
  }

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return verdict, err
  }

  return verdict, nil
}