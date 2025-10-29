package telegram

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

func ClientTelegram() *telegram.Client {

	appID := os.Getenv("TELEGRAM_APP_ID")
	appHash := os.Getenv("TELEGRAM_APP_HASH")

	// Setting up session storage.
	// This is needed to reuse session and not login every time.
	sessionDir := filepath.Join("session", os.Getenv("TELEGRAM_PHONE"))
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		panic(err)
	}

	// logFilePath := filepath.Join(sessionDir, "log.jsonl")
	fmt.Printf("Storing session in %s\n", sessionDir)

	// So, we are storing session information in current directory, under subdirectory "session/phone_hash"
	sessionStorage := &telegram.FileSessionStorage{
		Path: filepath.Join(sessionDir, "session.json"),
	}

	options := telegram.Options{
		// Logger:         lg,              // Passing logger for observability.
		SessionStorage: sessionStorage, // Setting up session sessionStorage to store auth data.
		// UpdateHandler:  updatesRecovery, // Setting up handler for updates from server.
	}

	// https://core.telegram.org/api/obtaining_api_id
	appIDInt, err := strconv.Atoi(appID)
	if err != nil {
		panic(err)
	}

	client := telegram.NewClient(appIDInt, appHash, options)

	return client
}

func AuthTelegram(client *telegram.Client, ctx context.Context) {

	phone := os.Getenv("TELEGRAM_PHONE")
	codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		// NB: Use "golang.org/x/crypto/ssh/terminal" to prompt password.
		fmt.Print("Enter code: ")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(code), nil
	}

	flow := auth.NewFlow(
		auth.CodeOnly(phone, auth.CodeAuthenticatorFunc(codePrompt)),
		// auth.Constant(phone, password, auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{},
	)

	// Perform auth if no session is available.
	if err := client.Auth().IfNecessary(ctx, flow); err != nil {
		panic(err)
	}
}

func SearchProductInChannel(ctx context.Context, raw *tg.Client, targetPeer *tg.InputPeerChannel, productName string) error {
	fmt.Printf("\n=== Searching for product: %s - %d ===\n", productName, targetPeer.ChannelID)
	// Perform the search
	results, err := raw.MessagesSearch(ctx, &tg.MessagesSearchRequest{
		Peer:    targetPeer,
		Q:       productName,
		Filter:  &tg.InputMessagesFilterEmpty{},             // Necessário para buscar todos os tipos de mensagem
		Limit:   2,                                          // Limitar resultados
		MinDate: int(time.Now().Add(-2 * time.Hour).Unix()), // últimas duas horas
	})
	if err != nil {
		return fmt.Errorf("erro ao buscar produto no canal: %w", err)
	}

	// Process the search results
	switch msgs := results.(type) {
	case *tg.MessagesMessages:
		fmt.Printf("✅ Encontrado %d mensagens:\n", len(msgs.Messages))
		for i, msg := range msgs.Messages {
			if message, ok := msg.(*tg.Message); ok {
				fmt.Printf("🔍 [%d] %s\n", i+1, message.Message)
				if message.Date > 0 {
					fmt.Printf("    📅 Data: %d\n", message.Date)
				}
				fmt.Println()
			}
		}
	case *tg.MessagesMessagesSlice:
		fmt.Printf("✅ Encontrado %d de %d mensagens:\n", len(msgs.Messages), msgs.Count)
		for i, msg := range msgs.Messages {
			if message, ok := msg.(*tg.Message); ok {
				fmt.Printf("🔍 [%d] %s\n", i+1, message.Message)
				if message.Date > 0 {
					fmt.Printf("    � Data: %d\n", message.Date)
				}
				fmt.Println()
			}
		}
	case *tg.MessagesChannelMessages:
		fmt.Printf("✅ Encontrado %d de %d mensagens no canal:\n", len(msgs.Messages), msgs.Count)
		for i, msg := range msgs.Messages {
			if message, ok := msg.(*tg.Message); ok {
				fmt.Printf("🔍 [%d] %s\n", i+1, message.Message)
				if message.Date > 0 {
					fmt.Printf("    📅 Data: %d\n", message.Date)
				}
				fmt.Println()
			}
		}
	default:
		fmt.Printf("❌ Tipo de resultado desconhecido: %T\n", results)
	}

	return nil
}

func ListChannelsFromFolders(ctx context.Context, raw *tg.Client, folderID int) ([]*tg.InputPeerChannel, error) {
	// Obter filtros de diálogo (pastas)
	dialogFilters, err := raw.MessagesGetDialogFilters(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar folders: %w", err)
	}

	// Encontrar a pasta específica
	var folder *tg.DialogFilter
	filters := dialogFilters.GetFilters()
	for _, filter := range filters {
		if f, ok := filter.(*tg.DialogFilter); ok && f.ID == folderID {
			folder = f
			break
		}
	}

	if folder == nil {
		return nil, fmt.Errorf("❌ Pasta com ID %d não encontrada.", folderID)
	}

	// Listar chats incluídos na pasta
	includedChats := folder.GetIncludePeers()
	if len(includedChats) == 0 {
		return nil, fmt.Errorf("❌ Pasta vazia.\n")
	}

	var includedChannels []*tg.InputPeerChannel
	// var includedGroups []*tg.InputPeerChat

	for _, chat := range includedChats {
		switch peer := chat.(type) {
		case *tg.InputPeerChannel:
			includedChannels = append(includedChannels, peer)
			// case *tg.InputPeerChat:
			// 	includedGroups = append(includedGroups, peer)
		}
	}

	return includedChannels, nil
}
