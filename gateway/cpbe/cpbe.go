package cpbe

import (
  "github.com/neo4j/neo4j-go-driver/neo4j"
)

type Brand struct {
  Name string `json:"name" binding:"required"`
}

type System struct {
  Name string `json:"name" binding:"required"`
}

type App struct {
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

func FetchPermissionsForIdentityForApplication(driver neo4j.Driver, identity Identity, app App) ([]Permission, error) {
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

    cypher := "MATCH (a:App {name:$app})-[:Exposes]->(o:Policy)-[:Grant]->(p:Permission) MATCH (i:Identity {sub: $sub})-[:IsGranted]->(r:Rule)-[:Grant]->(o)-[:Grant]->(p) RETURN p.name ORDER BY p.name"
    params := map[string]interface{}{"sub": identity.Subject, "app": app.Name}
    if result, err = tx.Run(cypher, params); err != nil {
      return nil, err
    }

    var permissions []Permission
    if result.Next() {
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
