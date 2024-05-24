package mockdb

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Database struct {
	TableName      string
	IdName         string
	DynamodbClient []interface{}
}

// ----------------- Helper -----------------
func (db *Database) Init(tableName string, idName string) error {
	db.TableName = tableName
	db.IdName = capitalizeFirstLetter(idName)
	db.InitializeTables()
	return nil
}

func (db *Database) InitializeTables() error {
	db.DynamodbClient = []interface{}{}
	return nil
}

// ----------------- Items -----------------
func (db *Database) GetFilter(attributeValue string, attributeName string, output interface{}) error {
	resultStore := []interface{}{}
	for _, item := range db.DynamodbClient {
		attribute, err := getIDValue(item, attributeName)
		if err != nil {
			return err
		}
		if attribute == attributeValue {
			resultStore = append(resultStore, item)
		}
	}
	if len(resultStore) == 0 {
		return fmt.Errorf("item with %s %s not found", attributeName, attributeValue)
	}

	outputValue := reflect.ValueOf(output)
	if outputValue.Kind() != reflect.Ptr || outputValue.IsNil() {
		return errors.New("output must be a non-nil pointer")
	}

	if len(resultStore) > 1 {
		outputValue.Elem().Set(reflect.ValueOf(resultStore))
	} else {
		outputValue.Elem().Set(reflect.ValueOf(resultStore[0]))
	}
	return nil
}

func (db *Database) GetAll(output interface{}) error {
	// Use reflection to ensure output is a pointer to a slice
	outVal := reflect.ValueOf(output)
	if outVal.Kind() != reflect.Ptr {
		return fmt.Errorf("output must be a pointer to a slice")
	}
	outVal = outVal.Elem()
	if outVal.Kind() != reflect.Slice {
		return fmt.Errorf("output must be a pointer to a slice")
	}

	// Iterate over the items in the mock database and append to the slice
	for _, item := range db.DynamodbClient {
		itemVal := reflect.ValueOf(item)
		if itemVal.Type().AssignableTo(outVal.Type().Elem()) {
			outVal.Set(reflect.Append(outVal, itemVal))
		} else {
			return fmt.Errorf("item type mismatch: expected %v but got %v", outVal.Type().Elem(), itemVal.Type())
		}
	}
	return nil
}

func (db *Database) CreateOrUpdate(object interface{}) error {
	id, err := getIDValue(object, db.IdName)
	if err != nil {
		return err
	}
	isUpdate := false
	for i, item := range db.DynamodbClient {
		itemID, err := getIDValue(item, db.IdName)
		if err != nil {
			return err
		}
		if itemID == id {
			db.DynamodbClient[i] = object
			isUpdate = true
			return nil
		}
	}
	if !isUpdate {
		db.DynamodbClient = append(db.DynamodbClient, object)
	}
	return nil
}

func (db *Database) Delete(idValue string) error {
	for i, item := range db.DynamodbClient {
		id, err := getIDValue(item, db.IdName)
		if err != nil {
			return err
		}
		if id == idValue {
			db.DynamodbClient = append(db.DynamodbClient[:i], db.DynamodbClient[i+1:]...)
		}
	}
	return nil

}

func (db *Database) DeleteFilter(attributeValue string, attrbuteName string) error {
	for i, item := range db.DynamodbClient {
		attribute, err := getIDValue(item, attrbuteName)
		if err != nil {
			return err
		}
		if attribute == attributeValue {
			db.DynamodbClient = append(db.DynamodbClient[:i], db.DynamodbClient[i+1:]...)
		}
	}
	return nil
}

func (db *Database) DeleteAll() error {
	db.DynamodbClient = []interface{}{}
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
