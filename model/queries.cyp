// Find permission exposed by client
MATCH (c:Client)-[IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p:Scope) return c, er, p

// Find permission exposed by client and by whom
MATCH (c:Client)-[IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p:Scope)
MATCH (er)-[:EXPOSED_BY]->(i:Identity)
return c, er, p, i

// Find permissions granted to client and by whom
MATCH (c:Client)-[IS_GRANTED]->(gr:GrantRule)-[:GRANT]->(p:Scope)
MATCH (gr)-[:GRANTED_BY]->(i:Identity)
return c, gr, p, i

// Find consents given to client by user
MATCH (c:Client)-[:IS_CONSENTED]->(cr:ConsentRule)-[:CONSENT]->(p:Scope)
MATCH (gr)-[:CONSENTED_BY]->(i:Identity)
return c, gr, p, i


// Auto consent IDPUI usage of IDPAPI on behalf of an Identity
MERGE (:Identity {sub:"user1", email:"user1@domain.com",  password:"$2a$10$SOyUCy0KLFQJa3xN90UgMe9q5wE.LfakmkCsfKLCIjRY6.CcRDYwu", name:"User 1", totp_required:false, totp_secret:""})

MATCH (user:Identity {sub:"user1"})
MATCH (idpui:Client {client_id:"idpui"})
MATCH (idp:Client {client_id:"idp"})

WITH idpui, idp, user

// Find all permission exposed by client that we want to consent on behalf of the user
OPTIONAL MATCH (idp)-[:IS_EXPOSED]->(exposeRule:ExposeRule)-[:EXPOSE]->(exposedScope:Scope) WHERE exposedScope.name in split("openid offline authenticate:identity read:identity update:identity delete:identity recover:identity logout:identity", " ")

WITH idpui, idp, user, collect(exposedScope) as exposedScope

FOREACH ( permission in exposedScopes |
  MERGE (user)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]-(permission)
  MERGE (idpui)-[:IS_CONSENTED]->(cr)
)
;
