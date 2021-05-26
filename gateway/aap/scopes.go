package aap

import (
	"errors"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"strings"
)

func CreateScope(tx neo4j.Transaction, iScope Scope, iRequest Identity) (rScope Scope, err error) {
	var result neo4j.Result
	var cypher string
	var params map[string]interface{}

	cypher = `
    // create scope and match it to the identity who created it
    MERGE (scope:Scope {name: $name})
    MERGE (mgscope:Scope {name: "mg:"+$name})
    MERGE (mmgscope:Scope {name: "0:mg:"+$name})

    MERGE (mmgscope)-[:MAY_GRANT]->(mmgscope)-[:MAY_GRANT]->(mgscope)-[:MAY_GRANT]->(scope)

    // Conclude
    return scope
  `

	params = map[string]interface{}{
		"name": iScope.Name,
	}

	logCypher(cypher, params)

	if result, err = tx.Run(cypher, params); err != nil {
		return Scope{}, err
	}

	if result.Next() {
		record := result.Record()

		scopeNode := record.GetByIndex(0)

		if scopeNode != nil {
			rScope = marshalNodeToScope(scopeNode.(neo4j.Node))
		}

	} else {
		return Scope{}, errors.New("Unable to create scope")
	}

	// Check if we encountered any error during record streaming
	if err = result.Err(); err != nil {
		return Scope{}, err
	}

	if err != nil {
		return Scope{}, err
	}

	return rScope, nil
}

func FetchScopes(driver neo4j.Driver, inputScopes []Scope) ([]Scope, error) {
	var err error
	var session neo4j.Session
	var neoResult interface{}

	session, err = driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	neoResult, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var result neo4j.Result

		var cypher string
		var params map[string]interface{}

		neoScopes := []string{}
		for _, scope := range inputScopes {
			neoScopes = append(neoScopes, scope.Name)
		}

		if inputScopes == nil {
			cypher = `
        MATCH (scope:Scope)

        OPTIONAL MATCH (scope)-[:CREATED_BY]->(identity:Identity)

        // Conclude
        return scope, identity
      `
		} else {
			cypher = `
        MATCH (scope:Scope)
        WHERE scope.name in split($requestedScopes, ",")

        // Conclude
        return scope
      `
			params = map[string]interface{}{
				"requestedScopes": strings.Join(neoScopes, ","),
			}
		}

		if result, err = tx.Run(cypher, params); err != nil {
			return nil, err
		}

		var outputScopes []Scope
		for result.Next() {
			record := result.Record()

			scopeNode := record.GetByIndex(0)

			if scopeNode != nil {
				scope := marshalNodeToScope(scopeNode.(neo4j.Node))

				outputScopes = append(outputScopes, scope)
			}

		}

		// Check if we encountered any error during record streaming
		if err = result.Err(); err != nil {
			return nil, err
		}
		return outputScopes, nil
	})

	if err != nil {
		return nil, err
	}

	return neoResult.([]Scope), nil
}
