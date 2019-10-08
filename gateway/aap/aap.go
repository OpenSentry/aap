package aap

import (
  "strings"
  "errors"
  "github.com/neo4j/neo4j-go-driver/neo4j"
  // log "github.com/sirupsen/logrus"
  "fmt"
)

type Identity struct {
  Id       string
}
func marshalNodeToIdentity(node neo4j.Node) (Identity) {
  p := node.Props()

  return Identity{
    Id:       p["id"].(string),
  }
}

type Human struct {
  Id        string
  Password  string
  Name      string
  Email     string
  CreatedBy Identity
}
func marshalNodeToHuman(node neo4j.Node) (Human) {
  p := node.Props()

  return Human{
    Id:       p["id"].(string),
    Password: p["password"].(string),
    Name:     p["name"].(string),
    Email:    p["email"].(string),
  }
}

type Client struct {
  ClientId     string
  ClientSecret string
  Name         string
  Description  string
  CreatedBy    Identity
}
func marshalNodeToClient(node neo4j.Node) (Client) {
  p := node.Props()

  return Client{
    ClientId:     p["client_id"].(string),
    ClientSecret: p["client_secret"].(string),
    Name:         p["name"].(string),
    Description:  p["description"].(string),
  }
}

type ResourceServer struct {
  Name        string
  Audience    string
  Description string
  CreatedBy   Identity
}
func marshalNodeToResourceServer(node neo4j.Node) (ResourceServer) {
  p := node.Props()

  return ResourceServer{
    Name:        p["name"].(string),
    Audience:    p["aud"].(string),
    Description: p["description"].(string),
  }
}

type Scope struct {
  Name        string
  Title       string
  Description string
  CreatedBy   Identity
  Labels      []string
}
func marshalNodeToScope(node neo4j.Node) (Scope) {
  p := node.Props()

  return Scope{
    Name:        p["name"].(string),
    //Title:       p["title"].(string),
    //Description: p["description"].(string),
    Labels:      node.Labels(),
  }
}

type Grant struct {
  Identity Identity
  Scope Scope
  PublishedBy Identity
  GrantedBy   Identity
}

type Consent struct {
  Identity
  Client
  ResourceServer
  Scope
}

func fetchRecord(result neo4j.Result) (neo4j.Record, error) {
  var err error

  if result.Next() {
    return result.Record(), nil
  }

  if err = result.Err(); err != nil {
    return nil, err
  }

  return nil, errors.New("No records found")
}

func fetchByIdentityId(id string, tx neo4j.Transaction) (identity Identity, err error) {
  var result neo4j.Result

  cypher := `MATCH (i:Identity {id:$id}) return i`
  params := map[string]interface{}{
    "id": id,
  }

  if result, err = tx.Run(cypher, params); err != nil {
    return Identity{}, err
  }

  record, err := fetchRecord(result)

  if err != nil || record != nil {
    return Identity{}, errors.New("Identity not found")
  }

  identityNode := record.GetByIndex(0)

  if identityNode != nil {
    identity = marshalNodeToIdentity(identityNode.(neo4j.Node))
  } else {
    return Identity{}, errors.New("Identity not found")
  }

  return identity, nil
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
    MATCH (resourceOwner:Human:Identity {id:$id})
      MATCH (client:Client:Identity {client_id:$clientId})

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
      MATCH (client:Client:Identity {client_id:$clientId})
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

func CreateScope(driver neo4j.Driver, scope Scope, createdByIdentity Identity) (Scope, Identity, error) {
  var err error
  var session neo4j.Session
  var neoResult interface{}
  type NeoReturnType struct{
    Scope Scope
    Identity Identity
  }

  session, err = driver.Session(neo4j.AccessModeWrite);
  if err != nil {
    return Scope{}, Identity{}, err
  }
  defer session.Close()

  neoResult, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result
    var cypher string
    var params map[string]interface{}

    cypher = `
      // FIXME ensure user exists and return errors
      // find out who created it
      MATCH (createdByIdentity:Human:Identity {id: $createdByIdentityId})

      // create scope and match it to the identity who created it
      MERGE (scope:Grant:Scope {name: $name, title: $title, description: $description})-[:CREATED_BY]->(createdByIdentity)
      MERGE (mgscope:MayGrant:Scope {name: "mg:"+$name, title: "May grant "+$name, description: ""})-[:CREATED_BY]->(createdByIdentity)
      MERGE (mmgscope:MayGrantMayGrant:Scope {name: "mmg:"+$name, title: "May grant "+$name, description: ""})-[:CREATED_BY]->(createdByIdentity)
      MERGE (mgscope)-[:MAY_GRANT]->(scope)
      MERGE (mmgscope)-[:MAY_GRANT]->(mgscope)

      // Conclude
      return scope, createdByIdentity
    `

    params = map[string]interface{}{
      "name":                scope.Name,
      "title":               scope.Title,
      "description":         scope.Description,
      "createdByIdentityId": createdByIdentity.Id,
    }

    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var scope Scope
    var identity Identity
    if result.Next() {
      record := result.Record()

      scopeNode := record.GetByIndex(0)
      identityNode := record.GetByIndex(1)

      if scopeNode != nil {
        scope = marshalNodeToScope(scopeNode.(neo4j.Node))
      }

      if identityNode != nil {
        identity = marshalNodeToIdentity(identityNode.(neo4j.Node))
      }

    } else {
      return nil, errors.New("Unable to create scope")
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }

    return NeoReturnType{Scope: scope, Identity: identity}, nil
  })

  if err != nil {
    return Scope{}, Identity{}, err
  }

  return neoResult.(NeoReturnType).Scope, neoResult.(NeoReturnType).Identity, nil
}

