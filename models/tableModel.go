package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Table struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`                                                      // MongoDB ObjectID
	Table_Number     *int               `bson:"table_number" json:"table_number" validate:"required"`               // Table number (required)
	Number_of_guests *int               `bson:"number_of_guests" json:"number_of_guests" validate:"required,min=1"` // Number of seats (required, minimum 1)
	Table_ID         string             `bson:"table_id,omitempty" json:"table_id,omitempty"`                       // Associated order (optional, if table is occupied)
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`                                       // Time of table creation
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`                                       // Time of last table update
}
