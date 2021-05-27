package errors

import (
	bulky "github.com/charmixer/bulky/errors"
)

const INPUT_VALIDATION_FAILED = 1
const EMPTY_REQUEST_NOT_ALLOWED = 2
const MAX_REQUESTS_EXCEEDED = 3
const FAILED_DUE_TO_OTHER_ERRORS = 4
const INTERNAL_SERVER_ERROR = 5

const CONSENT_NOT_FOUND = 10
const NO_SUBSCRIPTIONS = 11
const INVALID_SCOPES = 12

func InitRestErrors() {
	bulky.AppendErrors(
		map[int]map[string]string{
			CONSENT_NOT_FOUND: {
				"en":  "Not found",
				"dev": "Consent not found",
			},
			NO_SUBSCRIPTIONS: {
				"en":  "No subscriptions",
				"dev": "No subscriptions. Hint: Client is missing subscription on any of the requested scopes and audiences.",
			},
			INVALID_SCOPES: {
				"en":  "Invalid scopes",
				"dev": "Invalid scopes. Hint: Atleast one requested scope is not subscribed for any audience.",
			},
		},
	)
}
