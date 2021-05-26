package aap

import (
	"errors"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"strings"
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
    MATCH (publisher)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(scope)

    OPTIONAL MATCH (subscriber)-[:SUBSCRIBES]-(existingSr:Subscribe:Rule)-[:SUBSCRIBES]->(pr)

    DETACH DELETE existingSr

    // Make the connection
    MERGE (subscriber)-[:SUBSCRIBES]-(sr:Subscribe:Rule)-[:SUBSCRIBES]->(pr)

    RETURN subscriber, publisher, scope
  `)

	logCypher(cypher, params)

	if result, err = tx.Run(cypher, params); err != nil {
		return Subscription{}, err
	}

	if result.Next() {
		record := result.Record()
		subscriberNode := record.GetByIndex(0)
		publisherNode := record.GetByIndex(1)
		scopeNode := record.GetByIndex(2)

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

func FetchSubscriptions(tx neo4j.Transaction, iFilterSubscriber Identity, iFilterPublisher Identity, iFilterScopes []Scope) (rSubscriptions []Subscription, err error) {
	var result neo4j.Result
	var cypher string
	var params = make(map[string]interface{})

	var filterSubscriberCypher = ""
	if iFilterSubscriber.Id != "" {
		filterSubscriberCypher = " {id: $filterSubscriber}"
		params["filterSubscriber"] = iFilterSubscriber.Id
	}

	var filterPublisherCypher = ""
	if iFilterPublisher.Id != "" {
		filterPublisherCypher = " {id: $filterPublisher}"
		params["filterPublisher"] = iFilterPublisher.Id
	}

	var filterScopesCypher = ""
	if len(iFilterScopes) > 0 {
		var filterScopes []string
		for _, e := range iFilterScopes {
			filterScopes = append(filterScopes, e.Name)
		}

		filterScopesCypher = " and scope.name in split($filterScopes, \",\")"
		params["filterScopes"] = strings.Join(filterScopes, ",")
	}

	cypher = fmt.Sprintf(`
    // Fetch subscriptions

    // Require subscribers existance
    MATCH (subscriber:Identity %s)

    MATCH (subscriber)-[:SUBSCRIBES]->(sr:Subscribe:Rule)-[:SUBSCRIBES]->(pr:Publish:Rule)-[:PUBLISH]->(scope:Scope)
    WHERE 1=1 %s
    MATCH (publisher:Identity %s)-[:PUBLISH]->(pr)

    RETURN subscriber, publisher, scope
  `, filterSubscriberCypher, filterScopesCypher, filterPublisherCypher)

	logCypher(cypher, params)

	if result, err = tx.Run(cypher, params); err != nil {
		return nil, err
	}

	for result.Next() {
		record := result.Record()
		subscriberNode := record.GetByIndex(0)
		publisherNode := record.GetByIndex(1)
		scopeNode := record.GetByIndex(2)

		var rSubscription Subscription

		if subscriberNode != nil {
			rSubscription.Subscriber = marshalNodeToIdentity(subscriberNode.(neo4j.Node))
		}
		if publisherNode != nil {
			rSubscription.Publisher = marshalNodeToIdentity(publisherNode.(neo4j.Node))
		}
		if scopeNode != nil {
			rSubscription.Scope = marshalNodeToScope(scopeNode.(neo4j.Node))
		}

		rSubscriptions = append(rSubscriptions, rSubscription)
	}

	// Check if we encountered any error during record streaming
	if err = result.Err(); err != nil {
		return nil, err
	}

	return rSubscriptions, nil
}
