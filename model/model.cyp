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

MATCH (idpapi:Identity {sub:"idpapi"})
MATCH (idpui:Identity {sub:"idpui"})
MATCH (aapapi:Identity {sub:"aapapi"})
MATCH (aapui:Identity {sub:"aapui"})
MATCH (ihydra:Identity {sub:"hydra"})
MERGE (yaiam)-[:Manages]->(idp)-[:Manages]->(idpapi)
MERGE (idp)-[:Manages]->(idpui)
MERGE (yaiam)-[:Manages]->(aap)-[:Manages]->(aapapi)
MERGE (aap)-[:Manages]->(aapui)
MERGE (yaiam)-[:Manages]->(hydra)-[:Manages]->(ihydra)
;

// Register which permission an app exposes trough a Policy
MATCH (idpui:Identity {sub:"idpui"})
MATCH (p:Permission {name:"openid"})
MERGE (idpui)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (idpui:Identity {sub:"idpui"})
MATCH (p:Permission {name:"offline"})
MERGE (idpui)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (idpui:Identity {sub:"idpui"})
MATCH (p:Permission {name:"email.read"})
MERGE (idpui)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;
