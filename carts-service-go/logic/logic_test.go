package logic

import (
	"testing"

	database "github.com/Draupniyr/carts-service/mockdb"
	"github.com/Draupniyr/carts-service/structs"
)

var db database.Database

func TestGetAllCarts(t *testing.T) {
	//setup
	db.Init("Test", "ID")
	db.DynamodbClient = append(db.DynamodbClient, CreateTestCart("TestID1", createTestGame("Game1")))
	db.DynamodbClient = append(db.DynamodbClient, CreateTestCart("TestID2", createTestGame("Game1")))
	db.DynamodbClient = append(db.DynamodbClient, CreateTestCart("TestID3", createTestGame("Game1")))

	// TestGetAllCarts tests the GetAllCarts function.
	carts, err := GetAllCarts(&db)
	if err != nil {
		t.Errorf("Error getting all carts: %v", err)
	}

	// It should return all carts in the database.
	if len(carts) != 3 {
		simpleAssert(t, 3, len(carts))
	}
}

func TestGetCart(t *testing.T) {
	db.Init("Test", "ID")
	db.DynamodbClient = append(db.DynamodbClient, CreateTestCart("TestID1", createTestGame("Game1")))
	db.DynamodbClient = append(db.DynamodbClient, CreateTestCart("TestID2", createTestGame("Game1")))
	db.DynamodbClient = append(db.DynamodbClient, CreateTestCart("TestID3", createTestGame("Game1")))
	// TestGetCart tests the GetCart function.
	// It should return a cart with the given userID.
	cart, err := GetCart("TestID2", &db)
	if err != nil {
		t.Errorf("Error getting cart: %v", err)
	}
	simpleAssert(t, "TestID2", cart.UserID)
}

func TestCreateOfCreateOrUpdateCart(t *testing.T) {
	db.Init("Test", "ID")
	// TestCreateOrUpdateCart tests the CreateOrUpdateCart function.
	CreateORUpdateCart("TestID1", createTestGame("Game1"), &db)
	CreateORUpdateCart("TestID2", createTestGame("Game2"), &db)
	CreateORUpdateCart("TestID3", createTestGame("Game3"), &db)
	// It should create a new cart if the user does not have one.
	simpleAssert(t, 3, len(db.DynamodbClient))
	simpleAssert(t, "Game1", db.DynamodbClient[0].(structs.Cart).Games[0].ID)
	simpleAssert(t, "Game2", db.DynamodbClient[1].(structs.Cart).Games[0].ID)
	simpleAssert(t, "Game3", db.DynamodbClient[2].(structs.Cart).Games[0].ID)
}

func TestUpdateOfCreateOrUpdateCart(t *testing.T) {
	db.Init("Test", "ID")
	// It should add a game to the cart if the user already has one.
	CreateORUpdateCart("TestID1", createTestGame("Game1"), &db)
	updateGame := createTestGame("Game2")

	CreateORUpdateCart("TestID1", updateGame, &db)

	simpleAssert(t, 1, len(db.DynamodbClient))
	simpleAssert(t, 2, len(db.DynamodbClient[0].(structs.Cart).Games))
}

func TestUpdateOfCreateOrUpdateCarttwo(t *testing.T) {
	db.Init("Test", "ID")
	// It should add a game to the cart if the user already has one.
	CreateORUpdateCart("TestID1", createTestGame("Game1"), &db)

	CreateORUpdateCart("TestID1", createTestGame("Game1"), &db)

	simpleAssert(t, 1, len(db.DynamodbClient))
	simpleAssert(t, 0, len(db.DynamodbClient[0].(structs.Cart).Games))
}

func TestDeletCart(t *testing.T) {
	db.Init("Test", "ID")
	// TestDeleteGameFromCart tests the DeleteGameFromCart function.
	db.DynamodbClient = append(db.DynamodbClient, CreateTestCart("TestID1", createTestGame("Game1")))
	db.DynamodbClient = append(db.DynamodbClient, CreateTestCart("TestID2", createTestGame("Game1")))
	db.DynamodbClient = append(db.DynamodbClient, CreateTestCart("TestID3", createTestGame("Game1")))

	DeleteCart("TestID2", &db)
	// It should remove the cart from the database.
	simpleAssert(t, 2, len(db.DynamodbClient))
}

// ----------------- Helper Functions -----------------
func simpleAssert[T comparable](t *testing.T, want, got T) {
	if got != want {
		t.Errorf("Expected %v got %v", want, got)
	}
}

func CreateTestCart(userID string, game structs.Game) structs.Cart {

	cart := structs.CreateCartRequest{
		UserID: userID,
		Game:   &game,
	}
	return cart.CreateCartRequestToCart()
}

func createTestGame(id string) structs.Game {
	return structs.Game{
		ID:          id,
		Title:       "TestTitle",
		Description: "TestDescription",
		Tags:        []string{"TestTag1", "TestTag2"},
		Price:       12.34,
		Published:   "TestPublished",
		Author:      "TestAuthor",
		AuthorID:    "TestAuthorID",
	}
}
