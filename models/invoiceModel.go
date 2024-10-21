package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Why use pointers (*string): There are a few reasons to use pointers in Go:

// Optional values: By using a pointer, you can represent an "optional" value. A regular string can't be nil (null), but a pointer to a string (*string) can be nil, meaning it hasn't been set.
// Avoid copying large data: If you pass large data around by value, Go makes a copy of the data. With pointers, you can avoid making copies and just pass the memory address, which is more efficient for large data.

type Invoice struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`                                                               // MongoDB ObjectID
	Invoice_id       string             `bson:"invoice_id" json:"invoice_id"`                                                // Custom invoice identifier
	Order_id         string             `bson:"order_id" json:"order_id"`                                                    // Reference to the associated order
	Payment_method   *string            `bson:"payment_method" json:"payment_method" validate:"eq=CARD|eq=CASH|eq="`         // The validation rule ensures that the Payment_method field can only be one of the specified values ("CARD" or "CASH") or left empty.
	Payment_status   *string            `bson:"payment_status" json:"payment_status" validate:"required,eq=PENDING|eq=PAID"` // Status of the payment (e.g., "paid", "pending"), required
	Payment_due_date time.Time          `bson:"payment_due_date" json:"payment_due_date" validate:"required"`                // Amount due, required
	Created_at       time.Time          `bson:"created_at" json:"created_at"`                                                // Time of invoice creation
	Updated_at       time.Time          `bson:"updated_at" json:"updated_at"`                                                // Time of last invoice update
}
