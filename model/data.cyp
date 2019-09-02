// # AAP

// ## Requirement: Depends on Identity from IDP

// ### Required clients
MERGE (:Client {name: "idpapi", client_id:"idpapi", client_secret:"", name: "IDP hydra client", description:"Used by the Identity Provider api to call Hydra"})
MERGE (:Client {name: "idpui",  client_id:"idpui",  client_secret:"", name: "IDP api client",   description:"Used by the Identity Provider UI to call the Identity Provider API"})
;

// ## IDPAPI
MERGE (:ResourceServer {name:"idpapi", aud:"idpapi", description:"Identity Provider"})
;

// ### Permission, IDPAPI

MERGE (:Permission {name:"openid"})
MERGE (:Permission {name:"offline"})
MERGE (:Permission {name:"authenticate:identity"})
MERGE (:Permission {name:"read:identity"})
MERGE (:Permission {name:"update:identity"})
MERGE (:Permission {name:"delete:identity"})
MERGE (:Permission {name:"recover:identity"})
MERGE (:Permission {name:"logout:identity"})
;

// ### Expose permissions for IDPAPI

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:ResourceServer {name:"idpapi"})
MATCH (p:Permission {name:"openid"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:ResourceServer {name:"idpapi"})
MATCH (p:Permission {name:"offline"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:ResourceServer {name:"idpapi"})
MATCH (p:Permission {name:"authenticate:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:ResourceServer {name:"idpapi"})
MATCH (p:Permission {name:"read:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:ResourceServer {name:"idpapi"})
MATCH (p:Permission {name:"update:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:ResourceServer {name:"idpapi"})
MATCH (p:Permission {name:"delete:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:ResourceServer {name:"idpapi"})
MATCH (p:Permission {name:"authenticate:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:ResourceServer {name:"idpapi"})
MATCH (p:Permission {name:"recover:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:ResourceServer {name:"idpapi"})
MATCH (p:Permission {name:"logout:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;


// ## IDPUI

// Grant IDPUI access to authenticate:identity in IDPAPI
MATCH (i:Identity {sub:"root"})
MATCH (idpui:Client {client_id:"idpui"})
MATCH (idpapi:ResourceServer {name:"idpapi"})
MATCH (p:Permission {name:"authenticate:identity"})
MERGE (idpui)-[:IS_GRANTED]->(gr:GrantRule)-[:GRANT]->(p)
MERGE (gr)-[:GRANTED_BY]->(i)
;


// # AAP
MERGE (:Client {client_id:"aapapi",  client_secret:"", name: "AAP hydra client", description:"Used by the Access and Authorization Provider api to call Hydra"})
MERGE (:Client {client_id:"aapui",  client_secret:"",  name: "AAP api client",   description:"Used by the Access and Authorization Provider UI to call the Access and Authorization API"})
;

// AAP API
MERGE (:ResourceServer {name:"aapapi", description:"Access and Authorization provider"})
;

// HYDRA API
MERGE (:ResourceServer {name:"hydra",  description:"OAuth2 API"})
;
