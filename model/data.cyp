// # AAP

// ## Requirement: Depends on Identity from IDP

// ### Scopes, IDP

// idp
MERGE (:Scope {name:"openid", title:"Login to the IDP system", description:"Allows access to login to the IDP system"})
MERGE (:Scope {name:"offline", title:"Remember me", description:"Allows access to remember your login session over a longer period of time"})

MERGE (:Scope {name:"idp:read:identities", title:"Read your identity", description:"Allows access to your profile information, such as email, name and profile picture"})

MERGE (:Scope {name:"idp:create:invites", title:"Create an Invite", description:"Allow creation of invites to the IDP system"})
MERGE (:Scope {name:"idp:read:invites", title:"Read an Invite", description:"Allow reading invites in the IDP system"})
MERGE (:Scope {name:"idp:send:invites", title:"Send an Invite", description:"Allow sending out invites from the IDP system"})
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

MERGE (:Scope {name:"idp:create:humans:emailchange", title:"Change Email", description:"Allow access to change email"})
MERGE (:Scope {name:"idp:update:humans:emailchange", title:"Update Email", description:"Allow access to update email"})

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

MERGE (:Scope {name:"idp:create:roles", title:"Create roles", description:"Allow access to create roles"})
MERGE (:Scope {name:"idp:read:roles", title:"Read roles", description:"Allow access to read roles"})
MERGE (:Scope {name:"idp:delete:roles", title:"Delete roles", description:"Allow access to delete roles"})




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
MERGE (:Scope {name:"aap:read:consents:authorize", title:"Read consent challenge", description:"Allow read consent challenge"})
MERGE (:Scope {name:"aap:create:consents:authorize", title:"Accept consent challenge", description:"Allow consenting to access to entity onbehalf of entity"})
MERGE (:Scope {name:"aap:create:consents:reject", title:"Reject consent to entity", description:"Allow rejecting access to entity on behalf of entity"})
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


// ## IDP (ResourceServer) grants to client used to call AAP
MATCH (client:Identity:Client {id:"8dc7ea3e-c61a-47cd-acf2-2f03615e3f8b"})
MATCH (rs:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope) where s.name in split("aap:read:entities:judge aap:create:entities aap:create:grants", " ")
MATCH (rs)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
MERGE (client)-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr)
MERGE (gr)-[:ON_BEHALF_OF]->(rs)
;

// ## IDP UI (Application) grants to required scopes which relates to credentials like password, otp codes etc. (Secret Grants)
MATCH (client:Identity:Client {id:"c7f1afc4-1e1f-484e-b3c2-0519419690cb"})
MATCH (rs:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope) where s.name in split("idp:create:humans:authenticate idp:read:humans idp:read:invites idp:create:invites idp:claim:invites idp:update:challenges:verify idp:read:challenges idp:create:humans idp:read:humans:logout idp:update:humans:logout idp:update:humans:deleteverification idp:create:humans:recover idp:update:humans:recoververification", " ")
MATCH (rs)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
MERGE (client)-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr)
MERGE (gr)-[:ON_BEHALF_OF]->(rs)
;

// ## IDP UI subscribes openid, offline
MATCH (subscriber:Identity:Client {id:"c7f1afc4-1e1f-484e-b3c2-0519419690cb"})
MATCH (publisher:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope) where s.name in split("openid offline", " ")
MATCH (publisher)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
MERGE (subscriber)-[:SUBSCRIBES]-(sr:Subscribe:Rule)-[:SUBSCRIBES]->(pr)
;

// ## IDP UI subscribes to IDP
MATCH (subscriber:Identity:Client {id:"c7f1afc4-1e1f-484e-b3c2-0519419690cb"})
MATCH (publisher:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope) where s.name in split("idp:read:identities idp:create:invites idp:read:invites idp:send:invites idp:claim:invites idp:read:humans idp:read:humans:logout idp:create:humans:logout idp:update:humans:logout idp:delete:humans idp:create:humans:recover idp:create:humans idp:create:humans:authenticate idp:update:humans:recoververification idp:update:humans:deleteverification idp:read:challenges idp:create:challenges idp:update:challenges:verify idp:update:humans:totp idp:update:humans:password idp:create:humans:emailchange", " ")
MATCH (publisher)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
MERGE (subscriber)-[:SUBSCRIBES]-(sr:Subscribe:Rule)-[:SUBSCRIBES]->(pr)
;




// ## AAP UI (Application) grants to required scopes which relates to consents (Consent Grants)
MATCH (client:Identity:Client {id:"919e2026-06af-4c82-9d84-6af4979d9e7a"})
MATCH (rs:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope) where s.name in split("aap:create:consents:authorize aap:read:consents:authorize aap:create:consents:reject aap:read:consents aap:create:consents aap:read:subscriptions aap:read:publishes", " ")
MATCH (rs)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
MERGE (client)-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr)
MERGE (gr)-[:ON_BEHALF_OF]->(rs)
;



// ## ME UI subscribes to OPENID, OFFLINE (skal den det her?)
MATCH (subscriber:Identity:Client {id:"20f2bfc6-44df-424a-b490-c024d009892c"})
MATCH (publisher:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope) where s.name in split("openid offline", " ")
MATCH (publisher)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
MERGE (subscriber)-[:SUBSCRIBES]-(sr:Subscribe:Rule)-[:SUBSCRIBES]->(pr)
;

// ## ME UI subscribes to IDP
MATCH (subscriber:Identity:Client {id:"20f2bfc6-44df-424a-b490-c024d009892c"})
MATCH (publisher:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope) where s.name in split("idp:read:identities idp:read:humans idp:update:humans idp:create:humans:recover idp:create:invites idp:read:invites idp:send:invites idp:claim:invites idp:create:resourceservers idp:read:resourceservers idp:delete:resourceservers idp:create:clients idp:read:clients idp:delete:clients idp:create:roles idp:read:roles idp:delete:roles", " ")
MATCH (publisher)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
MERGE (subscriber)-[:SUBSCRIBES]-(sr:Subscribe:Rule)-[:SUBSCRIBES]->(pr)
;

// ## ME UI subscribes to AAP
MATCH (subscriber:Identity:Client {id:"20f2bfc6-44df-424a-b490-c024d009892c"})
MATCH (publisher:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope) where s.name in split("aap:read:scopes aap:create:scopes aap:update:scopes aap:read:grants aap:create:grants aap:delete:grants aap:read:publishes aap:create:publishes aap:delete:publishes aap:read:consents aap:delete:consents aap:create:subscriptions aap:delete:subscriptions aap:read:subscriptions", " ")
MATCH (publisher)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
MERGE (subscriber)-[:SUBSCRIBES]-(sr:Subscribe:Rule)-[:SUBSCRIBES]->(pr)
;
