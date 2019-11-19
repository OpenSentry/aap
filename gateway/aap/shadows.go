package aap

import (
  _ "strings"
  "fmt"
  "github.com/neo4j/neo4j-go-driver/neo4j"

  _ "github.com/charmixer/aap/config"
)

func CreateShadows(tx neo4j.Transaction, iShadow Shadow) (rShadow Shadow, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  cypher = fmt.Sprintf(`
    MATCH (identity:Identity {id:$id})
    MATCH (shadow:Identity {id:$shadow}

    MERGE (identity)-[:IS_GRANTED]-(gr:Grant:Rule)-[:GRANTS]->(shadow)
    MERGE (gr)-[:ON_BEHALF_OF]->(identity)

    RETURN identity, gr, shadow
  `)

  logCypher(cypher, params)
  if result, err = tx.Run(cypher, params); err != nil {
    return Shadow{}, err
  }

  if result.Next() {
    record         := result.Record()
    identityNode   := record.GetByIndex(0)
    grantRuleNode  := record.GetByIndex(1)
    shadowNode     := record.GetByIndex(2)

    if identityNode != nil {
      identity := marshalNodeToIdentity(identityNode.(neo4j.Node))
      rShadow.Identity = identity
    }

    if grantRuleNode != nil {
      grantRule := marshalNodeToGrantRule(grantRuleNode.(neo4j.Node))
      rShadow.GrantRule = grantRule
    }

    if shadowNode != nil {
      shadow := marshalNodeToIdentity(shadowNode.(neo4j.Node))
      rShadow.Shadow = shadow
    }
  }

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return Shadow{}, err
  }

  return rShadow, nil
}

func FetchShadows(tx neo4j.Transaction, iEntity Identity) (rShadows []Shadow, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  cypher = fmt.Sprintf(`
    MATCH (i:Identity {id:$id})-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(shadow:Identity)
    RETURN i, gr, shadow
  `)

  logCypher(cypher, params)
  if result, err = tx.Run(cypher, params); err != nil {
    return nil, err
  }

  for result.Next() {
    record         := result.Record()
    identityNode   := record.GetByIndex(0)
    grantRuleNode  := record.GetByIndex(1)
    shadowNode     := record.GetByIndex(2)

    var shadow Shadow

    if identityNode != nil {
      identity := marshalNodeToIdentity(identityNode.(neo4j.Node))
      shadow.Identity = identity
    }

    if grantRuleNode != nil {
      grantRule := marshalNodeToGrantRule(grantRuleNode.(neo4j.Node))
      shadow.GrantRule = grantRule
    }

    if shadowNode != nil {
      shadowIdentity := marshalNodeToIdentity(shadowNode.(neo4j.Node))
      shadow.Shadow = shadowIdentity
    }

    rShadows = append(rShadows, shadow)
  }

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return nil, err
  }

  return rShadows, nil
}
