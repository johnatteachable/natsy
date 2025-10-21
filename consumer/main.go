package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

var (
	pingEventSubjects   = "ping.*.status"
	priceCreatedSubject = "price.created"
)

type (
	EventEnvelope struct {
		ID        uuid.UUID       `json:"id"`
		Type      string          `json:"type"`
		Version   int64           `json:"version"`
		Timestamp time.Time       `json:"timestamp"`
		Data      json.RawMessage `json:"data"`
	}

	PriceCreatedEvent struct {
		StbxProductID uuid.UUID `json:"stbx_product_id"`
		StbxPriceID   uuid.UUID `json:"stbx_price_id"`
		Name          string    `json:"name"`
		Amount        int       `json:"amount"`
		Currency      string    `json:"currency"`
	}

	PingEvent struct {
		Status string `json:"status"`
	}

	DBStore struct {
		Data map[uuid.UUID]string
	}
)

func main() {
	log.Print("Starting Consumer...")

	natsURI := os.Getenv("NATS_URI")
	log.Printf("Attempting to connect to NATS at %s", natsURI)
	nc, err := nats.Connect(natsURI)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	log.Print("Succesfully connected to NATS")

	//initialize "DB"
	dataMap := make(map[uuid.UUID]string, 1000)
	db := &DBStore{Data: dataMap}

	log.Printf("Subscribing to NATS subject:%s", pingEventSubjects)
	_, err = nc.Subscribe(pingEventSubjects, func(msg *nats.Msg) {
		var event EventEnvelope
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Error unmarshaling struct %+v, message data: %v", msg.Data, err)
			return
		}
		log.Printf("Received event:%s on subject %s", event.ID.String(), msg.Subject)
		err := event.Process(db)
		if err != nil {
			log.Printf("Error processing nats event data: %v", err)
			return
		}
	})
	if err != nil {
		log.Fatalf("Error subscribing to subject %s: %v", pingEventSubjects, err)
	}

	//Subscribing to product.created
	log.Printf("Subscribing to NATS subject:%s", priceCreatedSubject)
	_, err = nc.Subscribe(priceCreatedSubject, func(msg *nats.Msg) {
		var event EventEnvelope
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Error unmarshaling struct %+v, message data: %v", msg.Data, err)
			return
		}
		log.Printf("Received event:%s on subject %s", event.ID.String(), msg.Subject)
		err := event.Process(db)
		if err != nil {
			log.Printf("Error processing nats event data: %v", err)
			return
		}
	})
	if err != nil {
		log.Fatalf("Error subscribing to subject %s: %v", priceCreatedSubject, err)
	}

	// Keep the connection alive
	select {}
}

func (ee *EventEnvelope) Process(db *DBStore) error {
	switch ee.Type {
	case priceCreatedSubject:
		//Can add versioning here by parsing ee.Version
		log.Printf("Parsing version %d of Product Created Event", ee.Version)
		var productEvent PriceCreatedEvent
		if err := json.Unmarshal(ee.Data, &productEvent); err != nil {
			return fmt.Errorf("error unmarshaling product created event data: %v", err)
		}
		log.Printf("Processing Product Created Event: %+v", productEvent)
		db.handleProductCreatedEvent(&productEvent)

	case "ping.api.status", "ping.db.status", "ping.cache.status":
		log.Printf("Processing Ping Status Event")
		var pingEvent PingEvent
		if err := json.Unmarshal(ee.Data, &pingEvent); err != nil {
			return fmt.Errorf("error unmarshaling ping status data: %v", err)
		}
		log.Printf("Ping Status for subject: %s:%s", ee.Type, pingEvent.Status)
	default:
		return fmt.Errorf("unknown event type: %s", ee.Type)

	}
	return nil
}

func (db *DBStore) handleProductCreatedEvent(event *PriceCreatedEvent) {
	log.Printf("Generating and storing fake Stripe IDs: %+v", event)
	fakeStripeProductID := fmt.Sprintf("prod_%s", uuid.NewString())
	fakeStripePriceID := fmt.Sprintf("price_%s", uuid.NewString())
	db.Data[event.StbxProductID] = fmt.Sprintf("%s:%s", fakeStripeProductID, fakeStripePriceID)
}
