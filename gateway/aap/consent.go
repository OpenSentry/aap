package aap

import (
  "strings"
  "errors"
  "fmt"
  "github.com/neo4j/neo4j-go-driver/neo4j"
)

func CreateConsent(tx neo4j.Transaction, iOwner Identity, iSubscriber Identity, iPublisher Identity, iScopes Scope) (consent Consent, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  if iOwner.Id == "" {
    return Consent{}, errors.New("Missing Owner Id")
  }
  params["owner_id"] = iOwner.Id

  if iSubscriber.Id == "" {
    return Consent{}, errors.New("Missing Subscriber Id")
  }
  params["subscriber_id"] = iSubscriber.Id

  if iPublisher.Id == "" {
    return Consent{}, errors.New("Missing Publisher Id")
  }
  params["publisher_id"] = iPublisher.Id

  if iScopes.Name == "" {
    return Consent{}, errors.New("Missing Scope")
  }
  params["scope"] = iScopes.Name

  cypher = fmt.Sprintf(`
    MATCH (owner:Human:Identity {id:$owner_id})
    MATCH (subscriber:Client:Identity {id:$subscriber_id})
    MATCH (publisher:ResourceServer:Identity {id:$publisher_id})-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(scope:Scope {name:$scope})

    MERGE (owner)-[:CONSENT]->(cr:Consent:Rule)-[:CONSENT]->(pr)
    MERGE (cr)-[:CONSENT]->(subscriber)

    // Conclude
    RETURN publisher, scope, owner, subscriber
  `)

  logCypher(cypher, params)

  if result, err = tx.Run(cypher, params); err != nil {
    return Consent{}, err
  }

  if result.Next() {
    record         := result.Record()
    publisherNode  := record.GetByIndex(0)
    scopeNode      := record.GetByIndex(1)
    ownerNode      := record.GetByIndex(2)
    subscriberNode := record.GetByIndex(3)

    if ownerNode != nil {
      consent.Identity = marshalNodeToIdentity(ownerNode.(neo4j.Node))
    }

    if subscriberNode != nil {
      consent.Subscriber = marshalNodeToIdentity(subscriberNode.(neo4j.Node))
    }

    if publisherNode != nil {
      consent.Publisher = marshalNodeToIdentity(publisherNode.(neo4j.Node))
    }

    if scopeNode != nil {
      consent.Scope = marshalNodeToScope(scopeNode.(neo4j.Node))
    }

  } else {
    return Consent{}, errors.New("Unable to create Consent")
  }


  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return Consent{}, err
  }

  return consent, nil
}


func FetchConsents(tx neo4j.Transaction, iOwner *Identity, iSubscriber *Identity, iPublisher *Identity, iScopes []Scope) (consents []Consent, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  cypOwner := ""
  if iOwner != nil {
    cypOwner = ` {id:$owner_id} `
    params["owner_id"] = iOwner
  }

  cypSubscriber := ""
  if iSubscriber != nil {
    cypSubscriber = ` {id:$subscriber_id} `
    params["subscriber_id"] = iSubscriber
  }

  cypPublisher := ""
  if iPublisher != nil {
    cypPublisher = ` {id:$publisher_id} `
    params["publisher_id"] = iPublisher
  }

  cypScopes := ""
  if len(iScopes) > 0 {
    var scopes []string
    for _, scope := range iScopes {
      scopes = append(scopes, scope.Name)
    }
    cypScopes = ` AND scope.id in split($scopes, ",") `
    params["scopes"] = strings.Join(scopes, ",")
  }

  cypher = fmt.Sprintf(`
    MATCH (publisher:Identity %s)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(scope:Scope) WHERE 1=1 %s
    MATCH (owner:Identity %s)-[:CONSENT]->(cr:Consent:Rule)-[:CONSENT]->(pr)
    MATCH (cr)-[:CONSENT]->(subscriber:Identity %s)

    // conclude
    RETURN publisher, scope, owner, subscriber
  `, cypPublisher, cypScopes, cypOwner, cypSubscriber)

  if result, err = tx.Run(cypher, params); err != nil {
    return nil, err
  }

  var consent Consent

  for result.Next() {
    record         := result.Record()
    publisherNode  := record.GetByIndex(0)
    scopeNode      := record.GetByIndex(1)
    ownerNode      := record.GetByIndex(2)
    subscriberNode := record.GetByIndex(3)

    consent = Consent{}

    if ownerNode != nil {
      consent.Identity = marshalNodeToIdentity(ownerNode.(neo4j.Node))
    }

    if subscriberNode != nil {
      consent.Subscriber = marshalNodeToIdentity(subscriberNode.(neo4j.Node))
    }

    if publisherNode != nil {
      consent.Publisher = marshalNodeToIdentity(publisherNode.(neo4j.Node))
    }

    if scopeNode != nil {
      consent.Scope = marshalNodeToScope(scopeNode.(neo4j.Node))
      consents = append(consents, consent) // Only care about consent if scope exists
    }
  }

  logCypher(cypher, params)

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return nil, err
  }

  return consents, nil
}
