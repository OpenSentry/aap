// # AAP

// ## Requirement: Depends on Identity from IDP

// ### Scopes, IDPAPI

MATCH (i:Identity:Human {username:"root"})
// idp
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"openid", title:"Login to the IDP system", description:"Allows access to login to the IDP system"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"offline", title:"Remember me", description:"Allows access to remember your login session over a longer period of time"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:authenticate:identity", title:"Authenticate and manage your password", description:"Allows access to authenticate you and update your password"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:read:identity", title:"Read your identity", description:"Allows access to your profile information, such as email, name and profile picture"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:update:identity", title:"Update your identity", description:"Allows access to update your profile information, such as email, name and profile picture"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:delete:identity", title:"Delete your identity", description:"Allows access to delete your profile from the system"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:recover:identity", title:"Recovering of password", description:"Allows access to initialize recover password process"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:logout:identity", title:"Logout from the IDP system", description:"Allows access to log you out from the IDP system"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:read:invite", title:"Read an Invite", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:create:invite", title:"Create an Invite", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:accept:invite", title:"Accept an Invite", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:send::invite", title:"Send an Invite", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:read:follow", title:"Read follows relations", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"idp:create:follow", title:"Create follows relations", description:""})

// aap
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:authorize:identity", title:"Authorize identity", description:"Allows to authorize or reject scopes on behalf of the identity"})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:reject:identity", title:"Not used?", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:read:scopes", title:"Read scopes", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:create:scopes", title:"Create new scopes", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:update:scopes", title:"Update existing scopes", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:read:grants", title:"Read grants", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:create:grants", title:"Create new grants", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:delete:grants", title:"Remove grants", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:read:publishes", title:"Read published scopes", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:create:publishes", title:"Publish scopes", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:delete:publishes", title:"Remove published scopes", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:read:consents", title:"Read consents", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:create:consents", title:"Consent to scopes", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:delete:consents", title:"Remove consent to scopes", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:authorizations:get", title:"Not used?", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:authorizations:post", title:"Not used?", description:""})
MERGE (i)<-[:CREATED_BY]-(:Scope {name:"aap:authorizations:put", title:"Not used?", description:""})
;

// ### Publish scopes for IDP

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"openid"})
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"offline"})
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"authenticate:identity"})
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"read:identity"})
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"update:identity"})
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"delete:identity"})
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"authenticate:identity"})
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"recover:identity"})
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"logout:identity"})
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

// ### Publish scopes for AAP

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"openid"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"offline"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"offline"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:authorize:identity"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:reject:identity"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:authorize:identity"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:read:scopes"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:create:scopes"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:update:scopes"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:read:grants"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:create:grants"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:delete:grants"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:read:publishes"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:create:publishes"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:delete:publishes"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:read:consents"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:create:consents"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:delete:consents"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:authorizations:get"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:authorizations:post"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;

MATCH (i:Identity:Human {username:"root"})
MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"aap:authorizations:put"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
MERGE (er)-[:PUBLISHED_BY]->(i)
;


// ## IDPUI

// Grant IDPUI access to authenticate:identity in IDPAPI
MATCH (i:Identity:Human {username:"root"})
MATCH (idpui:Identity:Client {client_id:"idpui"})
MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"authenticate:identity"})
MERGE (idpui)-[:IS_GRANTED]->(gr:GrantRule)-[:GRANT]->(s)
MERGE (gr)-[:GRANTED_BY]->(i)
;
