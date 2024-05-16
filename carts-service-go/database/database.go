package database

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	structs "github.com/Draupniyr/carts-service/structs"
)

type Database struct {
	DynamodbClient *dynamodb.DynamoDB
}

func (db *Database) Init() {
	log.Println("Initializing database")
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	log.Println("Endpoint:", endpoint)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Endpoint: aws.String(endpoint),
		},
	}))

	db.DynamodbClient = dynamodb.New(sess) //ineffective assignment to field DataBase.DynamodbClient (SA4005)

	// if Carts table does not exist, create it maybe
	_, err := db.DynamodbClient.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String("Carts"),
	})
	if err != nil {
		log.Println("Table does not exist, creating it")
		db.InitializeTables()
	}
}

func (db *Database) InitializeTables() error {
		_, err := db.DynamodbClient.CreateTable(&dynamodb.CreateTableInput{
			TableName: aws.
				String("Carts"),
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("ID"),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("ID"),
					KeyType: aws.String("HASH"),
				},
			},
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(5),
				WriteCapacityUnits: aws.Int64(5),
			},
		})
		if err != nil {
			log.Fatal(err)
		}
	return nil
}

// ----------------- Carts -----------------
func (db *Database) GetCart(ID string) (*structs.Cart, error) {
	result, err := db.DynamodbClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Carts"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(ID),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	Cart := structs.Cart{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &Cart)
	if err != nil {
		return nil, err
	}

	return &Cart, nil
}

func (db *Database) GetAllCarts() ([]structs.Cart, error) {
	result, err := db.DynamodbClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("Carts"),
	})
	if err != nil {
		return nil, err
	}

	Carts := []structs.Cart{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &Carts)
	if err != nil {
		return nil, err
	}

	return Carts, nil
}

func (db *Database) CreateAndUpdateCart(Cart structs.Cart) error {
	item, err := dynamodbattribute.MarshalMap(Cart)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
        TableName: aws.String("Carts"),
        Item:      item,
    }
	db.DynamodbClient.PutItem(input)

	return nil
}

func (db *Database) AddOrRemoveFromCart(ID string, gamesList []structs.Game) error {
	cartOG, err := db.GetCart(ID)
	if err != nil {
		return err
	}

	contains := false
	for _, game := range cartOG.Games {
		for _, game2 := range gamesList {
			if game.ID == game2.ID {
				contains = true
			}
		}
	}
	if contains {
		newGamesList := []structs.Game{}
		for _, game := range cartOG.Games {
			for _, game2 := range gamesList {
				if game.ID != game2.ID {
					newGamesList = append(newGamesList, game)
				}
			}
		}
		cartOG.Games = newGamesList
	} else {
		cartOG.Games = append(cartOG.Games, gamesList...)
	}

	db.CreateAndUpdateCart(*cartOG)
	return nil
}

func (db *Database) DeleteCart(ID string) error {
	_, err := db.DynamodbClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("Carts"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(ID),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) DeleteAll() error {
	db.DynamodbClient.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String("Carts"),
	})
	db.InitializeTables()
	return nil
}
