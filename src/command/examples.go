package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
)

// Exemplos de como listar channels e grupos usando gotd/td

// listChannelsAndGroupsSimple - Versão mais simples usando apenas os dados dos diálogos
func listChannelsAndGroupsSimple(ctx context.Context, raw *tg.Client) error {
	fmt.Println("\n=== Versão Simples - Listando Channels e Grupos ===")

	iter := query.GetDialogs(raw).Iter()

	for iter.Next(ctx) {
		elem := iter.Value()

		if elem.Deleted() {
			continue
		}

		switch peer := elem.Peer.(type) {
		case *tg.InputPeerChannel:
			fmt.Printf("📢/👥 Channel/Supergroup - ID: %d (AccessHash: %d)\n",
				peer.ChannelID, peer.AccessHash)
		case *tg.InputPeerChat:
			fmt.Printf("💬 Group - ID: %d\n", peer.ChatID)
		}
	}

	return iter.Err()
}

// listChannelsOnly - Lista apenas channels (broadcasts)
func listChannelsOnly(ctx context.Context, raw *tg.Client) error {
	fmt.Println("\n=== Apenas Channels (Broadcasts) ===")

	iter := query.GetDialogs(raw).Iter()

	for iter.Next(ctx) {
		elem := iter.Value()

		if elem.Deleted() {
			continue
		}

		if peer, ok := elem.Peer.(*tg.InputPeerChannel); ok {
			inputChannel := &tg.InputChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: peer.AccessHash,
			}

			channels, err := raw.ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})
			if err != nil {
				log.Printf("Erro ao buscar channel %d: %v", peer.ChannelID, err)
				continue
			}

			for _, chat := range channels.GetChats() {
				if channel, ok := chat.(*tg.Channel); ok && channel.Broadcast {
					fmt.Printf("📢 %s", channel.Title)
					if channel.Username != "" {
						fmt.Printf(" (@%s)", channel.Username)
					}
					fmt.Printf(" - ID: %d\n", channel.ID)
				}
			}
		}
	}

	return iter.Err()
}

// listSupergroupsOnly - Lista apenas supergroups
func listSupergroupsOnly(ctx context.Context, raw *tg.Client) error {
	fmt.Println("\n=== Apenas Supergroups ===")

	iter := query.GetDialogs(raw).Iter()

	for iter.Next(ctx) {
		elem := iter.Value()

		if elem.Deleted() {
			continue
		}

		if peer, ok := elem.Peer.(*tg.InputPeerChannel); ok {
			inputChannel := &tg.InputChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: peer.AccessHash,
			}

			channels, err := raw.ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})
			if err != nil {
				log.Printf("Erro ao buscar channel %d: %v", peer.ChannelID, err)
				continue
			}

			for _, chat := range channels.GetChats() {
				if channel, ok := chat.(*tg.Channel); ok && !channel.Broadcast {
					fmt.Printf("👥 %s", channel.Title)
					if channel.Username != "" {
						fmt.Printf(" (@%s)", channel.Username)
					}
					fmt.Printf(" - ID: %d\n", channel.ID)
				}
			}
		}
	}

	return iter.Err()
}

// listLegacyGroupsOnly - Lista apenas groups legados (chats antigos)
func listLegacyGroupsOnly(ctx context.Context, raw *tg.Client) error {
	fmt.Println("\n=== Apenas Groups Legados ===")

	iter := query.GetDialogs(raw).Iter()

	for iter.Next(ctx) {
		elem := iter.Value()

		if elem.Deleted() {
			continue
		}

		if peer, ok := elem.Peer.(*tg.InputPeerChat); ok {
			chats, err := raw.MessagesGetChats(ctx, []int64{peer.ChatID})
			if err != nil {
				log.Printf("Erro ao buscar chat %d: %v", peer.ChatID, err)
				continue
			}

			for _, chat := range chats.GetChats() {
				if legacyChat, ok := chat.(*tg.Chat); ok {
					fmt.Printf("💬 %s - ID: %d (Membros: %d)\n",
						legacyChat.Title, legacyChat.ID, legacyChat.ParticipantsCount)
				}
			}
		}
	}

	return iter.Err()
}

