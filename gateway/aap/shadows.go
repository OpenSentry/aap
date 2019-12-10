package aap

import (
  "strings"
  "errors"
  "fmt"
  "github.com/neo4j/neo4j-go-driver/neo4j"

  _ "github.com/opensentry/aap/config"
)

func CreateShadow(tx neo4j.Transaction, iIdentity Identity, iShadow Identity, iNotBefore int64, iExpire int64) (rShadow Shadow, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  if iIdentity.Id == "" {
    return Shadow{}, errors.New("Missing iIdentity.Id")
  }
  params["identity"] = iIdentity.Id

  if iShadow.Id == "" {
    return Shadow{}, errors.New("Missing iShadow.Id")
  }
  params["shadow"] = iShadow.Id

  params["nbf"] = iNotBefore
  params["exp"] = iExpire

  cypher = fmt.Sprintf(`
    // CreateShadows

    MATCH (identity:Identity {id:$identity})
    MATCH (shadow:Identity {id:$shadow})

    OPTIONAL MATCH (identity)-[:IS_GRANTED]->(existingGr:Grant:Rule)-[:GRANTS]->(shadow)

    DETACH DELETE existingGr

    CREATE (identity)-[:IS_GRANTED]->(gr:Grant:Rule {nbf: $nbf, exp: $exp})-[:GRANTS]->(shadow)

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

func DeleteShadow(tx neo4j.Transaction, iShadow Shadow) (err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  if iShadow.Identity.Id == "" {
    return errors.New("Missing iShadow.Identity.Id")
  }
  params["identity"] = iShadow.Identity.Id

  if iShadow.Shadow.Id == "" {
    return errors.New("Missing iShadow.Shadow.Id")
  }
  params["shadow"] = iShadow.Shadow.Id

  cypher = fmt.Sprintf(`
    // DeleteShadows

    MATCH (identity:Identity {id:$id})
    MATCH (shadow:Identity {id:$shadow}

    MATCH (identity)-[:IS_GRANTED]-(gr:Grant:Rule)-[:GRANTS]->(shadow)

    DETACH DELETE gr
  `)

  logCypher(cypher, params)
  if result, err = tx.Run(cypher, params); err != nil {
    return err
  }

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return err
  }

  return nil
}

func FetchShadows(tx neo4j.Transaction, iFilterIdentities []Identity, iFilterShadows []Identity) (rShadows []Shadow, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  var cypFilterIdentities string
  if len(iFilterIdentities) > 0 {
    var filterIdentities []string
    for _,e := range iFilterIdentities {
      if e.Id == "" {
        continue;
      }

      filterIdentities = append(filterIdentities, e.Id)
    }

    if len(filterIdentities) > 0 {
      cypFilterIdentities = `and i.id in split($filterIdentities, ",")`
      params["filterIdentities"] = strings.Join(filterIdentities, ",")
    }
  }

  var cypFilterShadows string
  if len(iFilterShadows) > 0 {
    var filterShadows []string
    for _,e := range iFilterShadows {
      if e.Id == "" {
        continue;
      }

      filterShadows = append(filterShadows, e.Id)
    }

    if len(filterShadows) > 0 {
      cypFilterShadows = `and shadow.id in split($filterShadows, ",")`
      params["filterShadows"] = strings.Join(filterShadows, ",")
    }
  }

  cypher = fmt.Sprintf(`
    // FetchShadows

    MATCH (i:Identity)-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(shadow:Identity)
    WHERE 1=1 %s %s
    RETURN i, gr, shadow
  `, cypFilterIdentities, cypFilterShadows)

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
