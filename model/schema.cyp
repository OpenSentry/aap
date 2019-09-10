// OBS: Schema changes cannot be run in same transaction as data queries

CREATE CONSTRAINT ON (s:Scope) ASSERT s.name IS UNIQUE;
CREATE CONSTRAINT ON (c:Client) ASSERT c.name IS UNIQUE;
CREATE CONSTRAINT ON (c:Client) ASSERT c.client_id IS UNIQUE;
CREATE CONSTRAINT ON (a:ResourceServer) ASSERT a.name IS UNIQUE;
CREATE CONSTRAINT ON (a:ResourceServer) ASSERT a.aud IS UNIQUE;
