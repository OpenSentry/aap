package aap

import (
  "strings"
  "errors"
  "github.com/neo4j/neo4j-go-driver/neo4j"
)

// CONSENT, CONSENTED_BY, IS_CONSENTED
func CreateConsentsForClientOnBehalfOfResourceOwner(driver neo4j.Driver, resourceOwner Identity, client Client, consentScopes []Scope, revokeScopes []Scope) ([]Scope, error) {
  if len(consentScopes) <= 0 && len(revokeScopes) <= 0 {
    return nil, errors.New("You must either consent scopes or revoke scopes or both, but it cannot be empty")
  }

  var err error
  var session neo4j.Session
  var perms interface{}

  session, err = driver.Session(neo4j.AccessModeWrite);
  if err != nil {
    return nil, err
  }
  defer session.Close()

  perms, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result
    var scopes []string

    //scopes = []string{}
    for _, scope := range consentScopes {
      scopes = append(scopes, scope.Name)
    }
    consentScopes := strings.Join(scopes, ",")

    scopes = []string{}
    for _, scope := range revokeScopes {
      scopes = append(scopes, scope.Name)
    }
    revokeScopes := strings.Join(scopes, ",")

    // NOTE: Ensure that user has MayGrant to scopes they are trying to grant? No! Ensure user has scope to "use" a granted scope is up to the resource server authorization check.
    var cypher string
    var params map[string]interface{}

    cypher = `
    MATCH (resourceOwner:Human:Identity {id:$id})
      MATCH (client:Client:Identity {id:$clientId})

      WITH resourceOwner, client

      // Find all scope exposed by resource server that we want to consent on behalf of the user
      OPTIONAL MATCH (consentScope:Scope) WHERE consentScope.name in split($consentScopes, ",")

      WITH resourceOwner, client, consentScope, collect(consentScope) as consentScopes

      // CONSENT
      FOREACH ( scope in consentScopes |
        MERGE (resourceOwner)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]-(scope)
        MERGE (client)-[:IS_CONSENTED]->(cr)
      )

      WITH resourceOwner, client, consentScope, consentScopes

      // REVOKE CONSENT
      // Find all scope exposed by client that we want to revoke consent on behalf of the user
      OPTIONAL MATCH (revokeScope:Scope) WHERE revokeScope.name in split($revokeScopes, ",")
      OPTIONAL MATCH (client)-[:IS_CONSENTED]->(cr:ConsentRule)-[:CONSENTED_BY]->(resourceOwner) WHERE (cr)-[:CONSENT]->(revokeScope)
      DETACH DELETE cr

      // Conclude
      return consentScope //, revokeScope
    `

    params = map[string]interface{}{"id":resourceOwner.Id, "clientId":client.ClientId, "consentScopes":consentScopes, "revokeScopes":revokeScopes,}
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var resultScopes []Scope
    for result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.
      consentScopeNode := record.GetByIndex(0)
      if consentScopeNode != nil {
        scope := marshalNodeToScope(consentScopeNode.(neo4j.Node))
        resultScopes = append(resultScopes, scope)
      }
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return resultScopes, nil
  })

  if err != nil {
    return nil, err
  }
  return perms.([]Scope), nil
}

