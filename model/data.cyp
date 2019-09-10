// # AAP

// ## Requirement: Depends on Identity from IDP

// ### Required clients
MERGE (:Client {name: "idp", client_id:"idp", client_secret:"", name: "IDP hydra client", description:"Used by the Identity Provider api to call Hydra"})
MERGE (:Client {name: "idpui",  client_id:"idpui",  client_secret:"", name: "IDP api client",   description:"Used by the Identity Provider UI to call the Identity Provider API"})
;

// ## IDPAPI
MERGE (:ResourceServer {name:"idp", aud:"idp", description:"Identity Provider"})
;

// ### Scope, IDPAPI

MERGE (:Scope {name:"openid"})
MERGE (:Scope {name:"offline"})
MERGE (:Scope {name:"authenticate:identity"})
MERGE (:Scope {name:"read:identity"})
MERGE (:Scope {name:"update:identity"})
MERGE (:Scope {name:"delete:identity"})
MERGE (:Scope {name:"recover:identity"})
MERGE (:Scope {name:"logout:identity"})
;

// ### Expose permissions for IDPAPI

MATCH (i:Identity {sub:"root"})
MATCH (idp:ResourceServer {name:"idp"})
MATCH (s:Scope {name:"openid"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idp:ResourceServer {name:"idp"})
MATCH (s:Scope {name:"offline"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idp:ResourceServer {name:"idp"})
MATCH (s:Scope {name:"authenticate:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idp:ResourceServer {name:"idp"})
MATCH (s:Scope {name:"read:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idp:ResourceServer {name:"idp"})
MATCH (s:Scope {name:"update:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idp:ResourceServer {name:"idp"})
MATCH (s:Scope {name:"delete:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idp:ResourceServer {name:"idp"})
MATCH (s:Scope {name:"authenticate:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idp:ResourceServer {name:"idp"})
MATCH (s:Scope {name:"recover:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idp:ResourceServer {name:"idp"})
MATCH (s:Scope {name:"logout:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;


// ## IDPUI

// Grant IDPUI access to authenticate:identity in IDPAPI
MATCH (i:Identity {sub:"root"})
MATCH (idpui:Client {client_id:"idpui"})
MATCH (idp:ResourceServer {name:"idp"})
MATCH (s:Scope {name:"authenticate:identity"})
MERGE (idpui)-[:IS_GRANTED]->(gr:GrantRule)-[:GRANT]->(p)
MERGE (gr)-[:GRANTED_BY]->(i)
;


// # AAP
MERGE (:Client {client_id:"aap",  client_secret:"", name: "AAP hydra client", description:"Used by the Access and Authorization Provider api to call Hydra"})
MERGE (:Client {client_id:"aapui",  client_secret:"",  name: "AAP api client",   description:"Used by the Access and Authorization Provider UI to call the Access and Authorization API"})
;

// AAP API
MERGE (:ResourceServer {name:"aap", description:"Access and Authorization provider"})
;

// HYDRA API
MERGE (:ResourceServer {name:"hydra",  description:"OAuth2 API"})
;
