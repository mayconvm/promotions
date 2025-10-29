package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
)

// Exemplos práticos de uso das funções de folders e channels

// Exemplo 1: Listar tudo - channels, grupos e folders
func listEverything(ctx context.Context, raw *tg.Client) error {
	fmt.Println("🚀 === LISTAGEM COMPLETA ===")

	// 1. Listar channels e grupos
	if err := listChannelsAndGroups(ctx, raw); err != nil {
		log.Printf("Erro ao listar channels/grupos: %v", err)
	}

	// 2. Listar folders
	if err := listFolders(ctx, raw); err != nil {
		log.Printf("Erro ao listar folders: %v", err)
	}

	// 3. Listar folders sugeridas
	if err := getSuggestedFolders(ctx, raw); err != nil {
		log.Printf("Erro ao listar folders sugeridas: %v", err)
	}

	return nil
}

// Exemplo 2: Análise de organização do Telegram
func analyzeOrganization(ctx context.Context, raw *tg.Client) error {
	fmt.Println("📊 === ANÁLISE DE ORGANIZAÇÃO ===")

	// Contar channels/grupos
	iter := query.GetDialogs(raw).Iter()

	var channelCount, supergroupCount, groupCount, userCount int

	for iter.Next(ctx) {
		elem := iter.Value()
		if elem.Deleted() {
			continue
		}

		switch peer := elem.Peer.(type) {
		case *tg.InputPeerChannel:
			inputChannel := &tg.InputChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: peer.AccessHash,
			}

			channels, err := raw.ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})
			if err != nil {
				continue
			}

			for _, chat := range channels.GetChats() {
				if channel, ok := chat.(*tg.Channel); ok {
					if channel.Broadcast {
						channelCount++
					} else {
						supergroupCount++
					}
				}
			}
		case *tg.InputPeerChat:
			groupCount++
		case *tg.InputPeerUser:
			userCount++
		}
	}

	// Contar folders
	dialogFilters, err := raw.MessagesGetDialogFilters(ctx)
	folderCount := 0
	if err == nil {
		folderCount = len(dialogFilters.GetFilters())
	}

	// Mostrar análise
	fmt.Printf("\n📈 ESTATÍSTICAS:\n")
	fmt.Printf("   📢 Channels: %d\n", channelCount)
	fmt.Printf("   👥 Supergroups: %d\n", supergroupCount)
	fmt.Printf("   💬 Groups legados: %d\n", groupCount)
	fmt.Printf("   👤 Conversas privadas: %d\n", userCount)
	fmt.Printf("   📁 Folders configuradas: %d\n", folderCount)

	totalChats := channelCount + supergroupCount + groupCount
	fmt.Printf("\n📋 TOTAL DE GRUPOS/CANAIS: %d\n", totalChats)

	if folderCount > 0 {
		avgChatsPerFolder := float64(totalChats) / float64(folderCount)
		fmt.Printf("📊 Média de chats por folder: %.1f\n", avgChatsPerFolder)
	}

	return nil
}

// Exemplo 3: Buscar channel/grupo específico por nome
func findChatByName(ctx context.Context, raw *tg.Client, searchName string) error {
	fmt.Printf("🔍 === BUSCANDO: %s ===\n", searchName)

	iter := query.GetDialogs(raw).Iter()
	found := false

	for iter.Next(ctx) {
		elem := iter.Value()
		if elem.Deleted() {
			continue
		}

		switch peer := elem.Peer.(type) {
		case *tg.InputPeerChannel:
			inputChannel := &tg.InputChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: peer.AccessHash,
			}

			channels, err := raw.ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})
			if err != nil {
				continue
			}

			for _, chat := range channels.GetChats() {
				if channel, ok := chat.(*tg.Channel); ok {
					if strings.Contains(strings.ToLower(channel.Title), strings.ToLower(searchName)) ||
						strings.Contains(strings.ToLower(channel.Username), strings.ToLower(searchName)) {
						found = true

						if channel.Broadcast {
							fmt.Printf("✅ ENCONTRADO - Channel: %s", channel.Title)
						} else {
							fmt.Printf("✅ ENCONTRADO - Supergroup: %s", channel.Title)
						}

						if channel.Username != "" {
							fmt.Printf(" (@%s)", channel.Username)
						}
						fmt.Printf(" - ID: %d\n", channel.ID)

						if count, ok := channel.GetParticipantsCount(); ok {
							fmt.Printf("   👥 Membros: %d\n", count)
						}
					}
				}
			}

		case *tg.InputPeerChat:
			chats, err := raw.MessagesGetChats(ctx, []int64{peer.ChatID})
			if err != nil {
				continue
			}

			for _, chat := range chats.GetChats() {
				if legacyChat, ok := chat.(*tg.Chat); ok {
					if strings.Contains(strings.ToLower(legacyChat.Title), strings.ToLower(searchName)) {
						found = true
						fmt.Printf("✅ ENCONTRADO - Group: %s - ID: %d\n", legacyChat.Title, legacyChat.ID)
						fmt.Printf("   👥 Membros: %d\n", legacyChat.ParticipantsCount)
					}
				}
			}
		}
	}

	if !found {
		fmt.Printf("❌ Nenhum chat encontrado com o nome '%s'\n", searchName)
	}

	return nil
}

