package database

import (
	"log"
	"os"
	"fmt"

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
				KeyType:       aws.String("HASH"),
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
func (db *Database) GetCart(userID string) (structs.Cart, error) {
    if userID == "" {
        return structs.Cart{}, fmt.Errorf("userID is required")
    }

    result, err := db.DynamodbClient.Scan(&dynamodb.ScanInput{
        TableName:        aws.String("Carts"),
        FilterExpression: aws.String("UserID = :userID"),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":userID": {
                S: aws.String(userID),
            },
        },
    })
    if err != nil {
        log.Printf("Error scanning Carts table: %v", err)
        return structs.Cart{}, err
    }

    if len(result.Items) == 0 {
        log.Printf("Cart not found for UserID: %s", userID)
        return structs.Cart{}, fmt.Errorf("cart not found for UserID: %s", userID)
    }

    if len(result.Items) > 1 {
        log.Printf("Multiple carts found for UserID: %s", userID)
        return structs.Cart{}, fmt.Errorf("multiple carts found for UserID: %s", userID)
    }

    var cart structs.Cart
    err = dynamodbattribute.UnmarshalMap(result.Items[0], &cart)
    if err != nil {
        log.Printf("Error unmarshaling cart: %v", err)
        return structs.Cart{}, err
    }

    log.Printf("Retrieved cart for UserID: %s, Cart: %+v", userID, cart)
    return cart, nil
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

func (db *Database) AddOrRemoveFromCart(ID string, gameToAddOrRemove structs.Game) error {
	cartOG, err := db.GetCart(ID)
	if err != nil {
		return err
	}

	newgames := []structs.Game{}
	contains := false
	for _, game := range cartOG.Games {
		log.Println("Checking: ", game.ID, " against: ", gameToAddOrRemove.ID)
		if game.ID == gameToAddOrRemove.ID {
			contains = true
			log.Println("Contains: ", game.ID)
		}
	}
	if contains {
		for _, game := range cartOG.Games {
			if game.ID != gameToAddOrRemove.ID {
				newgames = append(newgames, game)
				log.Println("Adding: ", game.ID)
			}
		}
		cartOG.Games = newgames
	} else {
		cartOG.Games = append(cartOG.Games, gameToAddOrRemove)
		log.Println("Adding new: ", gameToAddOrRemove.ID)
	}

	db.CreateAndUpdateCart(cartOG)
	return nil
}

func (db *Database) DeleteCart(UserID string) error {

	cart, err := db.GetCart(UserID)
	if err != nil {
		return err
	}

	_, err = db.DynamodbClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("Carts"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(cart.ID),
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
