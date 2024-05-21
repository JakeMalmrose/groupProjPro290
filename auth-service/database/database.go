package database

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user object
type User struct {
	ID       string `json:"id"`
	Audience string `json:"audience"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var db *dynamodb.DynamoDB

func init() {
	// Initialize DynamoDB session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Endpoint: aws.String("http://VaporAuthDynamoDB:8000"),
		},
	}))

	db = dynamodb.New(sess)

	err := InitializeTables()
	if err != nil {
		panic(err)
	}
}

func AuthenticateUser(username, password string) (*User, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String("users"),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(username),
			},
		},
	}

	result, err := db.GetItem(input)
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, fmt.Errorf("user not found")
	}

	var user User
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, err
	}

	// Compare the provided password with the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

func SaveUser(user User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("users"),
		Item:      item,
	}

	_, err = db.PutItem(input)
	if err != nil {
		return err
	}

	return nil
}

func GetUserByUsername(username string) (*User, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String("users"),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(username),
			},
		},
	}

	result, err := db.GetItem(input)
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var user User
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUserRole(username, role string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("users"),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(username),
			},
		},
		UpdateExpression: aws.String("SET #a = :val"),
		ExpressionAttributeNames: map[string]*string{
			"#a": aws.String("audience"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":val": {
				S: aws.String(role),
			},
		},
	}

	_, err := db.UpdateItem(input)
	if err != nil {
		return err
	}

	return nil

}

func InitializeTables() error {
	// Check if the "users" table exists
	_, err := db.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String("users"),
	})

	if err != nil {
		// If the table doesn't exist, create it
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
			err = createUsersTable()
			if err != nil {
				return err
			}

		} else {
			return err
		}
	}
    addDefaultAdmin()
	return nil
}

func addDefaultAdmin() error {
	// Check if the default admin user exists
	_, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("users"),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String("admin"),
			},
		},
	})

	if err != nil {
		// If the user doesn't exist, create it
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
			user := User{
				Username: "gorf",
				Password: "admin",
				Audience: "admin",
			}

			err = SaveUser(user)
			if err != nil {
				return err
			}

		} else {
			return err
		}
	}

	return nil
}

func createUsersTable() error {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("username"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("username"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String("users"),
	}

	_, err := db.CreateTable(input)
	if err != nil {
		return err
	}

	return nil
}