// Exemplo 4: Organizar automaticamente em folders
func autoOrganizeFolders(ctx context.Context, raw *tg.Client) error {
	fmt.Println("🤖 === ORGANIZAÇÃO AUTOMÁTICA DE FOLDERS ===")

	// Este é apenas um exemplo conceitual
	// Na prática, você precisaria implementar a lógica específica

	fmt.Println("📝 Passos para organização automática:")
	fmt.Println("1. Analisar todos os chats")
	fmt.Println("2. Categorizar por tipo (work, personal, news, etc)")
	fmt.Println("3. Criar folders automaticamente")
	fmt.Println("4. Adicionar chats aos folders apropriados")

	// Exemplos de categorização:
	categories := map[string][]string{
		"Trabalho":   {"work", "empresa", "projeto", "equipe"},
		"Notícias":   {"news", "notícias", "jornal", "mídia"},
		"Família":    {"família", "family", "parente"},
		"Amigos":     {"amigos", "friends", "pessoal"},
		"Tecnologia": {"tech", "dev", "programming", "código"},
	}

	fmt.Println("\n🏷️  Categorias sugeridas:")
	for category, keywords := range categories {
		fmt.Printf("   📁 %s: %v\n", category, keywords)
	}

	fmt.Println("\n⚠️  Nota: Este é um exemplo conceitual.")
	fmt.Println("⚠️  Implemente a lógica específica conforme suas necessidades.")

	return nil
}

// Exemplo 5: Backup das configurações de folders
func backupFolders(ctx context.Context, raw *tg.Client) error {
	fmt.Println("💾 === BACKUP DE FOLDERS ===")

	dialogFilters, err := raw.MessagesGetDialogFilters(ctx)
	if err != nil {
		return fmt.Errorf("erro ao buscar folders para backup: %w", err)
	}

	filters := dialogFilters.GetFilters()

	fmt.Printf("📦 Fazendo backup de %d folders...\n", len(filters))

	for i, filter := range filters {
		switch f := filter.(type) {
		case *tg.DialogFilter:
			fmt.Printf("✅ [%d] Folder: %s (ID: %d)\n", i+1, f.Title.Text, f.ID)
			fmt.Printf("    - Emoji: %s\n", f.Emoticon)
			fmt.Printf("    - Contatos: %t\n", f.Contacts)
			fmt.Printf("    - Grupos: %t\n", f.Groups)
			fmt.Printf("    - Canais: %t\n", f.Broadcasts)
			fmt.Printf("    - Chats incluídos: %d\n", len(f.IncludePeers))
			fmt.Printf("    - Chats excluídos: %d\n", len(f.ExcludePeers))

		case *tg.DialogFilterChatlist:
			fmt.Printf("✅ [%d] Folder Compartilhada: %s (ID: %d)\n", i+1, f.Title.Text, f.ID)
			if color, ok := f.GetColor(); ok {
				fmt.Printf("    - Cor: %s\n", getColorName(color))
			}
			fmt.Printf("    - Chats incluídos: %d\n", len(f.IncludePeers))
		}
	}

	fmt.Println("\n✅ Backup concluído!")
	fmt.Println("💡 Dica: Salve essas informações em um arquivo para restaurar depois.")

	return nil
}

// As funções de strings são importadas do package "strings" do Go
