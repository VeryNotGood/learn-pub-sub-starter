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
	fmt.Println("Starting Peril server...")

	conn, connErr := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if connErr != nil {
		log.Fatalf("could not connect to rabbitmq...%v", connErr)
	}
	ch, chErr := conn.Channel()
	if chErr != nil {
		log.Fatalf("could not create channel...%v", connErr)
	}
	_, _, declareBindErr := pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+"."+"*", 2)
	if declareBindErr != nil {
		log.Fatalf("could not create topic...%v", declareBindErr)
	}

	defer conn.Close()
	fmt.Println("Connection Successful!")

	gamelogic.PrintServerHelp()

	for {
		userInput := gamelogic.GetInput()
		switch userInput[0] {
		case "":
			continue
		case "pause":
			fmt.Println("Sending pause message...")
			playingState := routing.PlayingState{
				IsPaused: true,
			}
			pubErr := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, playingState)
			if pubErr != nil {
				fmt.Printf("Could not publish Playing State...%v\n", pubErr)
			}
		case "resume":
			fmt.Println("Sending resume message...")
			playingState := routing.PlayingState{
				IsPaused: false,
			}
			pubErr := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, playingState)
			if pubErr != nil {
				fmt.Printf("Could not publish Playing State...%v\n", pubErr)
			}
		case "quit":
			fmt.Println("Quitting...")
			goto replKill
		default:
			fmt.Println("Unknown command. Try again.")
		}
	}

replKill:
	os.Exit(1)
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
}
