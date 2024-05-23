package database

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Database struct {
	TableName      string
	IdName         string
	DynamodbClient *dynamodb.DynamoDB
}

// ----------------- Helper -----------------
func (db *Database) Init(tableName string, idName string) error {
	db.TableName = tableName
	db.IdName = capitalizeFirstLetter(idName)
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Endpoint: aws.String(endpoint),
		},
	}))

	db.DynamodbClient = dynamodb.New(sess) //ineffective assignment to field DataBase.DynamodbClient (SA4005)

	// if table does not exist, create it
	_, err := db.DynamodbClient.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(db.TableName),
	})
	if err != nil {
		log.Println("Table does not exist, creating it")
		err = db.InitializeTables()
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *Database) InitializeTables() error {
	_, err := db.DynamodbClient.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.
			String(db.TableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(db.IdName),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(db.IdName),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// ----------------- Items -----------------
func (db *Database) GetFilter(attributeValue string, attributeName string, output interface{}) error {
	if attributeValue == "" {
		return fmt.Errorf("attributeValue is required")
	}
	if attributeName == "" {
		return fmt.Errorf("attributeName is required")
	}
	filterExpression := capitalizeFirstLetter(attributeName) + " = :" + lowercaseFirstLetter(attributeName)

	result, err := db.DynamodbClient.Scan(&dynamodb.ScanInput{
		TableName:        aws.String(db.TableName),
		FilterExpression: aws.String(filterExpression),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":" + lowercaseFirstLetter(attributeName): {
				S: aws.String(attributeValue),
			},
		},
	})
	if err != nil {
		return err
	}
	if len(result.Items) == 0 {
		return fmt.Errorf("item not found for "+attributeName+": ", attributeValue)
	}
	// Check the type of output and unmarshal accordingly
	if reflect.TypeOf(output).Kind() == reflect.Ptr && reflect.TypeOf(output).Elem().Kind() == reflect.Slice {
		// Unmarshal list of items
		err = dynamodbattribute.UnmarshalListOfMaps(result.Items, output)
		if err != nil {
			return err
		}
	} else {
		// Unmarshal single item
		if len(result.Items) > 1 {
			return fmt.Errorf("more than one item found for %s: %s", attributeName, attributeValue)
		}
		err = dynamodbattribute.UnmarshalMap(result.Items[0], output)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *Database) GetAll(output interface{}) error {
	result, err := db.DynamodbClient.Scan(&dynamodb.ScanInput{
		TableName: aws.String(db.TableName),
	})
	if err != nil {
		return err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, output)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) CreateOrUpdate(object interface{}) error {
	item, err := dynamodbattribute.MarshalMap(object)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(db.TableName),
		Item:      item,
	}
	db.DynamodbClient.PutItem(input)
	return nil
}

func (db *Database) Delete(idValue string) error {
	_, err := db.DynamodbClient.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(db.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			db.IdName: {
				S: aws.String(idValue),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteFilter(attributeValue string, attrbuteName string) error {
	array := []any{}
	db.GetFilter(attributeValue, attrbuteName, array)
	if len(array) == 0 {
		return fmt.Errorf("item not found for "+attrbuteName+": ", attributeValue)
	}

	for _, item := range array {
		id, err := getIDValue(item, db.IdName)
		if err != nil {
			return err
		}
		err = db.Delete(id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *Database) DeleteAll() error {
	db.DynamodbClient.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(db.TableName),
	})
	db.InitializeTables()
	return nil
}

func getIDValue(item interface{}, fieldName string) (string, error) {
	r := reflect.ValueOf(item)
	f := reflect.Indirect(r).FieldByName(fieldName)
	if !f.IsValid() {
		return "", fmt.Errorf("field %s not found", fieldName)
	}
	return f.String(), nil
}

func capitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func lowercaseFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(string(s[0])) + s[1:]
}
