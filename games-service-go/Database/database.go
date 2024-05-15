package database

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	structs "github.com/Draupniyr/games-service/structs"
)

type Database struct {
	DynamodbClient *dynamodb.DynamoDB
}

func (db *Database) Init() {
	region := os.Getenv("AWS_REGION")
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String(region),
		Endpoint: aws.String(endpoint),
	})
	if err != nil {
		log.Fatal(err)
	}

	db.DynamodbClient = dynamodb.New(sess) //ineffective assignment to field DataBase.DynamodbClient (SA4005)

	// if Games table does not exist, create it maybe
	_, err = db.DynamodbClient.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String("Games"),
	})
	if err != nil {
		_, err = db.DynamodbClient.CreateTable(&dynamodb.CreateTableInput{
			TableName: aws.
				String("Games"),
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("ID"),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("ID"),
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
	}
}

// ----------------- Games -----------------
func (db *Database) GetGame(ID string) (*structs.Game, error) {
	result, err := db.DynamodbClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Games"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(ID),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	game := structs.Game{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &game)
	if err != nil {
		return nil, err
	}

	return &game, nil
}
func (db *Database) SearchGames(search string) ([]structs.Game, error) {
	result, err := db.DynamodbClient.Scan(&dynamodb.ScanInput{
		TableName:        aws.String("Games"),
		FilterExpression: aws.String("contains(Title, :search) OR contains(Description, :search) OR contains(Tags, :search)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":search": {
				S: aws.String(search),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	games := []structs.Game{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &games)
	if err != nil {
		return nil, err
	}

	return games, nil
}
func (db *Database) GetAllGames() ([]structs.Game, error) {
	result, err := db.DynamodbClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String("Games"),
	})
	if err != nil {
		return nil, err
	}

	games := []structs.Game{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &games)
	if err != nil {
		return nil, err
	}

	return games, nil
}
func (db *Database) CreateGame(game structs.Game) error {
	_, err := db.DynamodbClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("Games"),
		Item:      game.GameToDynamoDBItem(),
	})
	if err != nil {
		return err
	}

	return nil
}
func (db *Database) UpdateGame(ID string, game structs.Game) error {
	game.ID = ID
	game.Published = ""
	updateString := game.GameToUpdateString()
	_, err := db.DynamodbClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Games"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(ID),
			},
		},
		UpdateExpression:          &updateString,
		ExpressionAttributeValues: game.GameToDynamoDBUpdateItem(),
		ReturnValues:              aws.String("UPDATED_NEW"),
	})
	if err != nil {
		return err
	}

	return nil
}
func (db *Database) DeleteGame(ID string) error {
	_, err := db.DynamodbClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("Games"),
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
		TableName: aws.String("Games"),
	})
	return nil
}

// ----------------- Updates -----------------
func (db *Database) CreateUpdate(ID string, update structs.Update) error {
	currentGame, err := db.GetGame(ID)
	currentGame.Updates = append(currentGame.Updates, update)
	if err != nil {
		return err
	}
	db.UpdateGame(ID, *currentGame)
	return nil
}
func (db *Database) DeleteUpdate(ID string, updateID string) error {
	currentGame, err := db.GetGame(ID)
	if err != nil {
		return err
	}
	for i, update := range currentGame.Updates {
		if update.ID == updateID {
			currentGame.Updates = append(currentGame.Updates[:i], currentGame.Updates[i+1:]...)
			break
		}
	}
	db.UpdateGame(ID, *currentGame)
	return nil
}
func (db *Database) GetUpdate(ID string, updateID string) (*structs.Update, error) {
	currentGame, err := db.GetGame(ID)
	if err != nil {
		return nil, err
	}
	for _, update := range currentGame.Updates {
		if update.ID == updateID {
			return &update, nil
		}
	}
	return nil, nil
}
func (db *Database) UpdateUpdate(ID string, updateID string, update structs.UpdatePostObject) error {
	currentGame, err := db.GetGame(ID)
	if err != nil {
		return err
	}
	for i, update := range currentGame.Updates {
		if update.ID == updateID {
			if update.Title != "" {
				currentGame.Updates[i].Title = update.Title
			}
			if update.Content != "" {
				currentGame.Updates[i].Content = update.Content
			}
			break
		}
	}
	db.UpdateGame(ID, *currentGame)
	return nil
}
