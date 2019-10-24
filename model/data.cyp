// # AAP

// ## Requirement: Depends on Identity from IDP

// ### Scopes, IDPAPI

// idp
MERGE (:Scope {name:"openid", title:"Login to the IDP system", description:"Allows access to login to the IDP system"})
MERGE (:Scope {name:"offline", title:"Remember me", description:"Allows access to remember your login session over a longer period of time"})
MERGE (:Scope {name:"idp:authenticate:human", title:"Authenticate and manage your password", description:"Allows access to authenticate you and update your password"})
MERGE (:Scope {name:"idp:read:identities", title:"Read your identity", description:"Allows access to your profile information, such as email, name and profile picture"})
MERGE (:Scope {name:"idp:update:identities", title:"Update your identity", description:"Allows access to update your profile information, such as email, name and profile picture"})
MERGE (:Scope {name:"idp:delete:identities", title:"Delete your identity", description:"Allows access to delete your profile from the system"})
MERGE (:Scope {name:"idp:recover:identities", title:"Recovering of password", description:"Allows access to initialize recover password process"})
MERGE (:Scope {name:"idp:logout:identities", title:"Logout from the IDP system", description:"Allows access to log you out from the IDP system"})
MERGE (:Scope {name:"idp:read:invites", title:"Read an Invite", description:""})
MERGE (:Scope {name:"idp:create:invites", title:"Create an Invite", description:""})
MERGE (:Scope {name:"idp:accept:invites", title:"Accept an Invite", description:""})
MERGE (:Scope {name:"idp:send::invites", title:"Send an Invite", description:""})
MERGE (:Scope {name:"idp:create:resourceservers", title:"Create resource servers", description:""})
MERGE (:Scope {name:"idp:read:resourceservers", title:"Read resource servers", description:""})
MERGE (:Scope {name:"idp:update:resourceservers", title:"Update resource servers", description:""})
MERGE (:Scope {name:"idp:delete:resourceservers", title:"Delete resource servers", description:""})

// aap
// @TODO fix identity -> human
MERGE (:Scope {name:"aap:authorize:identities", title:"Authorize identity", description:"Allows to authorize or reject scopes on behalf of the identity"})
MERGE (:Scope {name:"aap:reject:identities", title:"Not used?", description:""})
MERGE (:Scope {name:"aap:read:scopes", title:"Read scopes", description:""})
MERGE (:Scope {name:"aap:create:scopes", title:"Create new scopes", description:""})
MERGE (:Scope {name:"aap:update:scopes", title:"Update existing scopes", description:""})
MERGE (:Scope {name:"aap:read:grants", title:"Read grants", description:""})
MERGE (:Scope {name:"aap:create:grants", title:"Create new grants", description:""})
MERGE (:Scope {name:"aap:delete:grants", title:"Remove grants", description:""})
MERGE (:Scope {name:"aap:read:publishes", title:"Read published scopes", description:""})
MERGE (:Scope {name:"aap:create:publishes", title:"Publish scopes", description:""})
MERGE (:Scope {name:"aap:delete:publishes", title:"Remove published scopes", description:""})
MERGE (:Scope {name:"aap:read:subscribes", title:"Read subscribed scopes", description:""})
MERGE (:Scope {name:"aap:create:subscribes", title:"Subscribe to scopes", description:""})
MERGE (:Scope {name:"aap:delete:subscribes", title:"Remove subscribed scopes", description:""})
MERGE (:Scope {name:"aap:read:consents", title:"Read consents", description:""})
MERGE (:Scope {name:"aap:create:consents", title:"Consent to scopes", description:""})
MERGE (:Scope {name:"aap:delete:consents", title:"Remove consent to scopes", description:""})

MERGE (:Scope {name:"aap:authorizations:get", title:"Not used?", description:""})
MERGE (:Scope {name:"aap:authorizations:post", title:"Not used?", description:""})
MERGE (:Scope {name:"aap:authorizations:put", title:"Not used?", description:""})
;

MATCH (s:Scope)
MERGE (s)<-[:MAY_GRANT]-(mgs:Scope {name:"mg:"+s.name, title:"May grant scope: "+s.name, description: ""})
MERGE (mgs)<-[:MAY_GRANT]-(root:Scope {name:"0:"+mgs.name, title:"May grant scope: "+mgs.name, description: ""})
;

// ### Publish scopes for IDP

// give all idp: named scope to IDP resource server

MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope)
WHERE s.name =~ "idp:.*"
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
;

MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"openid"})
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
;

MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"offline"})
MERGE (idp)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
;

// ### Publish scopes for AAP

// give all aap: named scope to AAP resource server

MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope)
WHERE s.name =~ "aap:.*"
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
;

MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"openid"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
;

MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"offline"})
MERGE (aap)-[:IS_PUBLISHING]->(er:Publish:Rule)-[:PUBLISH]->(s)
;


// create and publish all may grant scopes for each resource server

MATCH (rs:Identity:ResourceServer)-[:IS_PUBLISHING]->(pr:Publish:Rule)-[:PUBLISH]->(s:Scope)
// find definition for this scope
MATCH (rootmgs)-[:MAY_GRANT]->(mgs:Scope)-[:MAY_GRANT]->(s:Scope)

MERGE (rs)-[:IS_PUBLISHING]->(mgpr:Publish:Rule)-[:PUBLISH]->(mgs)
MERGE (mgpr)-[:MAY_GRANT]->(pr)
MERGE (rs)-[:IS_PUBLISHING]->(rootpr:Publish:Rule)-[:PUBLISH]->(rootmgs)
MERGE (rootpr)-[:MAY_GRANT]->(mgpr)
;

// Copy title and description from scopes to publish rules
MATCH (pr:Publish:Rule)-[:PUBLISH]->(s:Scope)
SET pr.title = s.title, pr.description = s.description
;

// delete title and desc from scopes
MATCH (s:Scope)
REMOVE s.title, s.description
;

MATCH (pr:Publish:Rule)
WHERE not ()-[:MAY_GRANT]->(pr)
MERGE (pr)-[:MAY_GRANT]->(pr)
;

// ## IDPUI

// Grant IDPUI access to authenticate:identity in IDPAPI
MATCH (idpui:Identity:Client {client_id:"idpui"})
MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"authenticate:identity"})
MERGE (idpui)-[:IS_GRANTED]->(gr:GrantRule)-[:GRANT]->(s)
;
