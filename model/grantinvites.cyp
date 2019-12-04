// Grant invites functionality to a subject (copied from gateway/grants.go)

MATCH (receiver:Identity {id:"?"})
MATCH (publisher:Identity {id: "e044d683-5daf-42af-a31a-938094611be9"})
MATCH (obo:Identity {id:"?"})
MATCH (scope:Scope {name: "idp:create:invites"})
MATCH (publisher)-[:PUBLISH]->(publishRule:Publish:Rule)-[:PUBLISH]->(scope)

OPTIONAL MATCH (receiver)-[:IS_GRANTED]->(existingGrantRule)-[:GRANTS]->(publishRule)
WHERE (existingGrantRule)-[:ON_BEHALF_OF]->(obo)

DETACH DELETE existingGrantRule

//WITH receiver, scope, publisher, publishRule, obo, existingGrantRule
//WHERE existingGrantRule is null

// ensure unique rules
CREATE (grantRule:Grant:Rule {nbf:0, exp:0})

// create scope and match it to the identity who created it
MERGE (receiver)-[:IS_GRANTED]->(grantRule)-[:GRANTS]->(publishRule)
MERGE (grantRule)-[:ON_BEHALF_OF]->(obo)
;

MATCH (receiver:Identity {id:"?"})
MATCH (publisher:Identity {id: "e044d683-5daf-42af-a31a-938094611be9"})
MATCH (obo:Identity {id:"?"})
MATCH (scope:Scope {name: "idp:read:invites"})
MATCH (publisher)-[:PUBLISH]->(publishRule:Publish:Rule)-[:PUBLISH]->(scope)

OPTIONAL MATCH (receiver)-[:IS_GRANTED]->(existingGrantRule)-[:GRANTS]->(publishRule)
WHERE (existingGrantRule)-[:ON_BEHALF_OF]->(obo)

DETACH DELETE existingGrantRule

//WITH receiver, scope, publisher, publishRule, obo, existingGrantRule
//WHERE existingGrantRule is null

// ensure unique rules
CREATE (grantRule:Grant:Rule {nbf:0, exp:0})

// create scope and match it to the identity who created it
MERGE (receiver)-[:IS_GRANTED]->(grantRule)-[:GRANTS]->(publishRule)
MERGE (grantRule)-[:ON_BEHALF_OF]->(obo)
;

MATCH (receiver:Identity {id:"?"})
MATCH (publisher:Identity {id: "e044d683-5daf-42af-a31a-938094611be9"})
MATCH (obo:Identity {id:"?"})
MATCH (scope:Scope {name: "idp:create:invites:send"})
MATCH (publisher)-[:PUBLISH]->(publishRule:Publish:Rule)-[:PUBLISH]->(scope)

OPTIONAL MATCH (receiver)-[:IS_GRANTED]->(existingGrantRule)-[:GRANTS]->(publishRule)
WHERE (existingGrantRule)-[:ON_BEHALF_OF]->(obo)

DETACH DELETE existingGrantRule

//WITH receiver, scope, publisher, publishRule, obo, existingGrantRule
//WHERE existingGrantRule is null

// ensure unique rules
CREATE (grantRule:Grant:Rule {nbf:0, exp:0})

// create scope and match it to the identity who created it
MERGE (receiver)-[:IS_GRANTED]->(grantRule)-[:GRANTS]->(publishRule)
MERGE (grantRule)-[:ON_BEHALF_OF]->(obo)
;