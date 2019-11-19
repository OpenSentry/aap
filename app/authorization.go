package app

import (
  "strings"
  "fmt"
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"
  "golang.org/x/oauth2"

  hydra "github.com/charmixer/hydra/client"

  "github.com/charmixer/aap/config"
  "github.com/charmixer/aap/gateway/aap"
)

func AuthorizationRequired(env *Environment, requiredScopes ...string) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{"func": "AuthorizationRequired"})

    // This is required to be here but should be garantueed by the authenticationRequired function.
    t, accessTokenExists := c.Get(env.Constants.AccessTokenKey)
    if accessTokenExists == false {
      c.AbortWithStatusJSON(http.StatusForbidden, JsonError{ErrorCode: ERROR_MISSING_BEARER_TOKEN, Error: "No access token found. Hint: Is bearer token missing?"})
      return
    }
    var accessToken *oauth2.Token = t.(*oauth2.Token)

    strRequiredScopes := strings.Join(requiredScopes, " ")
    log.WithFields(logrus.Fields{"scope": strRequiredScopes}).Debug("Checking required scopes");

    // See #3 of QTNA
    // log.WithFields(logrus.Fields{"fixme": 1, "qtna": 3}).Debug("Missing check if access token is granted the required scopes")
    hydraClient := hydra.NewHydraClient(env.OAuth2Delegator.Config)

    log.WithFields(logrus.Fields{"token": accessToken.AccessToken}).Debug("Introspecting token")

    introspectRequest := hydra.IntrospectRequest{
      Token: accessToken.AccessToken,
      Scope: strRequiredScopes, // This will make hydra check that all scopes are present else introspect.active will be false.
    }
    introspectResponse, err := hydra.IntrospectToken(env.OAuth2Delegator.IntrospectTokenUrl, hydraClient, introspectRequest)
    if err != nil {
      log.WithFields(logrus.Fields{"scope": strRequiredScopes}).Debug(err.Error())
      c.AbortWithStatus(http.StatusInternalServerError)
      return
    }

    var grantedScopes []string
    var missingScopes []string

    if introspectResponse.Active == true {

      if introspectResponse.TokenType != "access_token" {
        log.Debug("Token is not an access_token")
        c.AbortWithStatus(http.StatusForbidden)
        return
      }

      // Check scopes. (is done by hydra according to doc)
      // https://www.ory.sh/docs/hydra/sdk/api#introspect-oauth2-tokens

      // See #4 of QTNA
      sub := introspectResponse.Sub
      //client := introspectResponse.ClientId

      publisherEntity := aap.Identity{ Id:config.GetString("id") }
      requestorEntity := aap.Identity{ Id:sub }
      ownerEntity := aap.Identity{ Id:sub }

      session, tx, err := aap.BeginReadTx(env.Driver)
      if err != nil {
        log.Debug(err.Error())
        c.AbortWithStatus(http.StatusInternalServerError)
        return
      }
      defer tx.Close() // rolls back if not already committed/rolled back
      defer session.Close()

      for _, scope := range requiredScopes {

        verdict, err := aap.Judge(tx, publisherEntity, requestorEntity, aap.Scope{ Name:scope }, []aap.Identity{ ownerEntity })
        if err != nil {
          log.WithFields(logrus.Fields{ "publisher":publisherEntity.Id, "requestor":requestorEntity.Id, "owner":ownerEntity.Id, "scope":scope }).Debug(err.Error())
          c.AbortWithStatus(http.StatusInternalServerError)
          return
        }

        if verdict.Granted == true {
          grantedScopes = append(grantedScopes, scope)
        } else {
          missingScopes = append(missingScopes, scope)
        }

      }

      if len(missingScopes) == 0 && len(grantedScopes) == len(requiredScopes) {
        log.WithFields(logrus.Fields{"sub": introspectResponse.Sub, "scope": strRequiredScopes}).Debug("Authorized")
        c.Set("sub", introspectResponse.Sub)
        c.Next() // Authentication successful, continue.
        return
      }
    }

    // Deny by default
    var strMissingScopes string = strings.Join(missingScopes, " ")
    log.WithFields(logrus.Fields{"scope": strMissingScopes}).Debug("Missing Scopes");
    c.AbortWithStatusJSON(http.StatusForbidden, JsonError{ErrorCode: ERROR_MISSING_REQUIRED_SCOPES, Error: fmt.Sprintf("Missing required scopes: %s. Hint: Some required scopes are missing, invalid or not granted", strMissingScopes)})
    return

  }
  return gin.HandlerFunc(fn)
}