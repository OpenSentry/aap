// OBS: Schema changes cannot be run in same transaction as data queries

CREATE CONSTRAINT ON (p:Permission) ASSERT p.name IS UNIQUE;
CREATE CONSTRAINT ON (c:Client) ASSERT c.client_id IS UNIQUE;