// getChannelInfo - Obtém informações detalhadas de um channel específico
func getChannelInfo(ctx context.Context, raw *tg.Client, channelID int64, accessHash int64) error {
	inputChannel := &tg.InputChannel{
		ChannelID:  channelID,
		AccessHash: accessHash,
	}

	// Obter informações básicas
	channels, err := raw.ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})
	if err != nil {
		return fmt.Errorf("erro ao buscar channel: %w", err)
	}

	for _, chat := range channels.GetChats() {
		if channel, ok := chat.(*tg.Channel); ok {
			fmt.Printf("\n=== Informações do Channel ===\n")
			fmt.Printf("Nome: %s\n", channel.Title)
			fmt.Printf("ID: %d\n", channel.ID)
			if channel.Username != "" {
				fmt.Printf("Username: @%s\n", channel.Username)
			}
			fmt.Printf("É Broadcast: %t\n", channel.Broadcast)
			fmt.Printf("É Megagroup: %t\n", channel.Megagroup)
			fmt.Printf("Verificado: %t\n", channel.Verified)
			fmt.Printf("Restrito: %t\n", channel.Restricted)

			if count, ok := channel.GetParticipantsCount(); ok {
				fmt.Printf("Membros: %d\n", count)
			}

			// Obter informações completas (opcional)
			fullChannel, err := raw.ChannelsGetFullChannel(ctx, inputChannel)
			if err == nil {
				if full, ok := fullChannel.FullChat.(*tg.ChannelFull); ok {
					fmt.Printf("Descrição: %s\n", full.About)
					fmt.Printf("Participantes: %d\n", full.ParticipantsCount)
				}
			}
		}
	}

	return nil
}

// Exemplos adicionais para trabalhar com folders

// listFoldersSimple - Versão simples para listar apenas nomes das pastas
func listFoldersSimple(ctx context.Context, raw *tg.Client) error {
	fmt.Println("\n=== Versão Simples - Listando Folders ===")

	dialogFilters, err := raw.MessagesGetDialogFilters(ctx)
	if err != nil {
		return fmt.Errorf("erro ao buscar folders: %w", err)
	}

	filters := dialogFilters.GetFilters()

	for i, filter := range filters {
		switch f := filter.(type) {
		case *tg.DialogFilter:
			fmt.Printf("%d. 📁 %s\n", i+1, f.Title.Text)
		case *tg.DialogFilterChatlist:
			fmt.Printf("%d. 🔗 %s (Compartilhada)\n", i+1, f.Title.Text)
		case *tg.DialogFilterDefault:
			fmt.Printf("%d. ⭐ Todos os Chats (Padrão)\n", i+1)
		}
	}

	return nil
}

// getFolderDetails - Obtém informações detalhadas de uma pasta específica
func getFolderDetails(ctx context.Context, raw *tg.Client, folderID int) error {
	fmt.Printf("\n=== Informações da Pasta ID: %d ===\n", folderID)

	dialogFilters, err := raw.MessagesGetDialogFilters(ctx)
	if err != nil {
		return fmt.Errorf("erro ao buscar folders: %w", err)
	}

	filters := dialogFilters.GetFilters()

	for _, filter := range filters {
		var targetID int

		switch f := filter.(type) {
		case *tg.DialogFilter:
			targetID = f.ID
			if targetID == folderID {
				fmt.Printf("📁 Nome: %s\n", f.Title.Text)
				if f.Emoticon != "" {
					fmt.Printf("😀 Emoji: %s\n", f.Emoticon)
				}

				fmt.Printf("📊 Estatísticas:\n")
				fmt.Printf("   - Chats incluídos: %d\n", len(f.IncludePeers))
				fmt.Printf("   - Chats excluídos: %d\n", len(f.ExcludePeers))
				fmt.Printf("   - Chats fixados: %d\n", len(f.PinnedPeers))

				return nil
			}
		case *tg.DialogFilterChatlist:
			targetID = f.ID
			if targetID == folderID {
				fmt.Printf("🔗 Nome: %s (Compartilhada)\n", f.Title.Text)
				if f.Emoticon != "" {
					fmt.Printf("😀 Emoji: %s\n", f.Emoticon)
				}

				if color, ok := f.GetColor(); ok {
					fmt.Printf("🎨 Cor: %s\n", getColorName(color))
				}

				fmt.Printf("📊 Estatísticas:\n")
				fmt.Printf("   - Chats incluídos: %d\n", len(f.IncludePeers))
				fmt.Printf("   - Chats fixados: %d\n", len(f.PinnedPeers))
				fmt.Printf("   - Tem convites próprios: %t\n", f.HasMyInvites)

				return nil
			}
		}
	}

	return fmt.Errorf("pasta com ID %d não encontrada", folderID)
}

