// # Bootstrap AAP

// ## Requirement: Depends on Identity from IDP

MERGE (:Client {client_id:"idpapi", client_secret:"", name: "Identity provider API, IDPAPI",                 description:"Responsible for identification of users"})
MERGE (:Client {client_id:"idpui",  client_secret:"", name: "Identity provider UI",                          description:"Responsible for given a user an interface to communicate with the IDP API"})
MERGE (:Client {client_id:"aapapi", client_secret:"", name: "Access and Authorization provider API, AAPAPI", description:"Responseible for consents and permission of users"})
MERGE (:Client {client_id:"aapui",  client_secret:"", name: "Access and Authorization provider UI",          description:"Responsible for showing consent to the user and communicating with the AAP API"})
MERGE (:Client {client_id:"hydra",  client_secret:"", name: "OAuth2 delegator",                              description:"Responsible for handling anything OAuth2 related aswell as OpenId connect"})
;

// ## IDPAPI

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
MATCH (idpapi:Client {client_id:"idpapi"})
MATCH (p:Permission {name:"openid"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:Client {client_id:"idpapi"})
MATCH (p:Permission {name:"offline"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:Client {client_id:"idpapi"})
MATCH (p:Permission {name:"authenticate:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:Client {client_id:"idpapi"})
MATCH (p:Permission {name:"read:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:Client {client_id:"idpapi"})
MATCH (p:Permission {name:"update:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:Client {client_id:"idpapi"})
MATCH (p:Permission {name:"delete:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:Client {client_id:"idpapi"})
MATCH (p:Permission {name:"authenticate:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:Client {client_id:"idpapi"})
MATCH (p:Permission {name:"recover:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;

MATCH (i:Identity {sub:"root"})
MATCH (idpapi:Client {client_id:"idpapi"})
MATCH (p:Permission {name:"logout:identity"})
MERGE (idpapi)-[:IS_EXPOSED]->(er:ExposeRule)-[:EXPOSE]->(p)
MERGE (er)-[:EXPOSED_BY]->(i)
;


// ## IDPUI

// Grant IDPUI access to authenticate:identity in IDPAPI
MATCH (i:Identity {sub:"root"})
MATCH (idpui:Client {client_id:"idpui"})
MATCH (idpapi:Client {client_id:"idpapi"})
MATCH (p:Permission {name:"authenticate:identity"})
MERGE (idpui)-[:IS_GRANTED]->(gr:GrantRule)-[:GRANT]->(p)
MERGE (gr)-[:GRANTED_BY]->(i)
;
