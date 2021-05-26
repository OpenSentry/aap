package client

import (
	bulky "github.com/charmixer/bulky/client"
)

type Subscription struct {
	Subscriber string `json:"subscriber_id" validate:"required,uuid"`
	Publisher  string `json:"publisher_id" validate:"required,uuid"`
	Scope      string `json:"scope" validate:"required"`
}

type CreateSubscriptionsResponse Subscription
type CreateSubscriptionsRequest struct {
	Subscriber string `json:"subscriber_id" validate:"required,uuid"`
	Publisher  string `json:"publisher_id" validate:"required,uuid"`
	Scope      string `json:"scope" validate:"required"`
}

type DeleteSubscriptionsResponse struct{}
type DeleteSubscriptionsRequest struct {
	Subscriber string `json:"subscriber_id" validate:"required,uuid"`
	Publisher  string `json:"publisher_id" validate:"required,uuid"`
	Scope      string `json:"scope" validate:"required"`
}

type ReadSubscriptionsResponse []Subscription
type ReadSubscriptionsRequest struct {
	Subscriber string   `json:"subscriber_id" validate:"required,uuid"`
	Publisher  string   `json:"publisher_id,omitempty" validate:"omitempty,uuid"`
	Scopes     []string `json:"scopes" validate:"omitempty"`
}

func CreateSubscriptions(client *AapClient, url string, requests []CreateSubscriptionsRequest) (status int, responses bulky.Responses, err error) {
	status, err = handleRequest(client, requests, "POST", url, &responses)

	if err != nil {
		return status, nil, err
	}

	return status, responses, nil
}

func DeleteSubscriptions(client *AapClient, url string, requests []DeleteSubscriptionsRequest) (status int, responses bulky.Responses, err error) {
	status, err = handleRequest(client, requests, "DELETE", url, &responses)

	if err != nil {
		return status, nil, err
	}

	return status, responses, nil
}

func ReadSubscriptions(client *AapClient, url string, requests []ReadSubscriptionsRequest) (status int, responses bulky.Responses, err error) {
	status, err = handleRequest(client, requests, "GET", url, &responses)

	if err != nil {
		return status, nil, err
	}

	return status, responses, nil
}
