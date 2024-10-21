package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	Quantity *string `bson:"quantity" json:"quantity" validate:"required, eq=S|eq=M|eq=L"`

	Order_ID      string `bson:"order_id" json:"order_id" validate:"required"`
	Order_Item_Id string `bson:"order_item_id" json:"order_item_id" `

	Unit_Price *float64 `bson:"unit_price" json:"unit_price" validate:"required"`
	Food_id    *string  `bson:"food_id" json:"food_id" validate:"required" `

	CreatedAt time.Time `bson:"created_at" json:"created_at"` // Time of order creation
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
