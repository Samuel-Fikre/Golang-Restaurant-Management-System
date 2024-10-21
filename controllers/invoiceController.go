package controllers

import (
	"context"
	"fmt"
	"net/http"
	"restaurant-management-system/database"
	"restaurant-management-system/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// the interface{} type is a special type that can represent any type of value. It essentially means "any type" or "no specific type." It is Go's way of providing flexibility, allowing a variable to hold values of any data type, such as strings, integers, structs, or even slices and maps.

// Different Purposes:

// Model Struct (Invoice): This is used for storing data in the database. It contains all the fields that are necessary for your application's internal logic and data persistence.
// View Struct (InvoiceViewFormat): This is specifically designed for sending data to the client (e.g., an API response). It can include only the fields you want the client to see.

// Using a separate struct helps prevent exposing sensitive or unnecessary information to the client. You might want to include only certain fields in the response and exclude others.

// eg

// type Invoice struct {
//Customer_email    string             `bson:"customer_email" json:"customer_email"` // Sensitive information
//}

// type InvoiceViewFormat struct {
// 	Omitting Customer_email to protect sensitive information
// }

type InvoiceViewFormat struct {
	Invoice_id       string      `json:"invoice_id"`
	Order_id         string      `json:"order_id"`
	Payment_method   string      `json:"payment_method"`
	Payment_status   *string     `json:"payment_status"`
	Payment_due      interface{} `json:"payment_due"`
	Table_number     interface{} `json:"table_number"`
	Payment_due_date time.Time   `json:"payment_due_date"`
	Order_details    interface{} `json:"order_details"`
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		result, err := invoiceCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		var allinvoices []bson.M
		if err = result.All(ctx, &allinvoices); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, allinvoices)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		invoiceId := c.Param("invoice_id")
		var invoice models.Invoice

		err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		var invoiceView InvoiceViewFormat
		//  The purpose of this function is likely to retrieve a list of items associated with a specific order
		allOrderItems, err := ItemsByOrder(invoice.Order_id)

		// we are assigning values from the invoice struct to the invoiceView struct. Heres a more detailed explanation of the direction of assignment:

		invoiceView.Order_id = invoice.Order_id
		invoiceView.Invoice_id = invoice.Invoice_id

		//The first line initializes the Payment_method in invoiceView to "null" as a default value.
		//The second line checks if the Payment_method in invoice is not nil, and if it has a valid value, assigns that value to invoiceView.Payment_method, overriding the default value.
		// he use of "null" (as a string) might imply that there was no payment method associated with the invoice, or it can be a placeholder for scenarios where the actual method is not available.
		// If the condition is true (meaning Payment_method has a value), it dereferences the pointer to set Payment_method in the invoiceView struct. The *invoice.Payment_method syntax retrieves the actual string value that the pointer points to. for eg CREDIT
		// the * operator is used to dereference the pointer to access the actual value of the Payment_method stored in the invoice struct. This is necessary because Payment_method is defined as a pointer (*string)
		invoiceView.Payment_method = "null"
		if invoice.Payment_method != nil {
			invoiceView.Payment_method = *invoice.Payment_method
		}

		// Payment_status is a field in the invoice struct that is defined as a pointer (*string). This means it can hold either nil (indicating no value) or a reference to a string value (e.g., "PAID" or "PENDING")
		invoiceView.Payment_status = &*invoice.Payment_status
		invoiceView.Payment_due = allOrderItems[0]["payment_due"]
		invoiceView.Table_number = allOrderItems[0]["table_number"]
		invoiceView.Order_details = allOrderItems[0]["order_items"]

		c.JSON(http.StatusOK, invoiceView)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var invoice models.Invoice

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var order models.Order

		err := orderCollection.FindOne(ctx, bson.M{"order_id": invoice.Order_id}).Decode(&order)
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("message: Order was not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		status := "PENDING"
		if invoice.Payment_method == nil {
			invoice.Payment_status = &status
		}

		// The underscore (_) is used to ignore any errors that might occur during parsing, meaning even if an error is returned by time.Parse, it's ignored here.
		invoice.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		// AddDate(0, 0, 1)
		// AddDate(years, months, days)
		// Here, you're adding 1 day to the current date because the parameters 0, 0, 1 mean "add 0 years, 0 months, and 1 day."
		// The result is a new date that is 1 day after the current date.
		invoice.Payment_due_date, _ = time.Parse(time.RFC3339, time.Now().AddDate(0, 0, 1).Format(time.RFC3339))
		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_id = invoice.ID.Hex()
		var validate = validator.New()

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		result, insertErr := invoiceCollection.InsertOne(ctx, invoice)
		if insertErr != nil {
			msg := fmt.Sprintf("Invoice item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var invoice models.Invoice
		invoiceId := c.Param("invoice_id")
		var updateObj primitive.D

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filter := bson.M{"invoice_id": invoiceId}

		if invoice.Payment_method != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_method", Value: invoice.Payment_method})
		}
		if invoice.Payment_status != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_status", Value: invoice.Payment_status})
		}

		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: invoice.Updated_at})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		//Memory Reference: By using &, you're not copying the string value "PENDING" into Payment_status. Instead, you're storing a reference to it.
		//Dynamic Changes: If you change the value of status later, Payment_status will still point to it and reflect the change.

		status := "PENDING"
		if invoice.Payment_method == nil {
			invoice.Payment_status = &status
		}

		result, err := invoiceCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer cancel()

		c.JSON(http.StatusOK, result)
	}
}
