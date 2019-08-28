// Bootstrap YAIAM

// Depends on Identity model from IDP

// Brands
MERGE (:Brand {name:"yaiam"}) // Yet another Identity Access Managemet system
;

// Systems
MERGE (:System {name:"idp"}) // The identity provider
MERGE (:System {name:"aap"}) // The authorization access provder
MERGE (:System {name:"hydra"}) // The oauth2 delegator
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

// # Permission
// Register which permission an app exposes trough a Policy
// Description: facts (immutable axioms)
// TODO: Should we split permission into sub groups of functional vs. data permissions?

// ## IDP API
MERGE (:Permission {name:"openid"})
MERGE (:Permission {name:"offline"})
MERGE (:Permission {name:"authenticate:identity"})
MERGE (:Permission {name:"read:identity"})
MERGE (:Permission {name:"update:identity"})
MERGE (:Permission {name:"delete:identity"})
MERGE (:Permission {name:"recover:identity"})
MERGE (:Permission {name:"logout:identity"})
;

// Permission exposed by idpapi
MATCH (idpapi:Identity {sub:"idpapi"})
MATCH (p:Permission {name:"openid"})
MERGE (idpapi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (idpapi:Identity {sub:"idpapi"})
MATCH (p:Permission {name:"offline"})
MERGE (idpapi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (idpapi:Identity {sub:"idpapi"})
MATCH (p:Permission {name:"authenticate:identity"})
MERGE (idpapi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (idpapi:Identity {sub:"idpapi"})
MATCH (p:Permission {name:"read:identity"})
MERGE (idpapi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (idpapi:Identity {sub:"idpapi"})
MATCH (p:Permission {name:"update:identity"})
MERGE (idpapi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (idpapi:Identity {sub:"idpapi"})
MATCH (p:Permission {name:"delete:identity"})
MERGE (idpapi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (idpapi:Identity {sub:"idpapi"})
MATCH (p:Permission {name:"authenticate:identity"})
MERGE (idpapi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (idpapi:Identity {sub:"idpapi"})
MATCH (p:Permission {name:"recover:identity"})
MERGE (idpapi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

MATCH (idpapi:Identity {sub:"idpapi"})
MATCH (p:Permission {name:"logout:identity"})
MERGE (idpapi)-[:Exposes]->(o:Policy)-[:Grant]->(p)
;

// ## IDP UI

// ### Grant IDP UI permission to access IDP API
MATCH (idpui:Identity {sub:"idpui"})
MATCH (idpapi:Identity {sub:"idpapi"})
MATCH (root:Identity {sub:"root"})

WITH idpui, idpapi, root

// Find policies exposed by the app to granted scopes and revoked scopes.
OPTIONAL MATCH (idpapi)-[:Exposes]->(policyGrant:Policy)-[:Grant]->(grantedPermissions:Permission) WHERE grantedPermissions.name in split("openid offline authenticate:identity read:identity update:identity delete:identity recover:identity logout:identity", " ")

WITH idpui, idpapi, root, grantedPermissions, policyGrant, collect(policyGrant) as grantedPolicies

FOREACH ( policy in grantedPolicies |
  MERGE (root)<-[granted_by:GrantedBy]-(r:Rule)-[:Grant]->(policyGrant) ON CREATE SET granted_by.created_dtm = timestamp()
  MERGE (idpui)-[is_granted:IsGranted]->(r) ON CREATE SET is_granted.created_dtm = timestamp()
)
;
