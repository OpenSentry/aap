// OBS: Schema changes cannot be run in same transaction as data queries

CREATE CONSTRAINT ON (p:Permission) ASSERT p.name IS UNIQUE;
CREATE CONSTRAINT ON (b:Brands) ASSERT b.name IS UNIQUE;
CREATE CONSTRAINT ON (s:System) ASSERT s.name IS UNIQUE;
CREATE CONSTRAINT ON (i:Identity) ASSERT i.sub IS UNIQUE;
