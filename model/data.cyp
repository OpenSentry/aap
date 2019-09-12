// # AAP

// ## Requirement: Depends on Identity from IDP

// ### Scopes, IDPAPI

MERGE (:Scope {name:"openid"})
MERGE (:Scope {name:"offline"})
MERGE (:Scope {name:"authenticate:identity"})
MERGE (:Scope {name:"read:identity"})
MERGE (:Scope {name:"update:identity"})
MERGE (:Scope {name:"delete:identity"})
MERGE (:Scope {name:"recover:identity"})
MERGE (:Scope {name:"logout:identity"})
;

// ### Expose scopes for IDPAPI

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {username:"idprs"})
MATCH (s:Scope {name:"openid"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(s)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {username:"idprs"})
MATCH (s:Scope {name:"offline"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(s)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {username:"idprs"})
MATCH (s:Scope {name:"authenticate:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(s)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {username:"idprs"})
MATCH (s:Scope {name:"read:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(s)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {username:"idprs"})
MATCH (s:Scope {name:"update:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(s)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {username:"idprs"})
MATCH (s:Scope {name:"delete:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(s)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {username:"idprs"})
MATCH (s:Scope {name:"authenticate:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(s)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {username:"idprs"})
MATCH (s:Scope {name:"recover:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(s)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {username:"idprs"})
MATCH (s:Scope {name:"logout:identity"})
MERGE (idp)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(s)
MERGE (er)-[:EXPOSED_BY]->(i)
;


// ## IDPUI

// Grant IDPUI access to authenticate:identity in IDPAPI
MATCH (i:Identity:Human {username:"root"})
MATCH (idpui:Identity:Client {client_id:"idpui"})
MATCH (idp:Identity:ResourceServer {username:"idprs"})
MATCH (s:Scope {name:"authenticate:identity"})
MERGE (idpui)-[:IS_GRANTED]->(gr:GrantRule)-[:GRANT]->(s)
MERGE (gr)-[:GRANTED_BY]->(i)
;


// # AAP
MERGE (:Identity:Client {client_id:"aap",  client_secret:"", name: "AAP hydra client", description:"Used by the Access and Authorization Provider api to call Hydra"})
MERGE (:Identity:Client {client_id:"aapui",  client_secret:"",  name: "AAP api client",   description:"Used by the Access and Authorization Provider UI to call the Access and Authorization API"})
;

// AAP API
MERGE (:Identity:ResourceServer {name:"aap", description:"Access and Authorization provider"})
;

// HYDRA API
MERGE (:Identity:ResourceServer {name:"hydra",  description:"OAuth2 API"})
;
