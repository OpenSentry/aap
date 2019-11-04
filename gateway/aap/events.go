package aap

import (
  nats "github.com/nats-io/nats.go"
  "fmt"
)

func EmitEventConsentCreated(natsConnection *nats.Conn, consent Consent) {
  e := fmt.Sprintf("{sub:%s, client_id:%s, aud:%s, scope:%s}", consent.Identity.Id, consent.Subscriber.Id, consent.Publisher.Id, consent.Scope.Name)
  natsConnection.Publish("aap.consent.created", []byte(e))
}