func CreateConsentsToResourceServerForClientOnBehalfOfResourceOwner(driver neo4j.Driver, resourceOwner Identity, client Client, resourceServer ResourceServer, consentScopes []Scope, revokeScopes []Scope) ([]Scope, error) {
  if len(consentScopes) <= 0 && len(revokeScopes) <= 0 {
    return nil, errors.New("You must either consent scopes or revoke scopes or both, but it cannot be empty")
  }

  var err error
  var session neo4j.Session
  var perms interface{}

  session, err = driver.Session(neo4j.AccessModeWrite);
  if err != nil {
    return nil, err
  }
  defer session.Close()

  perms, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result
    var scopes []string

    //scopes = []string{}
    for _, scope := range consentScopes {
      scopes = append(scopes, scope.Name)
    }
    consentScopes := strings.Join(scopes, ",")

    scopes = []string{}
    for _, scope := range revokeScopes {
      scopes = append(scopes, scope.Name)
    }
    revokeScopes := strings.Join(scopes, ",")

    // NOTE: Ensure that user has MayGrant to scopes they are trying to grant? No! Ensure user has scope to "use" a granted scope is up to the resource server authorization check.
    var cypher string
    var params map[string]interface{}

    cypher = `
      MATCH (resourceOwner:Human:Identity {id:$id})
      MATCH (client:Client:Identity {id:$clientId})
      MATCH (resourceServer:ResourceServer:Identity {name:$rsName})

      WITH resourceOwner, client, resourceServer

      // Find all scope exposed by resource server that we want to consent on behalf of the user
      OPTIONAL MATCH (resourceServer)-[:IS_EXPOSED]->(erConsent:ExposeRule)-[:EXPOSE]->(consentScope:Scope) WHERE consentScope.name in split($consentScopes, ",")

      WITH resourceOwner, client, resourceServer, consentScope, collect(consentScope) as consentScopes

      // CONSENT
      FOREACH ( scope in consentScopes |
        MERGE (resourceOwner)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]-(scope)
        MERGE (client)-[:IS_CONSENTED]->(cr)
      )

      WITH resourceOwner, client, resourceServer, consentScope, consentScopes

      // REVOKE CONSENT
      // Find all scope exposed by client that we want to revoke consent on behalf of the user
      OPTIONAL MATCH (resourceServer)-[:IS_EXPOSED]->(erConsent:ExposeRule)-[:EXPOSE]->(revokeScope:Scope) WHERE revokeScope.name in split($revokeScopes, ",")
      OPTIONAL MATCH (client)-[:IS_CONSENTED]->(cr:ConsentRule)-[:CONSENTED_BY]->(resourceOwner) WHERE (cr)-[:CONSENT]->(revokeScope)
      DETACH DELETE cr

      // Conclude
      return consentScope //, revokeScope
    `

    params = map[string]interface{}{"id":resourceOwner.Id, "clientId":client.ClientId, "rsName":resourceServer.Name, "consentScopes":consentScopes, "revokeScopes":revokeScopes,}
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var resultScopes []Scope
    for result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.
      consentScopeNode := record.GetByIndex(0)
      if consentScopeNode != nil {
        scope := marshalNodeToScope(consentScopeNode.(neo4j.Node))
        resultScopes = append(resultScopes, scope)
      }
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return resultScopes, nil
  })

  if err != nil {
    return nil, err
  }
  return perms.([]Scope), nil
}

// IS_CONSENTED, CONSENT, CONSENTED_BY
func FetchConsentsForResourceOwner(driver neo4j.Driver, resourceOwner Identity, requestedScopes []Scope) ([]Consent, error) {
  var err error
  var session neo4j.Session
  var perms interface{}

  session, err = driver.Session(neo4j.AccessModeRead);
  if err != nil {
    return nil, err
  }
  defer session.Close()

  perms, err = session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result

    var scopes []string
    for _, scope := range requestedScopes {
      scopes = append(scopes, scope.Name)
    }
    requestedScopes := strings.Join(scopes, ",")


    var cypher string
    var params map[string]interface{}
    if (requestedScopes == "") {
      cypher = `
        MATCH (i:Human:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(s:Scope)
        MATCH (c:Client:Identity)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer:Identity)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(s)
        return i, c, rs, s
      `
      params = map[string]interface{}{"id": resourceOwner.Id}
    } else {
      cypher = `
        MATCH (i:Human:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(s:Scope) WHERE s.name in split($requestedScopes, ",")
        MATCH (c:Client:Identity)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer:Identity)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(s)
        return i, c, rs, s
      `
      params = map[string]interface{}{"id":resourceOwner.Id, "requestedScopes":requestedScopes}
    }
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var consents []Consent
    for result.Next() {
      record := result.Record()

      humanNode := record.GetByIndex(0)
      clientNode := record.GetByIndex(1)
      resourceServerNode := record.GetByIndex(2)
      scopeNode := record.GetByIndex(3)

      var human Identity
      if humanNode != nil {
        human = marshalNodeToIdentity(humanNode.(neo4j.Node)) // TODO FIXME human or identity struct?
      }

      var scope Scope
      if scopeNode != nil {
        scope = marshalNodeToScope(scopeNode.(neo4j.Node))
      }

      var resourceServer ResourceServer
      if resourceServerNode != nil {
        resourceServer = marshalNodeToResourceServer(resourceServerNode.(neo4j.Node))
      }

      var client Client
      if clientNode != nil {
        client = marshalNodeToClient(clientNode.(neo4j.Node))
      }

      consents = append(consents, Consent{
        Identity: human,
        Client: client,
        ResourceServer: resourceServer,
        Scope: scope,
      })
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return consents, nil
  })
  if err != nil {
    return nil, err
  }
  return perms.([]Consent), nil
}

