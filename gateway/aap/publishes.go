package aap

import (
  "strings"
  "errors"
  "github.com/neo4j/neo4j-go-driver/neo4j"
  "fmt"
)

func CreatePublishes(tx neo4j.Transaction, requestedBy Identity, newPublish Publish) (publish Publish, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  if newPublish.Publisher.Id == "" {
    return Publish{}, errors.New("Missing Publish.Publiser.Id")
  }
  params["publisher_id"] = newPublish.Publisher.Id

  if newPublish.Scope.Name == "" {
    return Publish{}, errors.New("Missing Publish.Scope.Name")
  }
  params["scope"] = newPublish.Scope.Name

  if newPublish.Rule.Title == "" {
    return Publish{}, errors.New("Missing Publish.Rule.Title")
  }
  params["title"] = newPublish.Rule.Title

  if newPublish.Rule.Description == "" {
    return Publish{}, errors.New("Missing Publish.Rule.Description")
  }
  params["description"] = newPublish.Rule.Description

  // ensure scope exists
  _, err = CreateScope(tx, newPublish.Scope, requestedBy)
  if err != nil {
    return Publish{}, err
  }

  cypher = fmt.Sprintf(`
    // Require scope existance
    MATCH (s:Scope {name:$scope})
    MATCH (mg:Scope)-[:MAY_GRANT]->(s)
    MATCH (rootmg:Scope)-[:MAY_GRANT]->(mg)

    // Require publisher existance
    MATCH (publisher:Identity {id:$publisher_id})

    MERGE (publisher)-[:PUBLISH]-(pr:Publish:Rule {title:$title, description:$description})-[:PUBLISH]->(s)
    MERGE (publisher)-[:PUBLISH]-(mgpr:Publish:Rule)-[:PUBLISH]->(mg)
    MERGE (publisher)-[:PUBLISH]-(rootmgpr:Publish:Rule)-[:PUBLISH]->(rootmg)

    MERGE (rootmgpr)-[:MAY_GRANT]->(rootmgpr)-[:MAY_GRANT]->(mgpr)-[:MAY_GRANT]->(pr)

    RETURN publisher, pr, s, rootmg
  `)

  logCypher(cypher, params)

  if result, err = tx.Run(cypher, params); err != nil {
    return Publish{}, err
  }

  var rootScope Scope

  if result.Next() {
    record        := result.Record()
    publisherNode := record.GetByIndex(0)
    prNode        := record.GetByIndex(1)
    scopeNode     := record.GetByIndex(2)
    rootScopeNode := record.GetByIndex(3)

    if publisherNode != nil {
      publish.Publisher = marshalNodeToIdentity(publisherNode.(neo4j.Node))
    }
    if prNode != nil {
      publish.Rule = marshalNodeToPublishRule(prNode.(neo4j.Node))
    }
    if scopeNode != nil {
      publish.Scope = marshalNodeToScope(scopeNode.(neo4j.Node))
    }
    if rootScopeNode != nil {
      rootScope = marshalNodeToScope(rootScopeNode.(neo4j.Node))
    }

  } else {
    return Publish{}, errors.New("Unable to create Publish")
  }

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return Publish{}, err
  }

  // Grant maygrant root on new publish rule to creator
  _, err = CreateGrant(tx, requestedBy, rootScope, publish.Publisher, publish.Publisher)
  if err != nil {
    return Publish{}, err
  }

  return publish, nil
}

func FetchPublishes(tx neo4j.Transaction, iFilterPublishers []Identity) (publishes []Publish, err error) {
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
    match (publisher:Identity)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(scope:Scope)
    where 1=1 %s
    optional match (pr)-[:MAY_GRANT]->(mgpr)-[:PUBLISH]->(mgscope:Scope)
    return publisher, pr, collect(mgpr), scope, collect(mgscope)
  `, wherePublishers)

  if result, err = tx.Run(cypher, params); err != nil {
    return nil, err
  }

  for result.Next() {
    record                 := result.Record()
    publisherNode          := record.GetByIndex(0)
    publishRuleNode        := record.GetByIndex(1)
    mgPublishRuleNodeSlice := record.GetByIndex(2)
    scopeNode              := record.GetByIndex(3)
    mgScopeNodeSlice       := record.GetByIndex(4)

    var publisher      Identity
    var publishRule    PublishRule
    var mgPublishRules []PublishRule
    var scope          Scope
    var mgScopes       []Scope

    if publisherNode != nil {
      publisher     = marshalNodeToIdentity(publisherNode.(neo4j.Node))
    }
    if publishRuleNode != nil {
      publishRule   = marshalNodeToPublishRule(publishRuleNode.(neo4j.Node))
    }
    if mgPublishRuleNodeSlice != nil {
      for _,node := range mgPublishRuleNodeSlice.([]interface{}) {
        mgPublishRules = append(mgPublishRules, marshalNodeToPublishRule(node.(neo4j.Node)))
      }
    }
    if scopeNode != nil {
      scope         = marshalNodeToScope(scopeNode.(neo4j.Node))
    }
    if mgScopeNodeSlice != nil {
      for _,node := range mgScopeNodeSlice.([]interface{}) {
        mgScopes = append(mgScopes, marshalNodeToScope(node.(neo4j.Node)))
      }
    }

    publishes = append(publishes, Publish{
      Publisher:      publisher,
      Scope:          scope,
      MayGrantScopes: mgScopes,
      Rule:           publishRule,
      MayGrantRules:  mgPublishRules,
    })
  }

  logCypher(cypher, params)

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return nil, err
  }

  return publishes, nil
}
