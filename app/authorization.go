package app

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"

	hydra "github.com/charmixer/hydra/client"

	"github.com/opensentry/aap/config"
	"github.com/opensentry/aap/gateway/aap"
)

func AuthorizationRequired(env *Environment, requiredScopes ...string) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
		log = log.WithFields(logrus.Fields{"func": "AuthorizationRequired"})

		strRequiredScopes := strings.Join(requiredScopes, " ")
		log = log.WithFields(logrus.Fields{"scopes.required": strRequiredScopes})

		accessToken := AccessToken(env, c)
		if accessToken == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, JsonError{ErrorCode: ERROR_MISSING_BEARER_TOKEN, Error: "No access token found. Hint: Is bearer token missing?"})
			return
		}

		iCaller := aap.Identity{} // Caller is the access token, meaning requestor. This will make judge set iCaller = iRequestor after introspection succeed.
		iPublisher := aap.Identity{Id: config.GetString("id")}
		iOwners := []aap.Identity{} // Owner is the access token, meaning requestor. This will make judge set iCaller = iRequestor after introspection succeed.
		/*for _, id := range r.Owners {
		  iOwners = append(iOwners, aap.Identity{Id:id})
		}*/

		var iScopes []aap.Scope
		for _, scope := range requiredScopes {
			iScopes = append(iScopes, aap.Scope{Name: scope})
		}

		hydraClient := hydra.NewHydraClient(env.OAuth2Delegator.Config)

		session, tx, err := aap.BeginReadTx(env.Driver)
		if err != nil {
			log.Debug(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer tx.Close() // rolls back if not already committed/rolled back
		defer session.Close()

		judgeVerdict, err := Judge(tx, accessToken, iPublisher, iScopes, iOwners, iCaller, hydraClient, env.OAuth2Delegator.IntrospectTokenUrl)
		if err != nil {
			log.Debug(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if judgeVerdict.Verdict.Granted == true {
			sub := judgeVerdict.Introspection.Subject.Id

			log = log.WithFields(logrus.Fields{"sub": sub})
			log.Debug("Authorized")

			c.Set("sub", sub)
			c.Next() // Authentication successful, continue.
			return
		}

		// Deny by default
		log.Debug("Unauthorized")
		c.AbortWithStatusJSON(http.StatusForbidden, JsonError{ErrorCode: ERROR_MISSING_REQUIRED_SCOPES, Error: judgeVerdict.Reason})
		return

	}
	return gin.HandlerFunc(fn)
}
