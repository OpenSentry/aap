// Bootstrap YAIAM

// Permission facts (immutable axioms)
// TODO: Should we split permission into sub groups of functional vs. data permissions?
CREATE CONSTRAINT ON (p:Permission) ASSERT p.name IS UNIQUE;
MERGE (:Permission {name:"openid"})
MERGE (:Permission {name:"offline"})
MERGE (:Permission {name:"email.read"})
;

// Brands
CREATE CONSTRAINT ON (b:Brands) ASSERT b.name IS UNIQUE;
MERGE (:Brand {name:"yaiam"}) // Yet another Identity Access Managemet system
;

// Systems
CREATE CONSTRAINT ON (s:System) ASSERT s.name IS UNIQUE;
MERGE (:System {name:"idp"}) // The identity provider
MERGE (:System {name:"aap"}) // The authorization access provder
MERGE (:System {name:"hydra"}) // The oauth2 delegator
;

// Apps
CREATE CONSTRAINT ON (a:App) ASSERT a.name IS UNIQUE;
MERGE (:App {name:"idpapi"}) // The idp api
MERGE (:App {name:"idpui"})  // The idp ui
MERGE (:App {name:"aapapi"}) // The aap api
MERGE (:App {name:"aapui"})  // The aap ui
MERGE (:App {name:"hydra"})  // The hydra
;

// Identity (pass: 123)
CREATE CONSTRAINT ON (i:Identity) ASSERT i.sub IS UNIQUE;
// Apps (client_id) (pass: 123), should probably be the client_secret
MERGE (:Identity {sub:"idpui", password:"$2a$10$SOyUCy0KLFQJa3xN90UgMe9q5wE.LfakmkCsfKLCIjRY6.CcRDYwu", name:"IdP UI"})
MERGE (:Identity {sub:"idpapi", password:"$2a$10$SOyUCy0KLFQJa3xN90UgMe9q5wE.LfakmkCsfKLCIjRY6.CcRDYwu", name:"IdP API"})
MERGE (:Identity {sub:"aapui", password:"$2a$10$SOyUCy0KLFQJa3xN90UgMe9q5wE.LfakmkCsfKLCIjRY6.CcRDYwu", name:"AaP UI"})
MERGE (:Identity {sub:"aapapi", password:"$2a$10$SOyUCy0KLFQJa3xN90UgMe9q5wE.LfakmkCsfKLCIjRY6.CcRDYwu", name:"AaP API"})
MERGE (:Identity {sub:"hydra", password:"$2a$10$SOyUCy0KLFQJa3xN90UgMe9q5wE.LfakmkCsfKLCIjRY6.CcRDYwu", name:"Hydra"})
// Humans (pass: 123)
MERGE (:Identity {sub:"wraix", password:"$2a$10$SOyUCy0KLFQJa3xN90UgMe9q5wE.LfakmkCsfKLCIjRY6.CcRDYwu", email:"wraix@domain.com", name:"Wraix"})
MERGE (:Identity {sub:"user-1", password:"$2a$10$SOyUCy0KLFQJa3xN90UgMe9q5wE.LfakmkCsfKLCIjRY6.CcRDYwu", email:"user-1@domain.com", name:"User 1"})
;

// Register relations between components
MATCH (yaiam:Brand {name:"yaiam"})
MATCH (idp:System {name:"idp"})
MATCH (aap:System {name:"aap"})
MATCH (hydra:System {name:"hydra"})
MATCH (appIdpApi:App {name:"idpapi"})
MATCH (appIdpUi:App {name:"idpui"})
MATCH (appAapApi:App {name:"aapapi"})
MATCH (appAapUi:App {name:"aapui"})
MATCH (appHydra:App {name:"hydra"})
MATCH (idIdpApi:Identity {sub:"idpapi"})
MATCH (idIdpUi:Identity {sub:"idpui"})
MATCH (idAapApi:Identity {sub:"aapapi"})
MATCH (idAapUi:Identity {sub:"aapui"})
MATCH (idHydra:Identity {sub:"hydra"})
MERGE (yaiam)-[:Manages]->(idp)-[:Manages]->(appIdpApi)-[:Manages]->(idIdpApi)
MERGE (idp)-[:Manages]->(appIdpUi)-[:Manages]->(idIdpUi)
MERGE (yaiam)-[:Manages]->(aap)-[:Manages]->(appAapApi)-[:Manages]->(idAapApi)
MERGE (aap)-[:Manages]->(appAapUi)-[:Manages]->(idAapUi)
MERGE (yaiam)-[:Manages]->(hydra)-[:Manages]->(appHydra)-[:Manages]->(idHydra)
;

// Register which permission an app exposes trough a Policy
MATCH (appIdpUi:App {name:"idpui"})
MATCH (p:Permission {name:"openid"})
MERGE (appIdpUi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (appIdpUi:App {name:"idpui"})
MATCH (p:Permission {name:"offline"})
MERGE (appIdpUi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (appIdpUi:App {name:"idpui"})
MATCH (p:Permission {name:"email.read"})
MERGE (appIdpUi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;
