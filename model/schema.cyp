// OBS: Schema changes cannot be run in same transaction as data queries

CREATE CONSTRAINT ON (s:Scope) ASSERT p.name IS UNIQUE;
