// OBS: Schema changes cannot be run in same transaction as data queries

CREATE CONSTRAINT ON (s:Scope) ASSERT s.name IS UNIQUE;
