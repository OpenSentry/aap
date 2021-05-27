package aap

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"strings"

	"github.com/opensentry/aap/config"
)

func CreateEntity(tx neo4j.Transaction, iEntity Identity, iCreator Identity, iScopes []string) (rEntity Identity, err error) {
	var aapIdentity = Identity{
		Id: config.GetString("id"),
	}

	var grants []Grant
	for _, scope := range iScopes {
		grant := Grant{
			Identity:   iCreator,
			Scope:      Scope{Name: scope},
			Publisher:  aapIdentity,
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

	logCypher(cypher, params)

	if result, err = tx.Run(cypher, params); err != nil {
		return nil, err
	}

	for result.Next() {
		record := result.Record()
		identityNode := record.GetByIndex(0)

		if identityNode != nil {
			i := marshalNodeToIdentity(identityNode.(neo4j.Node))

			entities = append(entities, i)
		}
	}

	// Check if we encountered any error during record streaming
	if err = result.Err(); err != nil {
		return nil, err
	}

	return entities, nil
}
