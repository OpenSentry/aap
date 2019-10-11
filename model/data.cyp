// # AAP

// ## Requirement: Depends on Identity from IDP

// ### Scopes, IDPAPI

// idp
MERGE (:Scope {name:"openid", title:"Login to the IDP system", description:"Allows access to login to the IDP system"})
MERGE (:Scope {name:"offline", title:"Remember me", description:"Allows access to remember your login session over a longer period of time"})
MERGE (:Scope {name:"idp:authenticate:identity", title:"Authenticate and manage your password", description:"Allows access to authenticate you and update your password"})
MERGE (:Scope {name:"idp:read:identity", title:"Read your identity", description:"Allows access to your profile information, such as email, name and profile picture"})
MERGE (:Scope {name:"idp:update:identity", title:"Update your identity", description:"Allows access to update your profile information, such as email, name and profile picture"})
MERGE (:Scope {name:"idp:delete:identity", title:"Delete your identity", description:"Allows access to delete your profile from the system"})
MERGE (:Scope {name:"idp:recover:identity", title:"Recovering of password", description:"Allows access to initialize recover password process"})
MERGE (:Scope {name:"idp:logout:identity", title:"Logout from the IDP system", description:"Allows access to log you out from the IDP system"})
MERGE (:Scope {name:"idp:read:invite", title:"Read an Invite", description:""})
MERGE (:Scope {name:"idp:create:invite", title:"Create an Invite", description:""})
MERGE (:Scope {name:"idp:accept:invite", title:"Accept an Invite", description:""})
MERGE (:Scope {name:"idp:send::invite", title:"Send an Invite", description:""})
MERGE (:Scope {name:"idp:read:follow", title:"Read follows relations", description:""})
MERGE (:Scope {name:"idp:create:follow", title:"Create follows relations", description:""})
MERGE (:Scope {name:"idp:create:resourceservers", title:"Create resource servers", description:""})
MERGE (:Scope {name:"idp:read:resourceservers", title:"Read resource servers", description:""})
MERGE (:Scope {name:"idp:update:resourceservers", title:"Update resource servers", description:""})
MERGE (:Scope {name:"idp:delete:resourceservers", title:"Delete resource servers", description:""})

// aap
MERGE (:Scope {name:"aap:authorize:identity", title:"Authorize identity", description:"Allows to authorize or reject scopes on behalf of the identity"})
MERGE (:Scope {name:"aap:reject:identity", title:"Not used?", description:""})
MERGE (:Scope {name:"aap:read:scopes", title:"Read scopes", description:""})
MERGE (:Scope {name:"aap:create:scopes", title:"Create new scopes", description:""})
MERGE (:Scope {name:"aap:update:scopes", title:"Update existing scopes", description:""})
MERGE (:Scope {name:"aap:read:grants", title:"Read grants", description:""})
MERGE (:Scope {name:"aap:create:grants", title:"Create new grants", description:""})
MERGE (:Scope {name:"aap:delete:grants", title:"Remove grants", description:""})
MERGE (:Scope {name:"aap:read:publishes", title:"Read published scopes", description:""})
MERGE (:Scope {name:"aap:create:publishes", title:"Publish scopes", description:""})
MERGE (:Scope {name:"aap:delete:publishes", title:"Remove published scopes", description:""})
MERGE (:Scope {name:"aap:read:consents", title:"Read consents", description:""})
MERGE (:Scope {name:"aap:create:consents", title:"Consent to scopes", description:""})
MERGE (:Scope {name:"aap:delete:consents", title:"Remove consent to scopes", description:""})
MERGE (:Scope {name:"aap:authorizations:get", title:"Not used?", description:""})
MERGE (:Scope {name:"aap:authorizations:post", title:"Not used?", description:""})
MERGE (:Scope {name:"aap:authorizations:put", title:"Not used?", description:""})
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
MERGE (mgs:Scope {name:"mg:"+s.name, title:"May grant scope '"+s.name+"' to others", description:""})
MERGE (rs)-[:IS_PUBLISHING]->(mgpr:Publish:Rule)-[:PUBLISH]->(mgs)
MERGE (mgpr)-[:MAY_GRANT]->(pr)

MERGE (root:Scope {name:"0:"+mgs.name, title:"May grant scope '"+mgs.name+"' to others", description:""})
MERGE (rs)-[:IS_PUBLISHING]->(rootpr:Publish:Rule)-[:PUBLISH]->(root)
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