func CreateGrant(tx neo4j.Transaction, iGrant Identity, iScope Scope, iPublish Identity, iRequest Identity) (rScope Scope, rPublisher Identity, rGranted Identity, rGranter Identity, err error) {
  var result neo4j.Result
  var cypher string
  var params map[string]interface{}

  cypher = `
    // FIXME ensure users exists and return errors
    MATCH (granter:Identity {id: $granterId})
    MATCH (granted:Identity {id: $grantedId})
    MATCH (publisher:Identity {id: $publisherId})
    MATCH (scope:Scope {name: $scopeName})
    MATCH (publisher)-[:IS_PUBLISHING]->(publishRule:Publish:Rule)-[:PUBLISH]->(scope)

    // create scope and match it to the identity who created it
    MERGE (granted)-[:IS_GRANTED]->(grantRule:Grant:Rule)-[:GRANTS]->(publishRule)
    MERGE (grantRule)-[:GRANTED_BY]->(granter)

    // Conclude
    return scope, publisher, granted, granter
  `

  params = map[string]interface{}{
    "scopeName":   iScope.Name,
    "granterId":   iGrant.Id,
    "grantedId":   iRequest.Id,
    "publisherId": iPublish.Id,
  }

  if result, err = tx.Run(cypher, params); err != nil {
    return rScope, rPublisher, rGranted, rGranter, err
  }

  if result.Next() {
    record := result.Record()

    scopeNode := record.GetByIndex(0)
    publisherNode := record.GetByIndex(1)
    grantedNode := record.GetByIndex(2)
    granterNode := record.GetByIndex(3)

    if scopeNode != nil {
      rScope = marshalNodeToScope(scopeNode.(neo4j.Node))
    }

    if publisherNode != nil {
      rPublisher = marshalNodeToIdentity(publisherNode.(neo4j.Node))
    }

    if grantedNode != nil {
      rGranted = marshalNodeToIdentity(grantedNode.(neo4j.Node))
    }

    if granterNode != nil {
      rGranter = marshalNodeToIdentity(granterNode.(neo4j.Node))
    }

  }

  logCypher(cypher, params)

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return rScope, rPublisher, rGranted, rGranter, err
  }

  if err != nil {
    return rScope, rPublisher, rGranted, rGranter, err
  }

  return rScope, rPublisher, rGranted, rGranter, nil
}

