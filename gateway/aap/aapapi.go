package aap

import (
  "strings"
  "errors"
  "github.com/neo4j/neo4j-go-driver/neo4j"
)

type Scope struct {
  Name string
  Title string
  Description string
}

type Identity struct {
  Id string
  Password string
  Name string
  Email string
}

type Client struct {
  ClientId     string
  ClientSecret string
  Name         string
  Description  string
}

type ResourceServer struct {
  Name string
  Audience string
  Description string
}

type Consent struct {
  Identity
  Client
  ResourceServer
  Scope
}

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
      MATCH (resourceOwner:Identity {id:$id})
      MATCH (client:Client {client_id:$clientId})

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
      return consentScope.name //, revokeScope.name
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
      name := record.GetByIndex(0)
      if name != nil {
        scope := Scope{
          Name: name.(string),
        }
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
      MATCH (resourceOwner:Identity {id:$id})
      MATCH (client:Client {client_id:$clientId})
      MATCH (resourceServer:ResourceServer {name:$rsName})

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
      return consentScope.name //, revokeScope.name
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
      name := record.GetByIndex(0)
      if name != nil {
        scope := Scope{
          Name: name.(string),
        }
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

func CreateScope(driver neo4j.Driver, scope Scope) (Scope, error) {
  var err error
  var session neo4j.Session
  var neoResult interface{}

  session, err = driver.Session(neo4j.AccessModeWrite);
  if err != nil {
    return Scope{}, err
  }
  defer session.Close()

  neoResult, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result

    var cypher string
    var params map[string]interface{}

    cypher = `
      MERGE (scope:Scope {name: $name, title: $title, description: $description})

      // Conclude
      return scope.name, scope.title, scope.description
    `

    params = map[string]interface{}{
      "name": scope.Name,
      "title": scope.Title,
      "description": scope.Description,
    }

    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var scope Scope
    for result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.

      name := record.GetByIndex(0)
      title := record.GetByIndex(1)
      desc := record.GetByIndex(2)
      if name != nil {
        scope = Scope{
          Name: name.(string),
          Title: title.(string),
          Description: desc.(string),
        }
      }
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return scope, nil
  })

  if err != nil {
    return Scope{}, err
  }

  return neoResult.(Scope), nil
}

func ReadScope(driver neo4j.Driver, scope Scope) (Scope, error) {
  var err error
  var session neo4j.Session
  var neoResult interface{}

  session, err = driver.Session(neo4j.AccessModeWrite);
  if err != nil {
    return Scope{}, err
  }
  defer session.Close()

  neoResult, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result

    var cypher string
    var params map[string]interface{}

    cypher = `
      MATCH (scope:Scope {name: $name})

      // Conclude
      return scope.name, scope.title, scope.description
    `

    params = map[string]interface{}{
      "name": scope.Name,
    }

    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var scope Scope
    for result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.

      name := record.GetByIndex(0)
      title := record.GetByIndex(1)
      desc := record.GetByIndex(2)
      if name != nil {
        scope = Scope{
          Name: name.(string),
          Title: title.(string),
          Description: desc.(string),
        }
      }
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return scope, nil
  })

  if err != nil {
    return Scope{}, err
  }

  return neoResult.(Scope), nil
}

func ReadScopesCreatedByIdentity(driver neo4j.Driver, identity Identity) ([]Scope, error) {
  var err error
  var session neo4j.Session
  var neoResult interface{}

  session, err = driver.Session(neo4j.AccessModeWrite);
  if err != nil {
    return []Scope{}, err
  }
  defer session.Close()

  neoResult, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result

    var cypher string
    var params map[string]interface{}

    cypher = `
      MATCH (scope:Scope {name: $name})

      // Conclude
      return scope.name, scope.title, scope.description
    `

    params = map[string]interface{}{
      "name": identity.Id,
    }

    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var scope Scope
    for result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.

      name := record.GetByIndex(0)
      title := record.GetByIndex(1)
      desc := record.GetByIndex(2)
      if name != nil {
        scope = Scope{
          Name: name.(string),
          Title: title.(string),
          Description: desc.(string),
        }
      }
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return scope, nil
  })

  if err != nil {
    return []Scope{}, err
  }

  return neoResult.([]Scope), nil
}

