package app

import (
  "strings"
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"
  "golang.org/x/oauth2"
)

func AuthenticationRequired(logKey string, accessTokenKey string) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(logKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "AuthenticationRequired",
    })

    log = log.WithFields(logrus.Fields{"authorization": "bearer"})
    log.Debug("Looking for access token")
    var token *oauth2.Token
    auth := c.Request.Header.Get("Authorization")
    split := strings.SplitN(auth, " ", 2)
    if len(split) == 2 || strings.EqualFold(split[0], "bearer") {

      token = &oauth2.Token{
        AccessToken: split[1],
        TokenType: split[0],
      }

      // See #2 of QTNA
      // https://godoc.org/golang.org/x/oauth2#Token.Valid
      if token.Valid() == true {

        // See #5 of QTNA
        log.WithFields(logrus.Fields{"fixme": 1, "qtna": 5}).Debug("Missing check against token-revoked-list to check if token is revoked")

        log.Debug("Authenticated")
        c.Set(accessTokenKey, token)
        c.Next() // Authentication successful, continue.
        return;
      }

      // Deny by default
      c.JSON(http.StatusUnauthorized, JsonError{ErrorCode: ERROR_INVALID_ACCESS_TOKEN, Error: "Invalid access token."})
      c.Abort()
      return
    }

    // Deny by default
    c.JSON(http.StatusUnauthorized, JsonError{ErrorCode: ERROR_MISSING_BEARER_TOKEN, Error: "Authorization: Bearer <token> not found in request"})
    c.Abort()
  }
  return gin.HandlerFunc(fn)
}

func AccessToken(env *Environment, c *gin.Context) (*oauth2.Token) {
  t, exists := c.Get(env.Constants.AccessTokenKey)
  if exists == true {
    return t.(*oauth2.Token)
  }
  return nil
}