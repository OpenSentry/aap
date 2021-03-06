package publishings

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/opensentry/aap/app"
	"github.com/opensentry/aap/client"
	"github.com/opensentry/aap/gateway/aap"

	bulky "github.com/charmixer/bulky/server"
)

func PostPublishes(env *app.Environment) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
		log = log.WithFields(logrus.Fields{
			"func": "PostPublishes",
		})

		var requests []client.CreatePublishesRequest
		err := c.BindJSON(&requests)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var handleRequests = func(iRequests []*bulky.Request) {

			session, tx, err := aap.BeginWriteTx(env.Driver)
			if err != nil {
				bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
				log.Debug(err.Error())
				return
			}
			defer tx.Close() // rolls back if not already committed/rolled back
			defer session.Close()

			requestor := c.MustGet("sub").(string)

			for _, request := range iRequests {
				r := request.Input.(client.CreatePublishesRequest)

				newPublish := aap.Publish{
					Publisher: aap.Identity{Id: r.Publisher},
					Scope:     aap.Scope{Name: r.Scope},
					Rule: aap.PublishRule{
						Title:       r.Title,
						Description: r.Description,
					},
				}
				db, err := aap.CreatePublishes(tx, aap.Identity{Id: requestor}, newPublish)
				if err != nil {
					e := tx.Rollback()
					if e != nil {
						log.Debug(e.Error())
					}
					bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
					request.Output = bulky.NewInternalErrorResponse(request.Index)
					log.Debug(err.Error())
					return
				}

				if db.Rule.Title != "" {
					var mgs = []string{}
					for _, e := range db.MayGrantScopes {
						mgs = append(mgs, e.Name)
					}

					ok := client.CreatePublishesResponse{
						Publisher:      db.Publisher.Id,
						Scope:          db.Scope.Name,
						Title:          db.Rule.Title,
						Description:    db.Rule.Description,
						MayGrantScopes: mgs,
					}
					request.Output = bulky.NewOkResponse(request.Index, ok)
					continue
				}

				// Deny by default
				e := tx.Rollback()
				if e != nil {
					log.Debug(e.Error())
				}
				bulky.FailAllRequestsWithServerOperationAbortedResponse(iRequests) // Fail all with abort
				request.Output = bulky.NewInternalErrorResponse(request.Index)     // Specify error on failed one
				return
			}

			err = bulky.OutputValidateRequests(iRequests)
			if err == nil {
				tx.Commit()
				return
			}

			// Deny by default
			tx.Rollback()
		}

		responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{MaxRequests: 1})
		c.JSON(http.StatusOK, responses)
	}
	return gin.HandlerFunc(fn)
}

func DeletePublishes(env *app.Environment) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
		log = log.WithFields(logrus.Fields{
			"func": "DeletePublishes",
		})

		c.AbortWithStatusJSON(http.StatusOK, gin.H{})
	}
	return gin.HandlerFunc(fn)
}

func GetPublishes(env *app.Environment) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		log := c.MustGet(env.Constants.LogKey).(*logrus.Entry)
		log = log.WithFields(logrus.Fields{
			"func": "GetPublishes",
		})

		var requests []client.ReadPublishesRequest
		err := c.BindJSON(&requests)
		if err != nil {
			log.Debug(err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var handleRequests = func(iRequests []*bulky.Request) {
			session, tx, err := aap.BeginReadTx(env.Driver)

			if err != nil {
				bulky.FailAllRequestsWithInternalErrorResponse(iRequests)
				log.Debug(err.Error())
				return
			}

			defer tx.Close() // rolls back if not already committed/rolled back
			defer session.Close()

			for _, request := range iRequests {
				var r client.ReadPublishesRequest
				if request.Input != nil {
					r = request.Input.(client.ReadPublishesRequest)
				}

				var iFilterPublisher aap.Identity
				if r.Publisher != "" {
					iFilterPublisher = aap.Identity{
						Id: r.Publisher,
					}
				}

				var iFilterScopes []aap.Scope
				if r.Scopes != nil {
					for _, scopeName := range r.Scopes {
						iFilterScopes = append(iFilterScopes, aap.Scope{
							Name: scopeName,
						})
					}
				}

				dbPublishes, _ := aap.FetchPublishes(tx, iFilterPublisher, iFilterScopes)

				var ok client.ReadPublishesResponse
				for _, db := range dbPublishes {
					var mgs []string
					for _, e := range db.MayGrantScopes {
						mgs = append(mgs, e.Name)
					}

					ok = append(ok, client.Publish{
						Publisher:      db.Publisher.Id,
						Scope:          db.Scope.Name,
						Title:          db.Rule.Title,
						Description:    db.Rule.Description,
						MayGrantScopes: mgs,
					})
				}

				request.Output = bulky.NewOkResponse(request.Index, ok)
			}
		}

		responses := bulky.HandleRequest(requests, handleRequests, bulky.HandleRequestParams{EnableEmptyRequest: true})

		c.JSON(http.StatusOK, responses)
	}
	return gin.HandlerFunc(fn)
}
