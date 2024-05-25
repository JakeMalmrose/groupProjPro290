package logic

import (
	"testing"

	"github.com/Draupniyr/games-service/structs"
	database "github.com/Draupniyr/games-service/mockdb"
)

var db database.Database

func TestGetAllGames(t *testing.T) {
	//setup
	db.Init("Test", "ID")
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game1", "User1"))
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game2", "User1"))
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game3", "User1"))

	// TestGetAllGames tests the GetAllGames function.
	Games, err := GetAllGames(&db)
	if err != nil {
		t.Errorf("Error getting all Games: %v", err)
	}

	// It should return all Games in the database.
	if len(Games) != 3 {
		simpleAssert(t, 3, len(Games))
	}
}

func TestGetGame(t *testing.T) {
	db.Init("Test", "ID")
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game1", "User1"))
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game2", "User2"))
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game3", "User3"))
	// TestGetGame tests the GetGame function.
	// It should return a Game with the given userID.
	Game, err := GetGame("Game2", &db)
	if err != nil {
		t.Errorf("Error getting Game: %v", err)
	}
	simpleAssert(t, "Game2", Game.ID)
	simpleAssert(t, "User2", Game.AuthorID)
}

func TestSearchGames(t *testing.T) {
	db.Init("Test", "ID")
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game1", "User1"))
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game2", "User2"))
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game3", "User3"))
	// TestSearchGames tests the SearchGames function.
	// It should return all Games with the given search string.
	Games, err := SearchGames("Test", &db)
	if err != nil {
		t.Errorf("Error searching Games: %v", err)
	}
	simpleAssert(t, 3, len(Games))
}

func TestGetGamesByAuthor(t *testing.T) {
	db.Init("Test", "ID")
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game1", "User1"))
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game2", "User2"))
	db.DynamodbClient = append(db.DynamodbClient, createTestGame("Game3", "User1"))
	// TestGetGamesByAuthor tests the GetGamesByAuthor function.
	// It should return all Games with the given authorID.
	Games, err := GetGamesByAuthor("User1", &db)
	if err != nil {
		t.Errorf("Error getting Games by author: %v", err)
	}
	simpleAssert(t, 2, len(Games))
}

func TestCreateOrUpdateGame(t *testing.T) {
	db.Init("Test", "ID")
	// TestCreateOrUpdateGame tests the CreateOrUpdateGame function.
	CreateGame(createTestGame("Game1", "User1"), &db)
	CreateGame(createTestGame("Game2", "User2"), &db)
	CreateGame(createTestGame("Game3", "User3"), &db)
	// It should create a new Game if the user does not have one.
	simpleAssert(t, 1, len(db.DynamodbClient)) // all share a name
	simpleAssert(t, "Game1", db.DynamodbClient[0].(structs.Game).ID)
}

func TestUpdateOfCreateOrUpdateGame(t *testing.T) {
	db.Init("Test", "ID")
	// It should add a Game to the Game if the user already has one.
	CreateGame(createTestGame("Game1", "User1"), &db)
	updateGame := createTestGame("Game1", "User1")
	updateGame.Title = "NewTitle"
	UpdateGame("Game1", "User1", updateGame, &db)
	simpleAssert(t, 1, len(db.DynamodbClient))
	simpleAssert(t, "NewTitle", db.DynamodbClient[0].(structs.Game).Title)
}

// ----------------- Helper Functions -----------------
func simpleAssert[T comparable](t *testing.T, want, got T) {
	if got != want {
		t.Errorf("Expected %v got %v", want, got)
	}
}

func createTestGame(id string, userID string) structs.Game {
	return structs.Game{
		ID:          id,
		Title:       "TestTitle",
		Description: "TestDescription",
		Tags:        []string{"TestTag1", "TestTag2"},
		Price:       12.34,
		Published:   "TestPublished",
		Author:      "TestAuthor",
		AuthorID:    userID,
	}
}
