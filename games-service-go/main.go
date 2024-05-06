package main

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
    "log"
    "net/http"
    "os"
)

type Game struct {
    ID     string  `json:"id"`
    Title  string  `json:"title"`
    Price  float64 `json:"price"`
    Owned  bool    `json:"owned"`
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
    http.HandleFunc("/games", getGamesHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func getGamesHandler(w http.ResponseWriter, r *http.Request) {
    // Retrieve games from DynamoDB
    result, err := dynamodbClient.Scan(&dynamodb.ScanInput{
        TableName: aws.String("Games"),
    })
    if err != nil {
        log.Println("Error scanning Games table:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    var games []Game
    err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &games)
    if err != nil {
        log.Println("Error unmarshaling games:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Render the games list template with the retrieved games data
    // ...
}