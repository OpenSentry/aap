package aap

import (
  "strings"
  "github.com/neo4j/neo4j-go-driver/neo4j"
  "fmt"
)

func FetchPublishes(tx neo4j.Transaction, iFilterPublishers []Identity) (publishRules []PublishRule, err error) {
  var result neo4j.Result
  var params = make(map[string]interface{})
  var wherePublishers string

  if len(iFilterPublishers) > 0 {
    var filterPublishers []string
    for _,e := range iFilterPublishers {
      filterPublishers = append(filterPublishers, e.Id)
    }

    wherePublishers = "and publisher.id in split($filterPublishers, \",\")"
    params["filterPublishers"] = strings.Join(filterPublishers, ",")
  }

  cypher := fmt.Sprintf(`
    match (publisher:Identity)-[:IS_PUBLISHING]->(pr:Publish:Rule)-[:PUBLISH]->(scope:Scope)
    where 1=1 %s
    optional match (pr)-[:MAY_GRANT]->(mgpr)-[:PUBLISH]->(mgscope:Scope)
    return publisher, pr, mgpr, scope, mgscope
  `, wherePublishers)

  if result, err = tx.Run(cypher, params); err != nil {
    return nil, err
  }

  for result.Next() {
    record            := result.Record()
    publisherNode     := record.GetByIndex(0)
    publishRuleNode   := record.GetByIndex(1)
    mgPublishRuleNode := record.GetByIndex(2)
    scopeNode         := record.GetByIndex(3)
    mgScopeNode       := record.GetByIndex(4)

    if publisherNode != nil &&
    publishRuleNode != nil &&
    mgPublishRuleNode != nil &&
    scopeNode != nil &&
    mgScopeNode != nil {
      publisher     := marshalNodeToIdentity(publisherNode.(neo4j.Node))
      publishRule   := marshalNodeToPublishRule(publishRuleNode.(neo4j.Node))
      mgPublishRule := marshalNodeToPublishRule(mgPublishRuleNode.(neo4j.Node))
      scope         := marshalNodeToScope(scopeNode.(neo4j.Node))
      mgScope       := marshalNodeToScope(mgScopeNode.(neo4j.Node))

      fmt.Println(publisher, publishRule, mgPublishRule, scope, mgScope)
    }
  }

  logCypher(cypher, params)

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return nil, err
  }

  return publishRules, nil
}
