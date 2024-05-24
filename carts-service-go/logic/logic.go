package logic

import (
	"log"

	"encoding/json"

	database "github.com/Draupniyr/carts-service/database"

	kafka "github.com/Draupniyr/carts-service/kafka"
	structs "github.com/Draupniyr/carts-service/structs"
)

// ----------------- Carts -----------------
func GetCart(userID string, db database.DatabaseFunctionality) (structs.Cart, error) {
	cart := structs.Cart{}
	err := db.GetFilter(userID, "UserID", &cart)
	if err != nil {
		log.Println("Error getting cart:", err)
		return cart, err
	}
	return cart, nil
}

func GetAllCarts(db database.DatabaseFunctionality) ([]structs.Cart, error) {
	carts := []structs.Cart{}
	err := db.GetAll(&carts)
	if err != nil {
		log.Println("Error getting all carts:", err)
		return nil, err
	}
	return carts, nil
}

func CreateORUpdateCart(userId string, game structs.Game, db database.DatabaseFunctionality) error {
	Cart := structs.Cart{}
	err := db.GetFilter(userId, "UserID", &Cart)
	if err != nil {
		cartRequest := structs.CreateCartRequest{
			UserID: userId,
			Game:   &game,
		}
		Cart = cartRequest.CreateCartRequestToCart()
		err = db.CreateOrUpdate(Cart)
		if err != nil {
			log.Println("Error creating cart:", err)
			return err
		}
		return nil
	} else {
		err = AddOrRemoveFromCart(userId, game, db)
		if err != nil {
			log.Println("Error adding or removing game from cart:", err)
			return err
		}
		return nil
	}
}

func AddOrRemoveFromCart(userID string, gameToAddOrRemove structs.Game, db database.DatabaseFunctionality) error {
	cartOG := structs.Cart{}
	err := db.GetFilter(userID, "UserID", &cartOG)
	if err != nil {
		log.Println("Error getting cart:", err)
		return err
	}

	newgames := []structs.Game{}
	contains := false
	for _, game := range cartOG.Games {
		if game.ID == gameToAddOrRemove.ID {
			contains = true
		}
	}
	if contains {
		for _, game := range cartOG.Games {
			if game.ID != gameToAddOrRemove.ID {
				newgames = append(newgames, game)
			}
		}
		cartOG.Games = newgames
	} else {
		cartOG.Games = append(cartOG.Games, gameToAddOrRemove)
	}

	err = db.CreateOrUpdate(cartOG)
	if err != nil {
		log.Println("Error adding or removing game from cart:", err)
		return err
	}
	return nil
}

func DeleteCart(UserID string, db database.DatabaseFunctionality) error {

	err := db.DeleteFilter(UserID, "UserID")
	if err != nil {
		log.Println("Error deleting cart:", err)
		return err
	}
	return nil
}

func DeleteAll(db database.DatabaseFunctionality) error {
	err := db.DeleteAll()
	if err != nil {
		log.Println("Error deleting all carts:", err)
		return err
	}
	return nil
}

func Checkout(userID string, db database.DatabaseFunctionality, kafka kafka.KafkaProducer) error {

	cart := structs.Cart{}
	err := db.GetFilter(userID, "UserID", &cart)
	if err != nil {
		log.Println("Error getting cart:", err)
		return err
	}

	// turn cart into a byte array
	cartJson, err := json.Marshal(cart)
	if err != nil {
		log.Println("Error marshaling cart:", err)
		return err
	}
	cartByte := []byte(cartJson)
	// turn gamejson into a byte array
	err = kafka.PushCommentToQueue("checkout", userID, cartByte)
	if err != nil {
		log.Println("Error pushing cart to kafka:", err)
		return err
	}
	err = db.Delete(cart.ID)
	if err != nil {
		log.Println("Error deleting cart:", err)
		return err
	}
	return nil
}