func FetchGrants(tx neo4j.Transaction, iGranted Identity, iFilterScopes []Scope, iFilterPublishers []Identity) (grants []Grant, err error) {
  var result neo4j.Result
  var cypher string
  var params = make(map[string]interface{})

  var where1 string
  var where2 string

  if len(iFilterScopes) > 0 {
    var filterScopes []string
    for _,e := range iFilterScopes {
      filterScopes = append(filterScopes, e.Name)
    }

    where1 = "and scope.name in split($filterScopes, \",\")"
    params["filterScopes"] = strings.Join(filterScopes, ",")
  }

  if len(iFilterPublishers) > 0 {
    var filterPublishers []string
    for _,e := range iFilterPublishers {
      filterPublishers = append(filterPublishers, e.Id)
    }

    where2 = "and publisher.id in split($filterPublishers, \",\")"
    params["filterPublishers"] = strings.Join(filterPublishers, ",")
  }

  cypher = fmt.Sprintf(`
    match (identity:Identity {id:$id})-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr:Publish:Rule)-[:PUBLISH]->(scope:Scope)
    where 1=1 %s
    match (publisher:Identity)-[:IS_PUBLISHING]->(pr)
    where 1=1 %s
    match (gr)-[:GRANTED_BY]->(granter:Identity)
    return identity, scope, publisher, granter
  `, where1, where2)

  params["id"] = iGranted.Id

  if result, err = tx.Run(cypher, params); err != nil {
    return nil, err
  }

  for result.Next() {
    record          := result.Record()
    identityNode    := record.GetByIndex(0)
    scopeNode       := record.GetByIndex(1)
    publishedByNode := record.GetByIndex(2)
    grantedByNode   := record.GetByIndex(3)

    if identityNode != nil && scopeNode != nil && publishedByNode != nil && grantedByNode != nil {
      i := marshalNodeToIdentity(identityNode.(neo4j.Node))
      s := marshalNodeToScope(scopeNode.(neo4j.Node))
      p := marshalNodeToIdentity(publishedByNode.(neo4j.Node))
      g := marshalNodeToIdentity(grantedByNode.(neo4j.Node))

      grants = append(grants, Grant{
        Identity: i,
        Scope: s,
        PublishedBy: p,
        GrantedBy: g,
      })
    }
  }

  logCypher(cypher, params)

  // Check if we encountered any error during record streaming
  if err = result.Err(); err != nil {
    return nil, err
  }

  return grants, nil
}
func UpdateScope(driver neo4j.Driver, scope Scope, createdByIdentity Identity) (Scope, Identity, error) {
  var err error
  var session neo4j.Session
  var neoResult interface{}
  type NeoReturnType struct{
    Scope Scope
    Identity Identity
  }

  session, err = driver.Session(neo4j.AccessModeWrite);
  if err != nil {
    return Scope{}, Identity{}, err
  }
  defer session.Close()

  neoResult, err = session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
    var result neo4j.Result
    var cypher string
    var params map[string]interface{}

    cypher = `
      // FIXME ensure user exists and return errors
      // find out who created it
      MATCH (createdByIdentity:Human:Identity {id: $createdByIdentityId})

      // create scope and match it to the identity who created it
      MERGE (scope:Scope {name: $name, title: $title, description: $description})-[:CREATED_BY]->(createdByIdentity)

      // Conclude
      return scope, createdByIdentity
    `

    params = map[string]interface{}{
      "name":                scope.Name,
      "title":               scope.Title,
      "description":         scope.Description,
      "createdByIdentityId": createdByIdentity.Id,
    }

    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var scope Scope
    var identity Identity
    if result.Next() {
      record := result.Record()

      scopeNode := record.GetByIndex(0)
      identityNode := record.GetByIndex(1)

      if scopeNode != nil {
        scope = marshalNodeToScope(scopeNode.(neo4j.Node))
      }

      if identityNode != nil {
        identity = marshalNodeToIdentity(identityNode.(neo4j.Node))
      }

    } else {
      return nil, errors.New("Unable to create scope")
    }

    // Check if we encountered any error during record streaming
    if err = result.Err(); err != nil {
      return nil, err
    }

    return NeoReturnType{Scope: scope, Identity: identity}, nil
  })

  if err != nil {
    return Scope{}, Identity{}, err
  }

  return neoResult.(NeoReturnType).Scope, neoResult.(NeoReturnType).Identity, nil
}

func FetchScopes(driver neo4j.Driver, inputScopes []Scope) ([]Scope, error) {
  var err error
  var session neo4j.Session
  var neoResult interface{}

  session, err = driver.Session(neo4j.AccessModeWrite);
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

        OPTIONAL MATCH (scope)-[:CREATED_BY]->(identity:Identity)

        // Conclude
        return scope, identity
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
      identityNode := record.GetByIndex(1)

      if scopeNode != nil {
        scope := marshalNodeToScope(scopeNode.(neo4j.Node))

        if identityNode != nil {
          scope.CreatedBy = marshalNodeToIdentity(identityNode.(neo4j.Node))
        }

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

func ReadScopesCreatedByIdentity(driver neo4j.Driver, inputScopes []Scope) ([]Scope, error) {
  var err error
  var session neo4j.Session
  var neoResult interface{}

  session, err = driver.Session(neo4j.AccessModeWrite);
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

        // Conclude
        return scope // scope.name, scope.title, scope.description
      `
    } else {
      cypher = `
        MATCH (scope:Scope)
        WHERE scope.name in split($requestedScopes, ",")

        // Conclude
        return scope // scope.name, scope.title, scope.description
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
    MATCH (rs:ResourceServer:Identity {aud:$aud}) return rs.name, rs.aud, rs.description
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
        MATCH (c:Client:Identity {client_id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer:Identity {name:$rsName})-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(s)
        return i, c, rs, s
      `
      params = map[string]interface{}{"id": resourceOwner.Id, "clientId":client.ClientId, "rsName":resourceServer.Name}
    } else {
      cypher = `
        MATCH (i:Human:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(s:Scope) WHERE p.name in split($requestedScopes, ",")
        MATCH (c:Client:Identity {client_id:$clientId})-[:IS_CONSENTED]->(cr)
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
        MATCH (c:Client:Identity {client_id:$clientId})-[:IS_CONSENTED]->(cr)
        MATCH (rs:ResourceServer:Identity)-[:IS_EXPOSED]->(:ExposeRule)-[:EXPOSE]->(s)
        return i, c, rs, s
      `
      params = map[string]interface{}{"id": resourceOwner.Id, "clientId": client.ClientId}
    } else {
      cypher = `
        MATCH (i:Human:Identity {id:$id})
        MATCH (i)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]->(s:Scope) WHERE s.name in split($requestedScopes, ",")
        MATCH (c:Client:Identity {client_id:$clientId})-[:IS_CONSENTED]->(cr)
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
