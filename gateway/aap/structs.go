package aap

import (
  "github.com/neo4j/neo4j-go-driver/neo4j"
)

type Identity struct {
  Id       string
}
func marshalNodeToIdentity(node neo4j.Node) (Identity) {
  p := node.Props()

  return Identity{
    Id:       p["id"].(string),
  }
}

type Human struct {
  Id        string
  Password  string
  Name      string
  Email     string
}
func marshalNodeToHuman(node neo4j.Node) (Human) {
  p := node.Props()

  return Human{
    Id:       p["id"].(string),
    Password: p["password"].(string),
    Name:     p["name"].(string),
    Email:    p["email"].(string),
  }
}

type Client struct {
  ClientId     string
  ClientSecret string
  Name         string
  Description  string
}
func marshalNodeToClient(node neo4j.Node) (Client) {
  p := node.Props()

  return Client{
    ClientId:     p["client_id"].(string),
    ClientSecret: p["client_secret"].(string),
    Name:         p["name"].(string),
    Description:  p["description"].(string),
  }
}

type ResourceServer struct {
  Name        string
  Audience    string
  Description string
}
func marshalNodeToResourceServer(node neo4j.Node) (ResourceServer) {
  p := node.Props()

  return ResourceServer{
    Name:        p["name"].(string),
    Audience:    p["aud"].(string),
    Description: p["description"].(string),
  }
}

type Scope struct {
  Name        string
}
func marshalNodeToScope(node neo4j.Node) (Scope) {
  p := node.Props()

  return Scope{
    Name:        p["name"].(string),
  }
}

type PublishRule struct {
  Title       string
  Description string
}
func marshalNodeToPublishRule(node neo4j.Node) (PublishRule) {
  p := node.Props()

  return PublishRule{
    Title:       p["title"].(string),
    Description: p["description"].(string),
  }
}
type Grant struct {
  Identity Identity
  Scope Scope
  Publisher Identity
}

type Consent struct {
  Identity
  Client
  ResourceServer
  Scope
}

type Publish struct {
  Publisher Identity
  Scope Scope
  MayGrantScope Scope
  Rule PublishRule
  MayGrantRule PublishRule
}
