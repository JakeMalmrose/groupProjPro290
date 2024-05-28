package logic

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	database "github.com/Draupniyr/games-service/database"
	structs "github.com/Draupniyr/games-service/structs"
)

// ----------------- Games -----------------
func GetGame(ID string, db database.DatabaseFunctionality) (*structs.Game, error) {
	game := structs.Game{}
	err := db.GetFilter(ID, "ID", &game)
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func SearchGames(search string, db database.DatabaseFunctionality) ([]structs.Game, error) {
	AllFoundGames := []structs.Game{}

	games := []structs.Game{}
	db.GetFilter(search, "Title", &games)
	AllFoundGames = append(AllFoundGames, games...)

	games = []structs.Game{}
	db.GetFilter(search, "Description", &games)
	AllFoundGames = append(AllFoundGames, games...)

	games = []structs.Game{}
	db.GetFilter(search, "Tags", &games)
	AllFoundGames = append(AllFoundGames, games...)

	if len(AllFoundGames) == 0 {
		return nil, errors.New("no games found")
	}

	// Remove duplicates
	encountered := map[string]bool{}
	uniqueGames := []structs.Game{}

	for _, game := range AllFoundGames {
		if !encountered[game.ID] {
			encountered[game.ID] = true
			uniqueGames = append(uniqueGames, game)
		}
	}

	return uniqueGames, nil
}

func GetGamesByAuthor(authorID string, db database.DatabaseFunctionality) ([]structs.Game, error) {
	games := []structs.Game{}
	err := db.GetFilter(authorID, "AuthorID", &games)
	if err != nil {
		return nil, err
	}

	return games, nil
}

func GetAllGames(db database.DatabaseFunctionality) ([]structs.Game, error) {
	games := []structs.Game{}
	err := db.GetAll(&games)
	if err != nil {
		return nil, err
	}

	return games, nil
}

func CreateGame(game structs.Game, db database.DatabaseFunctionality) error {
	games := []structs.Game{}
	err := db.GetFilter(game.Title, "Title", &games)
	if err != nil {
		err := db.CreateOrUpdate(game)
		if err != nil {
			return err
		}
	}
	return nil

}

func UpdateGame(ID string, userid string, game structs.Game, db database.DatabaseFunctionality) error {
	ogGame := structs.Game{}
	err := db.GetFilter(ID, "ID", &ogGame)
	if err != nil {
		return err
	}
	if ogGame.AuthorID == userid {
		game.ID = ID
		err := db.CreateOrUpdate(game)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateGameField(ID string, field string, value string, db database.Database) error {
	_, err := db.DynamodbClient.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Games"),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(ID),
			},
		},
		UpdateExpression: aws.String("set " + field + " = :v"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {
				S: aws.String(value),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func DeleteGame(ID string, userId string, db database.DatabaseFunctionality) error {
	game := structs.Game{}
	err := db.GetFilter(ID, "ID", &game)
	if err != nil {
		return err
	}
	if game.AuthorID != userId {
		err := db.Delete(ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteGameByID(ID string, db database.DatabaseFunctionality) error {
	err := db.Delete(ID)
	if err != nil {
		return err
	}
	return nil
}
func DeleteAll(db database.DatabaseFunctionality) error {
	db.DeleteAll()
	return nil
}

// ----------------- Updates -----------------
func CreateUpdate(ID string, userId string, update structs.Update, db database.DatabaseFunctionality) error {
	currentGame := structs.Game{}
	err := db.GetFilter(ID, "ID", &currentGame)
	if err != nil {
		return err
	}
	if currentGame.AuthorID != userId {
		return nil
	}
	currentGame.Updates = append(currentGame.Updates, update)
	db.CreateOrUpdate(currentGame)
	return nil
}

func DeleteUpdate(ID string, userId string, updateID string, db database.DatabaseFunctionality) error {
	currentGame := structs.Game{}
	err := db.GetFilter(ID, "ID", &currentGame)
	if err != nil {
		return err
	}
	if currentGame.AuthorID != userId {
		return nil
	}
	for i, update := range currentGame.Updates {
		if update.ID == updateID {
			currentGame.Updates = append(currentGame.Updates[:i], currentGame.Updates[i+1:]...)
			break
		}
	}
	db.CreateOrUpdate(currentGame)
	return nil
}

func GetUpdate(ID string, updateID string, db database.DatabaseFunctionality) (*structs.Update, error) {
	currentGame := structs.Game{}
	err := db.GetFilter(ID, "ID", &currentGame)
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

func UpdateUpdate(ID string, userId string, updateID string, update structs.UpdatePostObject, db database.DatabaseFunctionality) error {
	currentGame := structs.Game{}
	err := db.GetFilter(ID, "ID", &currentGame)
	if err != nil {
		return err
	}
	if currentGame.AuthorID != userId {
		return nil
	}
	for i, ogupdate := range currentGame.Updates {
		if ogupdate.ID == updateID {
			currentGame.Updates[i].Title = update.Title
			currentGame.Updates[i].Content = update.Content
			db.CreateOrUpdate(currentGame)
			return nil
		}
	}
	return nil
}
