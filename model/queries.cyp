// Find permission exposed by client
MATCH (c:Client)-[IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p:Permission) return c, er, p

// Find permission exposed by client and by whom
MATCH (c:Client)-[IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p:Permission)
MATCH (er)-[:EXPOSED_BY]->(i:Identity)
return c, er, p, i

// Find permissions granted to client and by whom
MATCH (c:Client)-[IS_GRANTED]->(gr:GrantRule)-[:GRANT]->(p:Permission)
MATCH (gr)-[:GRANTED_BY]->(i:Identity)
return c, gr, p, i

// Find consents given to client by user
MATCH (c:Client)-[:IS_CONSENTED]->(cr:ConsentRule)-[:CONSENT]->(p:Permission)
MATCH (gr)-[:CONSENTED_BY]->(i:Identity)
return c, gr, p, i


// Auto consent IDPUI usage of IDPAPI on behalf of an Identity
MERGE (:Identity {sub:"user1", email:"user1@domain.com",  password:"$2a$10$SOyUCy0KLFQJa3xN90UgMe9q5wE.LfakmkCsfKLCIjRY6.CcRDYwu", name:"User 1", totp_required:false, totp_secret:"", otp_recover_code:"", otp_recover_code_expire:0, otp_delete_code:"", otp_delete_code_expire:0})

MATCH (user:Identity {sub:"user1"})
MATCH (idpui:Client {client_id:"idpui"})
MATCH (idp:Client {client_id:"idp"})

WITH idpui, idp, user

// Find all permission exposed by client that we want to consent on behalf of the user
OPTIONAL MATCH (idp)-[:IS_EXPOSED]->(exposeRule:ExposeRule)-[:EXPOSE]->(exposedPermission:Permission) WHERE exposedPermission.name in split("openid offline authenticate:identity read:identity update:identity delete:identity recover:identity logout:identity", " ")

WITH idpui, idp, user, collect(exposedPermission) as exposedPermissions

FOREACH ( permission in exposedPermissions |
  MERGE (user)<-[:CONSENTED_BY]-(cr:ConsentRule)-[:CONSENT]-(permission)
  MERGE (idpui)-[:IS_CONSENTED]->(cr)
)
;
