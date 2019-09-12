// OBS: Schema changes cannot be run in same transaction as data queries

CREATE CONSTRAINT ON (p:Permission) ASSERT p.name IS UNIQUE;
