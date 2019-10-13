package aap

import (
  "github.com/neo4j/neo4j-go-driver/neo4j"
)

// @TODO @FIXME this is idp stuff
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
