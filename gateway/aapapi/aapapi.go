package aap

import (
  "strings"
  "errors"
  "github.com/neo4j/neo4j-go-driver/neo4j"
)

type Permission struct {
  Name string `json:"name" binding:"required"`
}

type Identity struct {
  Subject string `json:"sub" binding:"required"`
  Password string `json:"password"`
  Name string `json:"name"`
  Email string `json:"email"`
}

type Client struct {
  ClientId     string `json:"client_id" binding:"required"`
  ClientSecret string `json:"client_secret,omitempty"`
  Name         string `json:"name,omitempty"`
  Description  string `json:"description,omitempty"`
}

type ResourceServer struct {
  Name string `json:"name" binding:"required"`
  Audience string `json:"aud,omitempty"`
  Description string `json:"description,omitempty"`
}

type Consent struct {
  Identity
  Client
  ResourceServer
  Permission
}

// CONSENT, CONSENTED_BY, IS_CONSENTED
func CreateConsentsForClientOnBehalfOfResourceOwner(driver neo4j.Driver, resourceOwner Identity, client Client, consentPermissions []Permission, revokePermissions []Permission) ([]Permission, error) {
  if len(consentPermissions) <= 0 && len(revokePermissions) <= 0 {
    return nil, errors.New("You must either consent permissions or revoke permissions or both, but it cannot be empty")
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
    for _, permission := range consentPermissions {
      scopes = append(scopes, permission.Name)
    }
    consentScopes := strings.Join(scopes, ",")

    scopes = []string{}
    for _, permission := range revokePermissions {
      scopes = append(scopes, permission.Name)
    }
    revokeScopes := strings.Join(scopes, ",")

    // NOTE: Ensure that user has MayGrant to permissions they are trying to grant? No! Ensure user has permission to "use" a granted permission is up to the resource server authorization check.
    var cypher string
    var params map[string]interface{}

    cypher = `
      MATCH (resourceOwner:Identity {sub:$sub})
      MATCH (client:Client {client_id:$clientId})

      WITH resourceOwner, client

      // Find all permission exposed by resource server that we want to consent on behalf of the user
      OPTIONAL MATCH (consentPermission:Permission) WHERE consentPermission.name in split($consentScopes, ",")

      WITH resourceOwner, client, consentPermission, collect(consentPermission) as consentPermissions

      // CONSENT
      FOREACH ( permission in consentPermissions |
        MERGE (resourceOwner)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]-(permission)
        MERGE (client)-[:IS_CONSENTED]->(cr)
      )

      WITH resourceOwner, client, consentPermission, consentPermissions

      // REVOKE CONSENT
      // Find all permission exposed by client that we want to revoke consent on behalf of the user
      OPTIONAL MATCH (revokePermission:Permission) WHERE revokePermission.name in split($revokeScopes, ",")
      OPTIONAL MATCH (client)-[:IS_CONSENTED]->(cr:ConsentRule)-[:CONSENTED_BY]->(resourceOwner) WHERE (cr)-[:CONSENT]->(revokePermission)
      DETACH DELETE cr

      // Conclude
      return consentPermission.name //, revokePermission.name
    `

    params = map[string]interface{}{"sub":resourceOwner.Subject, "clientId":client.ClientId, "consentScopes":consentScopes, "revokeScopes":revokeScopes,}
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var permissions []Permission
    for result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.
      name := record.GetByIndex(0)
      if name != nil {
        permission := Permission{
          Name: name.(string),
        }
        permissions = append(permissions, permission)
      }
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return permissions, nil
  })

  if err != nil {
    return nil, err
  }
  return perms.([]Permission), nil
}