func FetchResourceServerByAudience(driver neo4j.Driver, aud string) (*ResourceServer, error) {
  var err error
  var session neo4j.Session
  var ret interface{}

  session, err = driver.Session(neo4j.AccessModeRead);
  if err != nil {
    return nil, err
  }
  defer session.Close()

  ret, err = session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result

    var cypher string
    var params map[string]interface{}
    cypher = `
      MATCH (rs:ResourceServer {aud:$aud}) return rs.name, rs.aud, rs.description
    `
    params = map[string]interface{}{"aud":aud}
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var rs *ResourceServer
    if result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.
      name := record.GetByIndex(0).(string)
      aud := record.GetByIndex(1).(string)
      description := record.GetByIndex(2).(string)

      rs = &ResourceServer{
        Name: name,
        Audience: aud,
        Description: description,
      }
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return rs, nil
  })
  if err != nil {
    return nil, err
  }
  return ret.(*ResourceServer), nil
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
        MATCH (i:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Scope)
        MATCH (c:Client)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.id, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"id": resourceOwner.Id}
    } else {
      cypher = `
        MATCH (i:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Scope) WHERE p.name in split($requestedScopes, ",")
        MATCH (c:Client)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.id, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"id":resourceOwner.Id, "requestedScopes":requestedScopes}
    }
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var consents []Consent
    for result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.
      id := record.GetByIndex(0).(string)
      clientId := record.GetByIndex(1).(string)
      resourceServerName := record.GetByIndex(2).(string)
      resourceServerAudience := record.GetByIndex(3).(string)
      scopeName := record.GetByIndex(4).(string)

      scope := Scope{
        Name: scopeName,
      }
      identity := Identity{
        Id: id,
      }
      client := Client{
        ClientId: clientId,
      }
      rs := ResourceServer{
        Name: resourceServerName,
        Audience: resourceServerAudience,
      }

      consents = append(consents, Consent{
        Identity: identity,
        Client: client,
        ResourceServer: rs,
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
        MATCH (i:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Scope)
        MATCH (c:Client {client_id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.id, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"id": resourceOwner.Id, "clientId":client.ClientId, "rsName":resourceServer.Name}
    } else {
      cypher = `
        MATCH (i:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Scope) WHERE p.name in split($requestedScopes, ",")
        MATCH (c:Client {client_id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.id, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"id":resourceOwner.Id, "clientId":client.ClientId, "rsName":resourceServer.Name, "requestedScopes":requestedScopes}
    }
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var consents []Consent
    for result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.
      id := record.GetByIndex(0).(string)
      clientId := record.GetByIndex(1).(string)
      resourceServerName := record.GetByIndex(2).(string)
      resourceServerAudience := record.GetByIndex(3).(string)
      scopeName := record.GetByIndex(4).(string)

      scope := Scope{
        Name: scopeName,
      }
      identity := Identity{
        Id: id,
      }
      client := Client{
        ClientId: clientId,
      }
      rs := ResourceServer{
        Name: resourceServerName,
        Audience: resourceServerAudience,
      }

      consents = append(consents, Consent{
        Identity: identity,
        Client: client,
        ResourceServer: rs,
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
        MATCH (i:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Scope)
        MATCH (c:Client)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.id, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"id": resourceOwner.Id, "rsName":resourceServer.Name}
    } else {
      cypher = `
        MATCH (i:Identity {sub:$sub})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Scope) WHERE p.name in split($requestedScopes, ",")
        MATCH (c:Client)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.id, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"id":resourceOwner.Id, "rsName":resourceServer.Name, "requestedScopes":requestedScopes,}
    }
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var consents []Consent
    for result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.
      id := record.GetByIndex(0).(string)
      clientId := record.GetByIndex(1).(string)
      resourceServerName := record.GetByIndex(2).(string)
      resourceServerAudience := record.GetByIndex(3).(string)
      scopeName := record.GetByIndex(4).(string)

      scope := Scope{
        Name: scopeName,
      }
      identity := Identity{
        Id: id,
      }
      client := Client{
        ClientId: clientId,
      }
      rs := ResourceServer{
        Name: resourceServerName,
        Audience: resourceServerAudience,
      }

      consents = append(consents, Consent{
        Identity: identity,
        Client: client,
        ResourceServer: rs,
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
        MATCH (i:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Scope)
        MATCH (c:Client {client_id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.id, c.client_id, rs.name, rs.aud, p.name
      `
      params = map[string]interface{}{"id": resourceOwner.Id, "clientId": client.ClientId}
    } else {
      cypher = `
        MATCH (i:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Scope) WHERE p.name in split($requestedScopes, ",")
        MATCH (c:Client {client_id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.id, c.client_id, rs.name, rs.aud, p.name
      `
      params = map[string]interface{}{"id":resourceOwner.Id, "clientId": client.ClientId, "requestedScopes":requestedScopes}
    }
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var consents []Consent
    for result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.
      id := record.GetByIndex(0).(string)
      clientId := record.GetByIndex(1).(string)
      resourceServerName := record.GetByIndex(2).(string)
      resourceServerAudience := record.GetByIndex(3).(string)
      scopeName := record.GetByIndex(4).(string)

      scope := Scope{
        Name: scopeName,
      }
      identity := Identity{
        Id: id,
      }
      client := Client{
        ClientId: clientId,
      }
      rs := ResourceServer{
        Name: resourceServerName,
        Audience: resourceServerAudience,
      }

      consents = append(consents, Consent{
        Identity: identity,
        Client: client,
        ResourceServer: rs,
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
