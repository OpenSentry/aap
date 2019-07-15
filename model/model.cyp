// Permission facts (immutable axioms)
// TODO: Should we split permission into sub groups of functional vs. data permissions?
CREATE CONSTRAINT ON (p:Permission) ASSERT p.name IS UNIQUE;
MERGE (:Permission {name:"openid"})

// Brands
CREATE CONSTRAINT ON (b:Brands) ASSERT b.name IS UNIQUE;
MERGE (:Brand {name:"YAIDP"}) // Yet another Identity Provider

// Systems
CREATE CONSTRAINT ON (s:System) ASSERT s.name IS UNIQUE;
MERGE (:System {name:"YAIAM"}) // Yet another IAM

// Apps
CREATE CONSTRAINT ON (a:App) ASSERT a.name IS UNIQUE;
MERGE (:App {name:"Idp"})
MERGE (:App {name:"Cp"})

// Identity (pass: 123)
CREATE CONSTRAINT ON (i:Identity) ASSERT i.sub IS UNIQUE
MERGE (:Identity {sub:"wraix", password:"$2a$10$SOyUCy0KLFQJa3xN90UgMe9q5wE.LfakmkCsfKLCIjRY6.CcRDYwu", email:"wraix@domain.com", name:"Wraix"})
MERGE (:Identity {sub:"user-1", password:"$2a$10$SOyUCy0KLFQJa3xN90UgMe9q5wE.LfakmkCsfKLCIjRY6.CcRDYwu", email:"user-1@domain.com", name:"User 1"})

// Register which brand manages system and which system manages app
MATCH (b:Brand {name:"YAIDP"})
MATCH (s:System {name:"YAIAM"})
MATCH (a:App {name:"Idp"})
MERGE (b)-[:Manages]->(s)
MERGE (s)-[:Manages]->(a)

// Register which permission an app exposes trough a Policy
MATCH (a:App {name:"Idp"})
MATCH (p:Permission {name:"openid"})
MERGE (a)-[:Exposes]->(o:Policy)-[:Grant]->(p)

// Assign permissions to identities trough rules which grant policies and tracks who granted it
MATCH (i:Identity {sub:"wraix"})
MATCH (:App {name:"Idp"})-[:Exposes]->(o:Policy)-[:Grant]->(p:Permission {name:"openid"})
MERGE (i)-[:IsGranted]->(r:Rule)-[:Grant]->(o)
MERGE (r)-[:GrantedBy]->(i)
