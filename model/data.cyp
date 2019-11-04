// # AAP

// ## Requirement: Depends on Identity from IDP

// ### Scopes, IDP

// idp
MERGE (:Scope {name:"openid", title:"Login to the IDP system", description:"Allows access to login to the IDP system"})
MERGE (:Scope {name:"offline", title:"Remember me", description:"Allows access to remember your login session over a longer period of time"})

MERGE (:Scope {name:"idp:read:identities", title:"Read your identity", description:"Allows access to your profile information, such as email, name and profile picture"})

MERGE (:Scope {name:"idp:create:invites", title:"Create an Invite", description:"Allow creation of invites to the IDP system"})
MERGE (:Scope {name:"idp:read:invites", title:"Read an Invite", description:"Allow reading invites in the IDP system"})
MERGE (:Scope {name:"idp:send::invites", title:"Send an Invite", description:"Allow sending out invites from the IDP system"})
MERGE (:Scope {name:"idp:claim:invites", title:"Claim an Invite", description:"Allow claiming of invites in the IDP system"})

MERGE (:Scope {name:"idp:create:humans", title:"Create a human", description:"Allow registering data on a claimed invite and converting the invite to a human"})
MERGE (:Scope {name:"idp:read:humans", title:"Read a human", description:""})
MERGE (:Scope {name:"idp:update:humans", title:"Update a human", description:"Allow updating data on human that is not related to the authentication process such as passwords and otp codes"})
MERGE (:Scope {name:"idp:delete:humans", title:"Delete a human", description:"Allow starting the deletion process of a human"})

MERGE (:Scope {name:"idp:create:humans:authenticate", title:"Authenticate a human", description:"Allow authentication of a human using the IDP system. This handles secrets"})
MERGE (:Scope {name:"idp:update:humans:password", title:"Change password of a human", description:"Allow changeing a password of a human. This handles secrets"})
MERGE (:Scope {name:"idp:update:humans:totp", title:"Change TOTP authentication setting", description:"Allow changeing the setup of TOTP for a human. This handles secrets"})

MERGE (:Scope {name:"idp:read:humans:logout", title:"Read a logout", description:""})
MERGE (:Scope {name:"idp:create:humans:logout", title:"Request logout", description:""})
MERGE (:Scope {name:"idp:update:humans:logout", title:"Accept logout", description:""})

MERGE (:Scope {name:"idp:create:humans:recover", title:"Recover human", description:"Allow starting the recovery process of a human."})
MERGE (:Scope {name:"idp:update:humans:recoververification", title:"Recover verification", description:"Allow handling verification of human recovery. This handles secrets"})

MERGE (:Scope {name:"idp:update:humans:deleteverification", title:"Delete verification", description:"Allow handling verification of human deletion. This handles secrets"})

MERGE (:Scope {name:"idp:read:challenges", title:"Read a Challenge", description:""})
MERGE (:Scope {name:"idp:create:challenges", title:"Create a Challenge", description:""})
MERGE (:Scope {name:"idp:update:challenges:verify", title:"Verify a Challenge", description:""})

MERGE (:Scope {name:"idp:create:clients", title:"Create clients", description:"Allow access to create clients"})
MERGE (:Scope {name:"idp:read:clients", title:"Read clients", description:"Allow access to read clients"})
//MERGE (:Scope {name:"idp:update:clients", title:"", description:""})
MERGE (:Scope {name:"idp:delete:clients", title:"Delete clients", description:"Allow access to delete clients"})

MERGE (:Scope {name:"idp:create:resourceservers", title:"Create resource servers", description:"Allow access to create resource servers"})
MERGE (:Scope {name:"idp:read:resourceservers", title:"Read resource servers", description:"Allow access to read resource servers"})
// MERGE (:Scope {name:"idp:update:resourceservers", title:"Update resource servers", description:""})
MERGE (:Scope {name:"idp:delete:resourceservers", title:"Delete resource servers", description:"Allow access to delete resource servers"})


// aap
// @TODO fix identity -> human
MERGE (:Scope {name:"aap:read:scopes", title:"Read scopes", description:""})
MERGE (:Scope {name:"aap:create:scopes", title:"Create new scopes", description:""})
MERGE (:Scope {name:"aap:update:scopes", title:"Update existing scopes", description:""})
MERGE (:Scope {name:"aap:read:grants", title:"Read grants", description:""})
MERGE (:Scope {name:"aap:create:grants", title:"Create new grants", description:""})
MERGE (:Scope {name:"aap:delete:grants", title:"Remove grants", description:""})
MERGE (:Scope {name:"aap:read:publishes", title:"Read published scopes", description:""})
MERGE (:Scope {name:"aap:create:publishes", title:"Publish scopes", description:""})
MERGE (:Scope {name:"aap:delete:publishes", title:"Remove published scopes", description:""})
MERGE (:Scope {name:"aap:read:subscriptions", title:"Read subscriptions", description:""})
MERGE (:Scope {name:"aap:create:subscriptions", title:"Create subscriptions", description:""})
MERGE (:Scope {name:"aap:delete:subscriptions", title:"Remove subscriptions", description:""})
MERGE (:Scope {name:"aap:read:consents", title:"Read consents", description:""})
MERGE (:Scope {name:"aap:create:consents", title:"Consent to scopes", description:""})
MERGE (:Scope {name:"aap:delete:consents", title:"Remove consent to scopes", description:""})
MERGE (:Scope {name:"aap:create:consents:authorize", title:"Authorize consent to entity", description:"Allow consenting to access to entity onbehalf of entity"})
MERGE (:Scope {name:"aap:create:consents:reject", title:Reject consent to entity", description:"Allow rejecting access to entity on behalf of entity"})

MERGE (:Scope {name:"aap:read:entities:judge", title:"Judge entities", description:"Allow to judge if authorized to perform request"})
MERGE (:Scope {name:"aap:create:entities", title:"Create entities", description:"Allow to create entities"})

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
MERGE (idp)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
;

MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"openid"})
MERGE (idp)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
;

MATCH (idp:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope {name:"offline"})
MERGE (idp)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
;

// ### Publish scopes for AAP

// give all aap: named scope to AAP resource server

MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope)
WHERE s.name =~ "aap:.*"
MERGE (aap)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
;

MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"openid"})
MERGE (aap)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
;

MATCH (aap:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope {name:"offline"})
MERGE (aap)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
;


// create and publish all may grant scopes for each resource server

MATCH (rs:Identity:ResourceServer)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s:Scope)
// find definition for this scope
MATCH (rootmgs)-[:MAY_GRANT]->(mgs:Scope)-[:MAY_GRANT]->(s:Scope)

MERGE (rs)-[:PUBLISH]->(mgpr:Publish:Rule)-[:PUBLISH]->(mgs)
MERGE (mgpr)-[:MAY_GRANT]->(pr)
MERGE (rs)-[:PUBLISH]->(rootpr:Publish:Rule)-[:PUBLISH]->(rootmgs)
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
