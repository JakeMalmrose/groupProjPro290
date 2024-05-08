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
	http.HandleFunc("/carts", CartsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func CartsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve carts from DynamoDB
	switch r.Method {
	case http.MethodGet:
		getCarts(w, r)
	case http.MethodPost:
		createCart(w, r)
	case http.MethodDelete:
		deleteCart(w, r)
	case http.MethodPut:
		updateCart(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func getCarts(w http.ResponseWriter, r *http.Request) {
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
	renderTemplate(w, "cart.html", map[string]interface{}{
		"Carts": carts,
	})
}

func createCart(w http.ResponseWriter, r *http.Request) {
	cart := Cart{
		ID:    r.FormValue("id"),
		Games: []string{},
	}
	cart.Games = append(cart.Games, r.FormValue("game"))
	av, err := dynamodbattribute.MarshalMap(cart)
	if err != nil {
		log.Println("Error marshaling cart:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	_, err = dynamodbClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("Carts"),
		Item:      av,
	})
	if err != nil {
		log.Println("Error putting item into Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// http.Redirect(w, r, "/carts", http.StatusSeeOther)
}
func deleteCart(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	_, err := dynamodbClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("Carts"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		log.Println("Error deleting item from Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/carts", http.StatusSeeOther)
}

func updateCart(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	game := r.FormValue("game")
	_, err := dynamodbClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Carts"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		UpdateExpression: aws.String("SET #games = list_append(#games, :game)"),
		ExpressionAttributeNames: map[string]*string{
			"#games": aws.String("games"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":game": {
				SS: []*string{aws.String(game)},
			},
		},
	})
	if err != nil {
		log.Println("Error updating item in Carts table:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/carts", http.StatusSeeOther)
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
