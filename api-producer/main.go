package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

var (
	pingedAPISubject    = "ping.api.status"
	pingedDBSubject     = "ping.db.status"
	pingedCacheSubject  = "ping.cache.status"
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

	PriceCreateRequest struct {
		Name     string `json:"name"`
		Amount   int    `json:"amount"`
		Currency string `json:"currency"`
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
)

func main() {
	// Startup API and Nats connection
	log.Print("Starting API Producer Service...")

	natsURI := os.Getenv("NATS_URI")
	log.Printf("Attempting to connect to NATS at %s", natsURI)

	nc, err := nats.Connect(natsURI)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	log.Print("Succesfully connected to NATS")

	r := gin.Default()

	// Setup routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	r.GET("/ping", func(c *gin.Context) {
		//API health good we got this
		apiMsg := PingEvent{Status: "Healthy"}
		//Pretend DB is healthy
		dbMsg := PingEvent{Status: "Healthy"}
		//Pretending cache is down
		cacheMsg := PingEvent{Status: "Down"}
		apiMsgBytes, _ := json.Marshal(apiMsg)
		dbMsgBytes, _ := json.Marshal(dbMsg)
		cacheMsgBytes, _ := json.Marshal(cacheMsg)
		publishMessage(nc, pingedAPISubject, apiMsgBytes)
		publishMessage(nc, pingedDBSubject, dbMsgBytes)
		publishMessage(nc, pingedCacheSubject, cacheMsgBytes)

		c.JSON(http.StatusOK, apiMsg)
	})

	r.POST("/products", func(c *gin.Context) {
		log.Print("Parsing product created event")
		var newProduct PriceCreateRequest
		if err := c.ShouldBindJSON(&newProduct); err != nil {
			log.Printf("Error binding json: %v", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//Mapping can happen in a manager not in the controller
		newProductEvent := PriceCreatedEvent{
			StbxProductID: uuid.New(),
			StbxPriceID:   uuid.New(),
			Name:          newProduct.Name,
			Amount:        newProduct.Amount,
			Currency:      newProduct.Currency,
		}
		msgBytes, err := json.Marshal(newProductEvent)
		if err != nil {
			log.Printf("Error marshaling product event: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal product event"})
			return
		}

		log.Printf("Publishing product created event: %+v", newProductEvent)
		publishMessage(nc, priceCreatedSubject, msgBytes)
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	r.Run(":8080")
}

func publishMessage(nc *nats.Conn, subject string, message []byte) {
	eventData := EventEnvelope{
		ID:        uuid.New(),
		Type:      subject,
		Version:   1,
		Timestamp: time.Now(),
		Data:      message,
	}
	eventBytes, err := json.Marshal(eventData)
	if err != nil {
		log.Printf("Error marshaling event data: %v", err)
		return
	}
	err = nc.Publish(subject, eventBytes)
	if err != nil {
		log.Printf("Error publishing message to subject %s: %v", subject, err)
	} else {
		log.Printf("Published message to subject %s: %s", subject, message)
	}
}
