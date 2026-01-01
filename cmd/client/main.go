package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(r routing.PlayingState) {
    return func(r routing.PlayingState) {
        defer fmt.Print("> ")
        gs.HandlePause(r)
	}
}

func main() {
	// ------------------------------------------------------------------
	// Initialize structured logger (writes to stderr by default)
	// ------------------------------------------------------------------
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	logger.Info("Starting Peril client...")

	// ------------------------------------------------------------------
	// Connect to RabbitMQ broker
	// ------------------------------------------------------------------
	connectionURL := "amqp://guest:guest@localhost:5672/"
	amqpConnection, err := amqp.Dial(connectionURL)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ broker", "error", err.Error())
		return
	}
	defer amqpConnection.Close()
	logger.Info("Broker connection established successfully")

	// ------------------------------------------------------------------
	// Prompt client for username (via gamelogic)
	// ------------------------------------------------------------------
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		logger.Error("Failed to get username", "error", err.Error())
		return
	}
	logger.Info("Client username acquired", "username", username)

	// ------------------------------------------------------------------
	// Declare and bind a transient queue for this client
	// ------------------------------------------------------------------
	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
	_, _, err = pubsub.DeclareAndBind(
		amqpConnection,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.QueueTypeTransient,
	)
	if err != nil {
		logger.Error("Failed to declare and bind queue", "queue", queueName, "error", err.Error())
		return
	}
	logger.Info("Queue declared and bound successfully", "queue", queueName)

	// ------------------------------------------------------------------
	// Initialize new game state
	// ------------------------------------------------------------------
	gameState := gamelogic.NewGameState(username)
	logger.Info("Game state created", "player", username)

		pubsub.SubscribeJSON(
		amqpConnection,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.QueueTypeTransient,
		handlerPause(gameState),
	)

	// ------------------------------------------------------------------
	// REPL loop
	// ------------------------------------------------------------------
REPL:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		command := words[0]

		switch command {
		case "spawn":
			if len(words) != 3 {
				fmt.Println("Usage: spawn <location> <unit_type>")
				continue
			}
			err := gameState.CommandSpawn(words)
			if err != nil {
				fmt.Printf("Error: %v", err.Error())
				continue
			}
		case "move":
			if len(words) != 3 {
				logger.Error("Usage: move <destination> <unit_id>")
				continue
			}
			arm_mov, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Printf("Error: %v", err.Error())
				continue
			}
			fmt.Printf("Move successful! : %v", arm_mov)

		case "status":
			gameState.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			break REPL

		default:
			fmt.Println("Unknown command. Type 'help' for a list of commands.")
		}
	}
}