// getSuggestedFolders - Lista pastas sugeridas pelo Telegram
func getSuggestedFolders(ctx context.Context, raw *tg.Client) error {
	fmt.Println("\n=== Pastas Sugeridas pelo Telegram ===")

	suggested, err := raw.MessagesGetSuggestedDialogFilters(ctx)
	if err != nil {
		return fmt.Errorf("erro ao buscar pastas sugeridas: %w", err)
	}

	if len(suggested) == 0 {
		fmt.Println("📭 Nenhuma pasta sugerida disponível")
		return nil
	}

	for i, suggestion := range suggested {
		filter := suggestion.Filter

		switch f := filter.(type) {
		case *tg.DialogFilter:
			fmt.Printf("%d. 📁 %s", i+1, f.Title.Text)
			if f.Emoticon != "" {
				fmt.Printf(" %s", f.Emoticon)
			}
			fmt.Printf("\n   📝 %s\n", suggestion.Description)
		case *tg.DialogFilterChatlist:
			fmt.Printf("%d. 🔗 %s", i+1, f.Title.Text)
			if f.Emoticon != "" {
				fmt.Printf(" %s", f.Emoticon)
			}
			fmt.Printf(" (Compartilhada)\n")
			fmt.Printf("   📝 %s\n", suggestion.Description)
		}
	}

	return nil
}

// createFolder - Exemplo de como criar uma nova pasta (dialog filter)
func createFolder(ctx context.Context, raw *tg.Client, title, emoticon string) error {
	fmt.Printf("\n=== Criando Nova Pasta: %s ===\n", title)

	// Buscar folders existentes para obter próximo ID
	dialogFilters, err := raw.MessagesGetDialogFilters(ctx)
	if err != nil {
		return fmt.Errorf("erro ao buscar folders existentes: %w", err)
	}

	// Encontrar próximo ID disponível
	maxID := 0
	for _, filter := range dialogFilters.GetFilters() {
		var id int
		switch f := filter.(type) {
		case *tg.DialogFilter:
			id = f.ID
		case *tg.DialogFilterChatlist:
			id = f.ID
		}
		if id > maxID {
			maxID = id
		}
	}
	nextID := maxID + 1

	// Criar nova pasta
	newFilter := &tg.DialogFilter{
		ID:       nextID,
		Title:    tg.TextWithEntities{Text: title},
		Emoticon: emoticon,
		// Configurações padrão - incluir todos os tipos
		Contacts:    true,
		NonContacts: true,
		Groups:      true,
		Broadcasts:  true,
		Bots:        false,
		// Peers vazios inicialmente
		IncludePeers: make([]tg.InputPeerClass, 0),
		ExcludePeers: make([]tg.InputPeerClass, 0),
		PinnedPeers:  make([]tg.InputPeerClass, 0),
	}

	// Atualizar pasta no servidor
	success, err := raw.MessagesUpdateDialogFilter(ctx, &tg.MessagesUpdateDialogFilterRequest{
		ID:     nextID,
		Filter: newFilter,
	})

	if err != nil {
		return fmt.Errorf("erro ao criar pasta: %w", err)
	}

	if success {
		fmt.Printf("✅ Pasta '%s' criada com sucesso! (ID: %d)\n", title, nextID)
	} else {
		fmt.Printf("❌ Falha ao criar pasta '%s'\n", title)
	}

	return nil
}

// deleteFolder - Exemplo de como deletar uma pasta
func deleteFolder(ctx context.Context, raw *tg.Client, folderID int) error {
	fmt.Printf("\n=== Deletando Pasta ID: %d ===\n", folderID)

	// Para deletar, chama updateDialogFilter sem o campo filter
	success, err := raw.MessagesUpdateDialogFilter(ctx, &tg.MessagesUpdateDialogFilterRequest{
		ID: folderID,
		// Filter omitido intencionalmente para deletar
	})

	if err != nil {
		return fmt.Errorf("erro ao deletar pasta: %w", err)
	}

	if success {
		fmt.Printf("✅ Pasta ID %d deletada com sucesso!\n", folderID)
	} else {
		fmt.Printf("❌ Falha ao deletar pasta ID %d\n", folderID)
	}

	return nil
}

// reorderFolders - Exemplo de como reordenar pastas
func reorderFolders(ctx context.Context, raw *tg.Client, order []int) error {
	fmt.Println("\n=== Reordenando Pastas ===")

	success, err := raw.MessagesUpdateDialogFiltersOrder(ctx, order)

	if err != nil {
		return fmt.Errorf("erro ao reordenar pastas: %w", err)
	}

	if success {
		fmt.Printf("✅ Pastas reordenadas com sucesso! Nova ordem: %v\n", order)
	} else {
		fmt.Println("❌ Falha ao reordenar pastas")
	}

	return nil
}
