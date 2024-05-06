package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Cart struct {
	ID    string   `json:"id"`
	Games []string `json:"games"`
}

var dynamodbClient *dynamodb.DynamoDB

func init() {
	region := os.Getenv("AWS_REGION")
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(region),
		Endpoint: aws.String(endpoint),
	})
	if err != nil {
		log.Fatal(err)
	}

	dynamodbClient = dynamodb.New(sess)
}

func main() {
	http.HandleFunc("/carts", getCartsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getCartsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve carts from DynamoDB
	result, err := dynamodbClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("Carts"),
	})
	if err != nil {
		log.Println("Error scanning Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var carts []Cart
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &carts)
	if err != nil {
		log.Println("Error unmarshaling carts:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// todo: change to cart.tml
	renderTemplate(w, "cart.html", map[string]interface{}{
		"Carts": carts,
	})

}

func renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	t, err := template.ParseFiles("templates/" + templateName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