func FetchConsentsForResourceOwnerToClientAndResourceServer(driver neo4j.Driver, resourceOwner Identity, client Client, resourceServer ResourceServer, requestedScopes []Scope) ([]Consent, error) {
  var err error
  var session neo4j.Session
  var perms interface{}

  session, err = driver.Session(neo4j.AccessModeRead);
  if err != nil {
    return nil, err
  }
  defer session.Close()

  perms, err = session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result

    var scopes []string
    for _, scope := range requestedScopes {
      scopes = append(scopes, scope.Name)
    }
    requestedScopes := strings.Join(scopes, ",")


    var cypher string
    var params map[string]interface{}
    if (requestedScopes == "") {
      cypher = `
        MATCH (i:Human:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(s:Scope)
        MATCH (c:Client:Identity {id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer:Identity {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(s)
        return i, c, rs, s
      `
      params = map[string]interface{}{"id": resourceOwner.Id, "clientId":client.ClientId, "rsName":resourceServer.Name}
    } else {
      cypher = `
        MATCH (i:Human:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(s:Scope) WHERE p.name in split($requestedScopes, ",")
        MATCH (c:Client:Identity {id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer:Identity {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(s)
        return i, c, rs, s
      `
      params = map[string]interface{}{"id":resourceOwner.Id, "clientId":client.ClientId, "rsName":resourceServer.Name, "requestedScopes":requestedScopes}
    }
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var consents []Consent
    for result.Next() {
      record := result.Record()

      humanNode := record.GetByIndex(0)
      clientNode := record.GetByIndex(1)
      resourceServerNode := record.GetByIndex(2)
      scopeNode := record.GetByIndex(3)

      var human Identity
      if humanNode != nil {
        human = marshalNodeToIdentity(humanNode.(neo4j.Node)) // TODO FIXME human or identity struct?
      }

      var scope Scope
      if scopeNode != nil {
        scope = marshalNodeToScope(scopeNode.(neo4j.Node))
      }

      var resourceServer ResourceServer
      if resourceServerNode != nil {
        resourceServer = marshalNodeToResourceServer(resourceServerNode.(neo4j.Node))
      }

      var client Client
      if clientNode != nil {
        client = marshalNodeToClient(clientNode.(neo4j.Node))
      }

      consents = append(consents, Consent{
        Identity: human,
        Client: client,
        ResourceServer: resourceServer,
        Scope: scope,
      })
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return consents, nil
  })
  if err != nil {
    return nil, err
  }
  return perms.([]Consent), nil
}

