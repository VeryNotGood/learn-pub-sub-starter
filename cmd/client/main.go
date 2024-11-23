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

	ch, chErr := conn.Channel()
	if chErr != nil {
		fmt.Printf("Could not create channel...%v", chErr)
	}

	username, usernameErr := gamelogic.ClientWelcome()
	if usernameErr != nil {
		log.Fatalf("could not create user... error: %v", usernameErr)
	}
	fmt.Println("Creating game state...")
	gs := gamelogic.NewGameState(username)
	fmt.Println("Game state created...")
	fmt.Println("Handler created...")

	fmt.Println("Subbing to pause...")
	pauseSubErr := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		"direct",
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gs),
	)
	if pauseSubErr != nil {
		log.Fatalf("Error subscribing...%v", pauseSubErr)
	}
	fmt.Println("Subbed...")

	fmt.Println("Subbing to army moves...")
	armyMoveSubErr := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"topic",
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+"."+"*",
		pubsub.Transient,
		handlerMove(gs),
	)
	if armyMoveSubErr != nil {
		log.Fatalf("Error subscribing...%v", armyMoveSubErr)
	}
	fmt.Println("Subbed...")

	for {
		fmt.Print(">")
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
			move, err := gs.CommandMove(userInput)
			if err != nil {
				fmt.Printf("Could not %s...%v\n", userInput, err)
			}
			pubErr := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+username,
				move,
			)
			if pubErr != nil {
				fmt.Printf("Could not publish Playing State...%v\n", pubErr)
			}
			fmt.Println("Move successfully published...")
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

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(am gamelogic.ArmyMove) pubsub.AckType {
		moveOutcome := gs.HandleMove(am)
		switch moveOutcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}
