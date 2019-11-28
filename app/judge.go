package app

import (
  "fmt"
  "strings"
  "github.com/neo4j/neo4j-go-driver/neo4j"
  "golang.org/x/oauth2"

  hydra "github.com/charmixer/hydra/client"

  "github.com/charmixer/aap/gateway/aap"
)

type Introspection struct {
  Subject aap.Identity
  Client aap.Identity
  Caller aap.Identity
}

type JudgeVerdict struct {
  Introspection Introspection
  Verdict aap.Verdict
  Reason string
}

func denyWithReason(msg string, introspection Introspection) (deny JudgeVerdict) {
  return JudgeVerdict{ Introspection:introspection, Reason:msg, Verdict:aap.Verdict{} }
}

func Judge(tx neo4j.Transaction, token *oauth2.Token, iPublisher aap.Identity, iScopes []aap.Scope, iOwners []aap.Identity, iCaller aap.Identity, hydraClient *hydra.HydraClient, introspectTokenUrl string) (judgeVerdict JudgeVerdict, err error) {

  isCallerSameAsRequestor := false
  if iCaller.Id == "" {
    isCallerSameAsRequestor = true
  }

  isOwnersSameAsRequestor := false
  if len(iOwners) <= 0 {
    isOwnersSameAsRequestor = true
  }

  var _s []string
  for _, iScope := range iScopes {
    if iScope.Name != "" {
       _s = append(_s, iScope.Name)
    }
  }
  scopes := strings.Join(_s, " ")

  // Perform introspection using the access token
  introspectRequest := hydra.IntrospectRequest{
    Token: token.AccessToken,
    Scope: scopes,
  }
  introspectResponse, err := hydra.IntrospectToken(introspectTokenUrl, hydraClient, introspectRequest)
  if err != nil {
    return JudgeVerdict{}, err
  }

  // Check scopes. (is done by hydra according to doc)
  // https://www.ory.sh/docs/hydra/sdk/api#introspect-oauth2-tokens
  if introspectResponse.Active == true {

    if introspectResponse.TokenType != "access_token" {
      return denyWithReason("Invalid token. Hint: Token is not an access_token", Introspection{}), nil
    }

    if introspectResponse.Sub == "" {
      return denyWithReason("Missing token sub", Introspection{}), nil
    }
    requestorId := introspectResponse.Sub

    if introspectResponse.ClientId == "" {
      return denyWithReason("Missing token client_id", Introspection{}), nil
    }
    clientId := introspectResponse.ClientId

    iClient := aap.Identity{Id:clientId}
    iRequestor := aap.Identity{Id:requestorId}

    if isCallerSameAsRequestor == true {
      iCaller = iRequestor
    }

    if isOwnersSameAsRequestor == true {
      iOwners = append(iOwners, iRequestor)
    }

    introspection := Introspection{ Subject: iRequestor, Client: iClient, Caller: iCaller }

    // No need to judge if there are no scopes in the request, then we only needed authentication of token from hydra (and to lookup a subject)
    if scopes == "" {
      // Token authenticated, return subject!
      verdictAuthenticated := aap.Verdict{
        Publisher: iPublisher,
        Requestor: iRequestor,
        RequestedScopes: iScopes,
        GrantedScopes: iScopes,
        MissingScopes: []aap.Scope{},
        Owners: iOwners,
        Granted: true,
      }
      return JudgeVerdict{ Introspection: introspection, Verdict: verdictAuthenticated }, nil
    }

    verdict, err := aap.Judge(tx, iPublisher, iRequestor, iScopes, iOwners)
    if err != nil {
      return denyWithReason("Server error occurred", introspection), err
    }

    if verdict.Granted == true {
      // Authorized!
      return JudgeVerdict{ Introspection: introspection, Verdict: verdict }, nil
    }

    var _missingScopes []string
    for _, scope := range verdict.MissingScopes {
       _missingScopes = append(_missingScopes, scope.Name)
    }

    return denyWithReason(fmt.Sprintf("Missing grants. Hint: Access token is missing required grants: %s", strings.Join(_missingScopes, " ")), introspection), nil
  }

  return denyWithReason(fmt.Sprintf("Missing required scopes. Hint: Access token is missing required oauth2 scopes: %s", scopes), Introspection{}), nil
}