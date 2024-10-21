package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Menu struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`                                             // MongoDB ObjectID
	Name       string             `bson:"name" json:"name" validate:"required,min=2,max=100"`        // Menu name (required, length 2-100)
	Category   string             `bson:"category" json:"category" validate:"required,min=3,max=50"` // Menu category (required, length 3-50)
	Start_date *time.Time         `bson:"start_date" json:"start_date" validate:"required"`          // Start date of the menu (required)
	End_date   *time.Time         `bson:"end_date" json:"end_date" validate:"required"`              // End date of the menu (required)
	Created_at time.Time          `bson:"created_at" json:"created_at"`                              // Time of menu creation
	Updated_at time.Time          `bson:"updated_at" json:"updated_at"`                              // Time of last menu update
	Menu_id    string             `bson:"menu_id" json:"menu_id"`                                    // Custom menu identifier
}
