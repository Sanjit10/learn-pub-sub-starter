package main

import (
	"log/slog"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func publishPauseState(ch *amqp.Channel, key string, paused bool, logger *slog.Logger) {
	data := routing.PlayingState{IsPaused: paused}
	err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, key, data)
	if err != nil {
		logger.Error("Error while publishing message", "error", err.Error())
	}
}

func RunInputLoop(channel *amqp.Channel, logger *slog.Logger) {
Loop:
	for {
		user_input := gamelogic.GetInput()
		if len(user_input) == 0 {
			continue
		}
		first_word := user_input[0]

		switch first_word {
		case "pause":
			logger.Info("Publishing message 'pause'")
			publishPauseState(channel, routing.PauseKey, true, logger)
		case "resume":
			logger.Info("Publishing message 'resume'")
			publishPauseState(channel, routing.PauseKey, false, logger)
		case "quit":
			logger.Info("Quitting input loop...")
			break Loop
		default:
			logger.Info("Undefined command received.")
		}
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Starting Peril server...")
	gamelogic.PrintServerHelp()

	connection_url := "amqp://guest:guest@localhost:5672/"
	amqp_connection, err := amqp.Dial(connection_url)
	if err != nil {
		logger.Error("Couldn't connect to Broker.")
		return
	}
	defer amqp_connection.Close()
	logger.Info("Broker Connection Succesful. ")

	// Channels
	channel, err := amqp_connection.Channel()
	if err != nil {
		logger.Error("Error while establishing channel.")
		return
	}

	//Topic queue binding
	_, _, err = pubsub.DeclareAndBind(
		amqp_connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"game_logs.*",
		pubsub.QueueTypeDurable,
	)
	if err != nil {
		logger.Error("Failed to declare and bind queue", "queue", routing.GameLogSlug, "error", err.Error())
		return
	}
	logger.Info("Queue declared and bound successfully", "queue", routing.GameLogSlug)

	RunInputLoop(channel, logger)
}