func CreateConsentsToResourceServerForClientOnBehalfOfResourceOwner(driver neo4j.Driver, resourceOwner Identity, client Client, resourceServer ResourceServer, consentPermissions []Permission, revokePermissions []Permission) ([]Permission, error) {
  if len(consentPermissions) <= 0 && len(revokePermissions) <= 0 {
    return nil, errors.New("You must either consent permissions or revoke permissions or both, but it cannot be empty")
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
    for _, permission := range consentPermissions {
      scopes = append(scopes, permission.Name)
    }
    consentScopes := strings.Join(scopes, ",")

    scopes = []string{}
    for _, permission := range revokePermissions {
      scopes = append(scopes, permission.Name)
    }
    revokeScopes := strings.Join(scopes, ",")

    // NOTE: Ensure that user has MayGrant to permissions they are trying to grant? No! Ensure user has permission to "use" a granted permission is up to the resource server authorization check.
    var cypher string
    var params map[string]interface{}

    cypher = `
      MATCH (resourceOwner:Identity {sub:$sub})
      MATCH (client:Client {client_id:$clientId})
      MATCH (resourceServer:ResourceServer {name:$rsName})

      WITH resourceOwner, client, resourceServer

      // Find all permission exposed by resource server that we want to consent on behalf of the user
      OPTIONAL MATCH (resourceServer)-[:IS_EXPOSED]->(erConsent:ExposeRule)-[:EXPOSE]->(consentPermission:Permission) WHERE consentPermission.name in split($consentScopes, ",")

      WITH resourceOwner, client, resourceServer, consentPermission, collect(consentPermission) as consentPermissions

      // CONSENT
      FOREACH ( permission in consentPermissions |
        MERGE (resourceOwner)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]-(permission)
        MERGE (client)-[:IS_CONSENTED]->(cr)
      )

      WITH resourceOwner, client, resourceServer, consentPermission, consentPermissions

      // REVOKE CONSENT
      // Find all permission exposed by client that we want to revoke consent on behalf of the user
      OPTIONAL MATCH (resourceServer)-[:IS_EXPOSED]->(erConsent:ExposeRule)-[:EXPOSE]->(revokePermission:Permission) WHERE revokePermission.name in split($revokeScopes, ",")
      OPTIONAL MATCH (client)-[:IS_CONSENTED]->(cr:ConsentRule)-[:CONSENTED_BY]->(resourceOwner) WHERE (cr)-[:CONSENT]->(revokePermission)
      DETACH DELETE cr

      // Conclude
      return consentPermission.name //, revokePermission.name
    `

    params = map[string]interface{}{"sub":resourceOwner.Subject, "clientId":client.ClientId, "rsName":resourceServer.Name, "consentScopes":consentScopes, "revokeScopes":revokeScopes,}
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var permissions []Permission
    for result.Next() {
      record := result.Record()

      // NOTE: This means the statment sequence of the RETURN (possible order by)
      // https://neo4j.com/docs/driver-manual/current/cypher-values/index.html
      // If results are consumed in the same order as they are produced, records merely pass through the buffer; if they are consumed out of order, the buffer will be utilized to retain records until
      // they are consumed by the application. For large results, this may require a significant amount of memory and impact performance. For this reason, it is recommended to consume results in order wherever possible.
      name := record.GetByIndex(0)
      if name != nil {
        permission := Permission{
          Name: name.(string),
        }
        permissions = append(permissions, permission)
      }
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }
    return permissions, nil
  })

  if err != nil {
    return nil, err
  }
  return perms.([]Permission), nil
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
func FetchConsentsForResourceOwner(driver neo4j.Driver, resourceOwner Identity, requestedPermissions []Permission) ([]Consent, error) {
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
    for _, permission := range requestedPermissions {
      scopes = append(scopes, permission.Name)
    }
    requestedScopes := strings.Join(scopes, ",")


    var cypher string
    var params map[string]interface{}
    if (requestedScopes == "") {
      cypher = `
        MATCH (i:Identity {sub:$sub})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Permission)
        MATCH (c:Client)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.sub, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"sub": resourceOwner.Subject}
    } else {
      cypher = `
        MATCH (i:Identity {sub:$sub})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Permission) WHERE p.name in split($requestedScopes, ",")
        MATCH (c:Client)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.sub, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"sub":resourceOwner.Subject, "requestedScopes":requestedScopes}
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
      sub := record.GetByIndex(0).(string)
      clientId := record.GetByIndex(1).(string)
      resourceServerName := record.GetByIndex(2).(string)
      resourceServerAudience := record.GetByIndex(3).(string)
      permissionName := record.GetByIndex(4).(string)

      permission := Permission{
        Name: permissionName,
      }
      identity := Identity{
        Subject: sub,
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
        Permission: permission,
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

func FetchConsentsForResourceOwnerToClientAndResourceServer(driver neo4j.Driver, resourceOwner Identity, client Client, resourceServer ResourceServer, requestedPermissions []Permission) ([]Consent, error) {
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
    for _, permission := range requestedPermissions {
      scopes = append(scopes, permission.Name)
    }
    requestedScopes := strings.Join(scopes, ",")


    var cypher string
    var params map[string]interface{}
    if (requestedScopes == "") {
      cypher = `
        MATCH (i:Identity {sub:$sub})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Permission)
        MATCH (c:Client {client_id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.sub, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"sub": resourceOwner.Subject, "clientId":client.ClientId, "rsName":resourceServer.Name}
    } else {
      cypher = `
        MATCH (i:Identity {sub:$sub})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Permission) WHERE p.name in split($requestedScopes, ",")
        MATCH (c:Client {client_id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.sub, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"sub":resourceOwner.Subject, "clientId":client.ClientId, "rsName":resourceServer.Name, "requestedScopes":requestedScopes}
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
      sub := record.GetByIndex(0).(string)
      clientId := record.GetByIndex(1).(string)
      resourceServerName := record.GetByIndex(2).(string)
      resourceServerAudience := record.GetByIndex(3).(string)
      permissionName := record.GetByIndex(4).(string)

      permission := Permission{
        Name: permissionName,
      }
      identity := Identity{
        Subject: sub,
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
        Permission: permission,
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

func FetchConsentsForResourceOwnerToResourceServer(driver neo4j.Driver, resourceOwner Identity, resourceServer ResourceServer, requestedPermissions []Permission) ([]Consent, error) {
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
    for _, permission := range requestedPermissions {
      scopes = append(scopes, permission.Name)
    }
    requestedScopes := strings.Join(scopes, ",")


    var cypher string
    var params map[string]interface{}
    if (requestedScopes == "") {
      cypher = `
        MATCH (i:Identity {sub:$sub})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Permission)
        MATCH (c:Client)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.sub, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"sub": resourceOwner.Subject, "rsName":resourceServer.Name}
    } else {
      cypher = `
        MATCH (i:Identity {sub:$sub})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Permission) WHERE p.name in split($requestedScopes, ",")
        MATCH (c:Client)-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.sub, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"sub":resourceOwner.Subject, "rsName":resourceServer.Name, "requestedScopes":requestedScopes,}
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
      sub := record.GetByIndex(0).(string)
      clientId := record.GetByIndex(1).(string)
      resourceServerName := record.GetByIndex(2).(string)
      resourceServerAudience := record.GetByIndex(3).(string)
      permissionName := record.GetByIndex(4).(string)

      permission := Permission{
        Name: permissionName,
      }
      identity := Identity{
        Subject: sub,
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
        Permission: permission,
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


func FetchConsentsForResourceOwnerToClient(driver neo4j.Driver, resourceOwner Identity, client Client, requestedPermissions []Permission) ([]Consent, error) {
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
    for _, permission := range requestedPermissions {
      scopes = append(scopes, permission.Name)
    }
    requestedScopes := strings.Join(scopes, ",")


    var cypher string
    var params map[string]interface{}
    if (requestedScopes == "") {
      cypher = `
        MATCH (i:Identity {sub:$sub})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Permission)
        MATCH (c:Client {client_id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.sub, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"sub": resourceOwner.Subject, "clientId": client.ClientId}
    } else {
      cypher = `
        MATCH (i:Identity {sub:$sub})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(p:Permission) WHERE p.name in split($requestedScopes, ",")
        MATCH (c:Client {client_id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(p)
        return i.sub, c.client_id, rs.name, p.name
      `
      params = map[string]interface{}{"sub":resourceOwner.Subject, "clientId": client.ClientId, "requestedScopes":requestedScopes}
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
      sub := record.GetByIndex(0).(string)
      clientId := record.GetByIndex(1).(string)
      resourceServerName := record.GetByIndex(2).(string)
      resourceServerAudience := record.GetByIndex(3).(string)
      permissionName := record.GetByIndex(4).(string)

      permission := Permission{
        Name: permissionName,
      }
      identity := Identity{
        Subject: sub,
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
        Permission: permission,
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
