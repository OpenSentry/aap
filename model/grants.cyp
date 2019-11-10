// # Grants required for IDP, AAP system to work

// ## IDP (ResourceServer) self grants to published scopes (Not needed RS has root access per default)
// MATCH (rs:Identity:ResourceServer {name:"IDP"})
// MATCH (rs)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s:Scope)
// MERGE (rs)-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr)
// MERGE (gr)-[:ON_BEHALF_OF]->(rs)

// ## IDP (ResourceServer) grants to client used to call Hydra
// Not implemented

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
MATCH (s:Scope) where s.name in split("idp:create:humans:authenticate idp:read:humans idp:read:invites idp:create:invites idp:claim:invites idp:update:challenges:verify idp:read:challenges idp:create:humans idp:update:humans:totp", " ")

MATCH (rs)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)

MERGE (client)-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr)
MERGE (gr)-[:ON_BEHALF_OF]->(rs)
;

// ## Human (Identity) required grants (Subject Grants)
// MATCH (i:Identity:Human)
// MATCH (rs:Identity:ResourceServer {name:"IDP"})
// MATCH (s:Scope) where s.name in split("idp:read:humans idp:create:humans:logout idp:read:humans:logout idp:update:humans:logout", " ")
//
// MATCH (rs)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)
//
// MERGE (i)-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr)
// MERGE (gr)-[:ON_BEHALF_OF]->(i)
// ;


// # Grants required for AAP to work

// ## AAP (ResourceServer) grants to call Hydra
// Not implemented

// ## AAP UI (Application) grants to required scopes which relates to consents (Consent Grants)
MATCH (client:Identity:Client {id:"919e2026-06af-4c82-9d84-6af4979d9e7a"})
MATCH (rs:Identity:ResourceServer {name:"AAP"})
MATCH (s:Scope) where s.name in split("aap:create:consents:authorize aap:read:consents aap:create:consents", " ")

MATCH (rs)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)

MERGE (client)-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr)
MERGE (gr)-[:ON_BEHALF_OF]->(rs)
;


// Grants required for MEUI to work
// # MEUI (Application) grants required to work
// NONE - everything should be client credentials flow, meaning owners should have required grants to the meui requested scopes. (so see Human grants)


// User grants handy when missing bootstraps
MATCH (h:Human:Identity {id:"d6ef56ff-5422-42f5-8730-6c1aa6b736fc"})
MATCH (rs:Identity:ResourceServer {name:"IDP"})
MATCH (s:Scope) where s.name in split("idp:read:invites idp:create:invites idp:send:invites", " ")

MATCH (rs)-[:PUBLISH]->(pr:Publish:Rule)-[:PUBLISH]->(s)

MERGE (h)-[:IS_GRANTED]->(gr:Grant:Rule)-[:GRANTS]->(pr)
MERGE (gr)-[:ON_BEHALF_OF]->(rs)
;