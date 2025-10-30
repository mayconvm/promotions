package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"bot-telegram/src/internal/domain"
	supabase "bot-telegram/src/pkg/supabase"
	"bot-telegram/src/pkg/telegram"

	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
	"github.com/joho/godotenv"
)

func main() {
	// Using ".env" file to load environment variables.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	sessionProductsToSearch := make(chan []domain.Product)

	client := telegram.ClientTelegram()
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// authenticate user
		telegram.AuthTelegram(client, ctx)

		raw := tg.NewClient(client)
		listChannels, err := telegram.ListChannelsFromFolders(ctx, raw, 4)
		if err != nil {
			return errors.Wrap(err, "list channels from folder")
		}

		go searchProductsInChannels(sessionProductsToSearch, listChannels, ctx, raw)

		if err := ListAllProducts(sessionProductsToSearch); err != nil {
			panic(err)
		}

		time.Sleep(10 * time.Second)
		close(sessionProductsToSearch)

		// Return to close client connection and free up resources.
		return nil
	}); err != nil {
		panic(err)
	}

	// inicia a connecção com o telegram e registra goroutines para buscar os produtos no telegram
	// 	criar um canal para executar a busca
	// Iniciar o cronjob e registrar as sessions
	// 	a cada execução, publicar os produtos a serem pesquisados nos canais

	// Client is closed.
}

func searchProductsInChannels(list chan []domain.Product, listChannels []*tg.InputPeerChannel, ctx context.Context, raw *tg.Client) {

	itemsProducts := <-list

	for _, channel := range listChannels {
		if err := telegram.SearchProductInChannel(ctx, raw, channel, itemsProducts[1].Name); err != nil {
			fmt.Print(errors.Wrap(err, "search product in channel"))
		}
	}
}

func ListAllProducts(channel chan []domain.Product) error {
	client, err := supabase.NewClient()
	if err != nil {
		return errors.Wrap(err, "error connect supabase")
	}

	listProducts, err := supabase.GetAllProducts(client, &domain.Session{SessionId: "bac"})
	if err != nil {
		return errors.Wrap(err, "error connect supabase")
	}

	channel <- listProducts

	for _, product := range listProducts {
		println("Product:", product.ProductID, product.Title)
	}

	return nil
}
