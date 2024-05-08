package main

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
    "log"
    "net/http"
    "os"
    "html/template"
    "time"
    "go.etcd.io/etcd/clientv3"
    "context"
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
    etcdEndpoint := os.Getenv("ETCD_ENDPOINT")
    cli, err := clientv3.New(clientv3.Config{
        Endpoints:   []string{etcdEndpoint},
        DialTimeout: 5 * time.Second,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer cli.Close()

    // Register the games service with etcd
    serviceKey := "/traefik/http/services/games/loadbalancer/servers/0/url"
    serviceValue := "http://games-service:8080"
    _, err = cli.Put(context.Background(), serviceKey, serviceValue)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Registered service: %s -> %s\n", serviceKey, serviceValue)


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

    renderTemplate(w, "gameslist.html", map[string]interface{}{
        "Games": games,
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