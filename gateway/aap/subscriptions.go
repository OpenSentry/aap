package aap

import (
  "errors"
  "github.com/neo4j/neo4j-go-driver/neo4j"
  "fmt"
)

func CreateSubscription(tx neo4j.Transaction, iSubscription Subscription, iRequest Identity) (rSubscription Subscription, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  if iSubscription.Subscriber.Id == "" {
    return Subscription{}, errors.New("Missing iSubscription.Subscriber.Id")
  }
  params["subscriber_id"] = iSubscription.Subscriber.Id

  if iSubscription.Publisher.Id == "" {
    return Subscription{}, errors.New("Missing iSubscription.Publisher.Id")
  }
  params["publisher_id"] = iSubscription.Publisher.Id

  if iSubscription.Scope.Name == "" {
    return Subscription{}, errors.New("Missing iSubscription.Scope.Name")
  }
  params["scope"] = iSubscription.Scope.Name

  cypher = fmt.Sprintf(`
    // Subscribe to a publish rule

    // Require publisher existance
    MATCH (publisher:Identity {id:$publisher_id})

    // Require subscriber existance
    MATCH (subscriber:Identity {id:$subscriber_id})

    // Require scope existance
    MATCH (scope:Scope {name:$scope})

    // Require publish rules existance
    MATCH (publisher)-[:PUBLISHES]->(pr:Publish:Rule)-[:PUBLISHES]->(scope)

    // Make the connection
    MERGE (subscriber)-[:SUBSCRIBES]-(sr:Subscribe:Rule)-[:SUBSCRIBES]->(pr)

    RETURN subscriber, publisher, scope
  `)

  logCypher(cypher, params)

  if result, err = tx.Run(cypher, params); err != nil {
    return Subscription{}, err
  }

  if result.Next() {
    record         := result.Record()
    subscriberNode := record.GetByIndex(0)
    publisherNode  := record.GetByIndex(1)
    scopeNode      := record.GetByIndex(2)

    if subscriberNode != nil {
      rSubscription.Subscriber = marshalNodeToIdentity(subscriberNode.(neo4j.Node))
    }
    if publisherNode != nil {
      rSubscription.Publisher = marshalNodeToIdentity(publisherNode.(neo4j.Node))
    }
    if scopeNode != nil {
      rSubscription.Scope = marshalNodeToScope(scopeNode.(neo4j.Node))
    }

  } else {
    return Subscription{}, errors.New("Unable to create Subscription")
  }

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return Subscription{}, err
  }

  return rSubscription, nil
}
