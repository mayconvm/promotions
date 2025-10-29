package main

import (
	"context"
	"fmt"
	"os"

	telegram "bot-telegram/src/pkg/telegram"

	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
	"github.com/joho/godotenv"
)

func main() {
	// Using ".env" file to load environment variables.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	client := telegram.ClientTelegram()

	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// authenticate user
		telegram.AuthTelegram(client, ctx)

		// It is only valid to use client while this function is not returned
		// and ctx is not cancelled.

		// Getting info about current user.
		self, err := client.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "call self")
		}

		name := self.FirstName
		if self.Username != "" {
			// Username is optional.
			name = fmt.Sprintf("%s (@%s)", name, self.Username)
		}
		fmt.Println("Current user:", name)

		raw := tg.NewClient(client)
		listChannels, err := telegram.ListChannelsFromFolders(ctx, raw, 4)
		if err != nil {
			return errors.Wrap(err, "list channels from folder")
		}

		for _, channel := range listChannels {
			if err := telegram.SearchProductInChannel(ctx, raw, channel, "SSD"); err != nil {
				return errors.Wrap(err, "search product in channel")
			}
		}

		// Return to close client connection and free up resources.
		return nil
	}); err != nil {
		panic(err)
	}
	// Client is closed.
}
