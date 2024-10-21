package models

import (
	"time" //updated and created is needed that's why

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// omitempty ensures that when creating a new record, the ID will be automatically generated.

//  pointer to a string (*string), with validation rules that ensure it's required and has a minimum length of 2 and a maximum length of 100.

// Name and Price are marked as pointers (*string and *float64), meaning they can be nil if not provided. If they weren't pointers, Go would initialize them to their zero values ("" for strings, 0 for floats), which might not represent the absence of data proper

// Use pointers if you want to allow the field to be omitted or set to nil.

type Food struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       *string            `json:"name" validate:"required,min=2,max=100"` // Name of the food, required and between 2-100 characters
	Price      *float64           `json:"price" validate:"required"`              // Price of the food, required
	Food_image *string            `json:"food_image" validate:"required"`
	Created_at time.Time          `bson:"created_at" json:"created_at"`               // Time of creation
	Updated_at time.Time          `bson:"updated_at" json:"updated_at"`               // Time of last update
	Food_id    string             `bson:"food_id" json:"food_id"`                     // Custom food identifier
	Menu_id    *string            `bson:"menu_id" json:"menu_id" validate:"required"` // Reference to the menu the food belongs to

}
