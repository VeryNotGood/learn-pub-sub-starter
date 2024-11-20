package main

import (
	"fmt"
	"log"
	"os"

	pubsub "github.com/bootdotdev/learn-pub-sub-starter/internal"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	conn, connErr := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if connErr != nil {
		log.Fatalf("could not connect to rabbitmq... error: %v", connErr)
	}

	defer conn.Close()
	fmt.Println("Connection Successful!")

	username, usernameErr := gamelogic.ClientWelcome()
	if usernameErr != nil {
		log.Fatalf("could not create user... error: %v", usernameErr)
	}

	_, _, pubErr := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, 1)
	if pubErr != nil {
		log.Fatalf("could not publish... error: %v", pubErr)
	}
	gs := gamelogic.NewGameState(username)

	for {
		userInput := gamelogic.GetInput()
		switch userInput[0] {
		case "":
			continue
		case "spawn":
			err := gs.CommandSpawn(userInput)
			if err != nil {
				fmt.Printf("Could not %s...%v\n", userInput, err)
			}
		case "move":
			_, err := gs.CommandMove(userInput)
			if err != nil {
				fmt.Printf("Could not %s...%v\n", userInput, err)
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			goto replKill
		default:
			fmt.Println("Unknown command. Please try again.")
		}
	}

replKill:
	os.Exit(1)

	// os.Exit(1)
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
}
