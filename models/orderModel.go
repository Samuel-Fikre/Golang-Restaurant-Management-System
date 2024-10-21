package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`                                    // MongoDB ObjectID
	Order_Date time.Time          `bson:"order_date" json:"order_date" validate:"required"` // Custom order identifier (required)
	Table_ID   *string            `bson:"table_id" json:"table_id" validate:"required"`     // Reference to the associated Table (required)
	Order_ID   string             `bson:"order_id" json:"order_id" `
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"` // Time of order creation
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"` // Time of last order update
}
