package aapapi

import (
  "strings"
  "errors"
  "github.com/neo4j/neo4j-go-driver/neo4j"
)

type Brand struct {
  Name string `json:"name" binding:"required"`
}

type System struct {
  Name string `json:"name" binding:"required"`
}

type Permission struct {
  Name string `json:"name" binding:"required"`
}

type Identity struct {
  Subject string `json:"sub" binding:"required"`
  Password string `json:"password"`
  Name string `json:"name"`
  Email string `json:"email"`
}

// Giving consent: The user identity grant permissions to the identity of the app
func CreateConsentsForIdentityToApplication(driver neo4j.Driver, identity Identity, appIdentity Identity, grantedPermissions []Permission, revokedPermissions []Permission) ([]Permission, error) {
  if len(grantedPermissions) <= 0 && len(revokedPermissions) <= 0 {
    return nil, errors.New("You must either grant permissions or revoke permissions or both, but it cannot be empty")
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
    for _, permission := range grantedPermissions {
      scopes = append(scopes, permission.Name)
    }
    grantedScopes := strings.Join(scopes, ",")

    scopes = []string{}
    for _, permission := range revokedPermissions {
      scopes = append(scopes, permission.Name)
    }
    revokedScopes := strings.Join(scopes, ",")

    // NOTE: Ensure that user has MayGrant to permissions they are trying to grant? No! Ensure user has permission to "use" a granted permission is up to the resource server authorization check.
    var cypher string
    var params map[string]interface{}

    cypher = `
      MATCH (i:Identity {sub:$sub})
      MATCH (app:Identity {sub:$appId})

      WITH i, app

      // Find policies exposed by the app to granted scopes and revoked scopes.
      OPTIONAL MATCH (app)-[:Exposes]->(policyGrant:Policy)-[:Grant]->(grantedPermissions:Permission) WHERE grantedPermissions.name in split($grantedScopes, ",")
      OPTIONAL MATCH (app)-[:Exposes]->(policyRevoke:Policy)-[:Grant]->(revokedPermissions:Permission) WHERE revokedPermissions.name in split($revokedScopes, ",")

      WITH i, app, grantedPermissions, policyGrant, collect(policyGrant) as grantedPolicies, revokedPermissions, policyRevoke, collect(policyRevoke) as revokedPolicies

      // BEGIN::GRANT
      FOREACH ( policy in grantedPolicies |
        MERGE (i)<-[granted_by:GrantedBy]-(r:Rule)-[:Grant]->(policyGrant) ON CREATE SET granted_by.created_dtm = timestamp()
        MERGE (app)-[is_granted:IsGranted]->(r) ON CREATE SET is_granted.created_dtm = timestamp()
        //MERGE (r)-[granted_by:GrantedBy]->(i) ON CREATE SET granted_by.created_dtm = timestamp()
      )
      // END::GRANT

      WITH i, app, grantedPermissions, policyGrant, grantedPolicies, revokedPermissions, policyRevoke, revokedPolicies

      // BEGIN::REVOKE
      OPTIONAL MATCH (app)-[:IsGranted]->(grantedRule:Rule)-[:GrantedBy]->(i) WHERE (grantedRule)-[:Grant]->(policyRevoke)
      DETACH DELETE grantedRule
      // END::REVOKE

      // Conclude
      return grantedPermissions.name //, revokedPermissions
    `

    params = map[string]interface{}{"sub": identity.Subject, "appId": appIdentity.Subject, "grantedScopes": grantedScopes, "revokedScopes": revokedScopes,}
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

// Looking for Consent: Fetch the permissions actually granted to the identity of the app for the requested permissions by the user identity
func FetchConsentsForIdentityToApplication(driver neo4j.Driver, identity Identity, appIdentity Identity, requestedPermissions []Permission) ([]Permission, error) {
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
        MATCH (app:Identity {sub:$appId})-[:Exposes]->(o:Policy)-[:Grant]->(grantedPermission:Permission)
        MATCH (app)-[:IsGranted]->(grantedRule:Rule)-[:Grant]->(o) WHERE NOT (grantedRule)<-[:IsRevoked]-()
        MATCH (grantedRule)-[:GrantedBy]->(i)
        return grantedPermission.name ORDER BY grantedPermission.name
      `
      params = map[string]interface{}{"sub": identity.Subject, "appId": appIdentity.Subject,}
    } else {
      cypher = `
        MATCH (i:Identity {sub:$sub})
        MATCH (app:Identity {sub:$appId})-[:Exposes]->(o:Policy)-[:Grant]->(grantedPermission:Permission) WHERE grantedPermission.name in split($requestedScopes, ",")
        MATCH (app)-[:IsGranted]->(grantedRule:Rule)-[:Grant]->(o) WHERE NOT (grantedRule)<-[:IsRevoked]-()
        MATCH (grantedRule)-[:GrantedBy]->(i)
        return grantedPermission.name ORDER BY grantedPermission.name
      `
      params = map[string]interface{}{"sub": identity.Subject, "appId": appIdentity.Subject, "requestedScopes": requestedScopes,}
    }
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
      name := record.GetByIndex(0).(string)
      permission := Permission{
        Name: name,
      }
      permissions = append(permissions, permission)
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

// Fetch the permissions actually granted [:IsGranted] to the identity [:Identity] for the app [:App] of the requested permissions [:Permission].
// (:App)-[:Exposes]->(o:Policy)-[:Grant]->(p:Permission)
// (:Identity)-[:IsGranted]->(:Rule)-[:Grant]->(o)-[:Grant]->(p)
func FetchPermissionsForIdentityForApplication(driver neo4j.Driver, identity Identity, appIdentity Identity, requestedPermissions []Permission) ([]Permission, error) {
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

    cypher := "MATCH (app:Identity {sub:$appId})-[:Exposes]->(o:Policy)-[:Grant]->(p:Permission) WHERE p.name in split($requestedScopes, \",\") MATCH (i:Identity {sub: $sub})-[:IsGranted]->(r:Rule)-[:Grant]->(o)-[:Grant]->(p) RETURN p.name ORDER BY p.name"
    params := map[string]interface{}{"sub": identity.Subject, "appId": appIdentity.Subject, "requestedScopes": scopes}
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
      name := record.GetByIndex(0).(string)
      permission := Permission{
        Name: name,
      }
      permissions = append(permissions, permission)
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
