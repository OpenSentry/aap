// # AAP

// ## Requirement: Depends on Identity from IDP

// ### Scopes, IDPAPI

MATCH (i:Identity:Human {username:"root"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"openid", title:"Login to the IDP system", description:"Allows access to login to the IDP system"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"offline", title:"Remember me", description:"Allows access to remember your login session over a longer period of time"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"authenticate:identity", title:"Authenticate and manage your password", description:"Allows access to authenticate you and update your password"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"read:identity", title:"Read your identity", description:"Allows access to your profile information, such as email, name and profile picture"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"update:identity", title:"Update your identity", description:"Allows access to update your profile information, such as email, name and profile picture"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"delete:identity", title:"Delete your identity", description:"Allows access to delete your profile from the system"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"recover:identity", title:"Recovering of password", description:"Allows access to initialize recover password process"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"logout:identity", title:"Logout from the IDP system", description:"Allows access to log you out from the IDP system"})
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