func FetchConsentsForResourceOwnerToResourceServer(driver neo4j.Driver, resourceOwner Identity, resourceServer ResourceServer, requestedScopes []Scope) ([]Consent, error) {
  var err error
  var session neo4j.Session
  var perms interface{}

  session, err = driver.Session(neo4j.AccessModeRead);
  if err != nil {
    return nil, err
  }
  defer session.Close()

  perms, err = session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result

    var scopes []string
    for _, scope := range requestedScopes {
      scopes = append(scopes, scope.Name)
    }
    requestedScopes := strings.Join(scopes, ",")


    var cypher string
    var params map[string]interface{}
    if (requestedScopes == "") {
      cypher = `
        MATCH (i:Human:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(s:Scope)
        MATCH (c:Client:Identity)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer:Identity {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(s)
        return i, c, rs, s
      `
      params = map[string]interface{}{"id": resourceOwner.Id, "rsName":resourceServer.Name}
    } else {
      cypher = `
        MATCH (i:Human:Identity {sub:$sub})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(s:Scope) WHERE p.name in split($requestedScopes, ",")
        MATCH (c:Client:Identity)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer:Identity {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(s)
        return i, c, rs, s
      `
      params = map[string]interface{}{"id":resourceOwner.Id, "rsName":resourceServer.Name, "requestedScopes":requestedScopes,}
    }
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var consents []Consent
    for result.Next() {
      record := result.Record()

      humanNode := record.GetByIndex(0)
      clientNode := record.GetByIndex(1)
      resourceServerNode := record.GetByIndex(2)
      scopeNode := record.GetByIndex(3)

      var human Identity
      if humanNode != nil {
        human = marshalNodeToIdentity(humanNode.(neo4j.Node)) // TODO FIXME human or identity struct?
      }

      var scope Scope
      if scopeNode != nil {
        scope = marshalNodeToScope(scopeNode.(neo4j.Node))
      }

      var resourceServer ResourceServer
      if resourceServerNode != nil {
        resourceServer = marshalNodeToResourceServer(resourceServerNode.(neo4j.Node))
      }

      var client Client
      if clientNode != nil {
        client = marshalNodeToClient(clientNode.(neo4j.Node))
      }

      consents = append(consents, Consent{
        Identity: human,
        Client: client,
        ResourceServer: resourceServer,
        Scope: scope,
      })

    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return consents, nil
  })
  if err != nil {
    return nil, err
  }
  return perms.([]Consent), nil
}


func FetchConsentsForResourceOwnerToClient(driver neo4j.Driver, resourceOwner Identity, client Client, requestedScopes []Scope) ([]Consent, error) {
  var err error
  var session neo4j.Session
  var perms interface{}

  session, err = driver.Session(neo4j.AccessModeRead);
  if err != nil {
    return nil, err
  }
  defer session.Close()

  perms, err = session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result

    var scopes []string
    for _, scope := range requestedScopes {
      scopes = append(scopes, scope.Name)
    }
    requestedScopes := strings.Join(scopes, ",")


    var cypher string
    var params map[string]interface{}
    if (requestedScopes == "") {
      cypher = `
        MATCH (i:Human:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(s:Scope)
        MATCH (c:Client:Identity {id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer:Identity)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(s)
        return i, c, rs, s
      `
      params = map[string]interface{}{"id": resourceOwner.Id, "clientId": client.ClientId}
    } else {
      cypher = `
        MATCH (i:Human:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(s:Scope) WHERE s.name in split($requestedScopes, ",")
        MATCH (c:Client:Identity {id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer:Identity)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(s)
        return i, c, rs, s
      `
      params = map[string]interface{}{"id":resourceOwner.Id, "clientId": client.ClientId, "requestedScopes":requestedScopes}
    }
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var consents []Consent
    for result.Next() {
      record := result.Record()

      humanNode := record.GetByIndex(0)
      clientNode := record.GetByIndex(1)
      resourceServerNode := record.GetByIndex(2)
      scopeNode := record.GetByIndex(3)

      var human Identity
      if humanNode != nil {
        human = marshalNodeToIdentity(humanNode.(neo4j.Node)) // TODO FIXME human or identity struct?
      }

      var scope Scope
      if scopeNode != nil {
        scope = marshalNodeToScope(scopeNode.(neo4j.Node))
      }

      var resourceServer ResourceServer
      if resourceServerNode != nil {
        resourceServer = marshalNodeToResourceServer(resourceServerNode.(neo4j.Node))
      }

      var client Client
      if clientNode != nil {
        client = marshalNodeToClient(clientNode.(neo4j.Node))
      }

      consents = append(consents, Consent{
        Identity: human,
        Client: client,
        ResourceServer: resourceServer,
        Scope: scope,
      })
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return consents, nil
  })
  if err != nil {
    return nil, err
  }
  return perms.([]Consent), nil
}
