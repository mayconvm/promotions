package examples

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
)

// listChannelsAndGroups lista todos os channels e grupos do usuário
func listChannelsAndGroups(ctx context.Context, raw *tg.Client) error {
	fmt.Println("\n=== Listando Channels e Grupos ===")

	// Usar query builder para obter diálogos
	iter := query.GetDialogs(raw).Iter()

	channelCount := 0
	groupCount := 0

	// Iterar sobre todos os diálogos
	for iter.Next(ctx) {
		elem := iter.Value()

		// Pular diálogos deletados
		if elem.Deleted() {
			continue
		}

		// Verificar o tipo de peer
		switch peer := elem.Peer.(type) {
		case *tg.InputPeerChannel:
			// É um channel ou supergroup
			channelID := peer.ChannelID

			// Criar InputChannel correto
			inputChannel := &tg.InputChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: peer.AccessHash,
			}

			// Buscar informações do channel
			channels, err := raw.ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})
			if err != nil {
				fmt.Printf("Erro ao buscar channel %d: %v\n", channelID, err)
				continue
			}

			// Processar channels retornados
			for _, chat := range channels.GetChats() {
				if channel, ok := chat.(*tg.Channel); ok {
					if channel.Broadcast {
						// É um channel (broadcast)
						channelCount++
						fmt.Printf("📢 Channel: %s", channel.Title)
						if channel.Username != "" {
							fmt.Printf(" (@%s)", channel.Username)
						}
						fmt.Printf(" - ID: %d\n", channel.ID)

						if count, ok := channel.GetParticipantsCount(); ok && count > 0 {
							fmt.Printf("   👥 Membros: %d\n", count)
						}
					} else {
						// É um supergroup
						groupCount++
						fmt.Printf("👥 Supergroup: %s", channel.Title)
						if channel.Username != "" {
							fmt.Printf(" (@%s)", channel.Username)
						}
						fmt.Printf(" - ID: %d\n", channel.ID)

						if count, ok := channel.GetParticipantsCount(); ok && count > 0 {
							fmt.Printf("   👥 Membros: %d\n", count)
						}
					}
				}
			}

		case *tg.InputPeerChat:
			// É um group (chat legado)
			groupCount++
			chatID := peer.ChatID

			// Buscar informações do chat
			chats, err := raw.MessagesGetChats(ctx, []int64{chatID})
			if err != nil {
				fmt.Printf("Erro ao buscar chat %d: %v\n", chatID, err)
				continue
			}

			// Processar chats retornados
			for _, chat := range chats.GetChats() {
				if legacyChat, ok := chat.(*tg.Chat); ok {
					fmt.Printf("💬 Group: %s - ID: %d\n", legacyChat.Title, legacyChat.ID)
					fmt.Printf("   👥 Membros: %d\n", legacyChat.ParticipantsCount)
				}
			}

		case *tg.InputPeerUser:
			// É uma conversa privada com usuário - pular
			continue
		}
	}

	// Verificar se houve erro na iteração
	if err := iter.Err(); err != nil {
		return errors.Wrap(err, "erro ao iterar diálogos")
	}

	fmt.Printf("\n📊 Resumo:\n")
	fmt.Printf("   📢 Channels: %d\n", channelCount)
	fmt.Printf("   👥 Groups/Supergroups: %d\n", groupCount)
	fmt.Printf("   📋 Total: %d\n", channelCount+groupCount)

	return nil
}

// listFolders lista todas as pastas (dialog filters) do usuário
func listFolders(ctx context.Context, raw *tg.Client) error {
	fmt.Println("\n=== 📁 Listando Folders (Pastas) ===")

	// Buscar todos os dialog filters (folders)
	dialogFilters, err := raw.MessagesGetDialogFilters(ctx)
	if err != nil {
		return errors.Wrap(err, "erro ao buscar dialog filters")
	}

	// Processar os filters retornados
	filters := dialogFilters.GetFilters()

	if len(filters) == 0 {
		fmt.Println("📭 Nenhuma pasta configurada")
		return nil
	}

	fmt.Printf("📊 Total de pastas: %d\n\n", len(filters))

	for i, filter := range filters {
		switch f := filter.(type) {
		case *tg.DialogFilter:
			// Pasta normal criada pelo usuário
			fmt.Printf("📁 [%d] %s", i+1, extractTitle(f.Title))
			if f.Emoticon != "" {
				fmt.Printf(" %s", f.Emoticon)
			}
			fmt.Printf(" (ID: %d)\n", f.ID)

			// Mostrar configurações da pasta
			showFilterSettings(f)

			// Mostrar peers incluídos
			if len(f.IncludePeers) > 0 {
				fmt.Printf("   ✅ Incluídos: %d chats\n", len(f.IncludePeers))
			}

			// Mostrar peers excluídos
			if len(f.ExcludePeers) > 0 {
				fmt.Printf("   ❌ Excluídos: %d chats\n", len(f.ExcludePeers))
			}

			// Mostrar peers fixados
			if len(f.PinnedPeers) > 0 {
				fmt.Printf("   📌 Fixados: %d chats\n", len(f.PinnedPeers))
			}

		case *tg.DialogFilterChatlist:
			// Pasta compartilhada (chatlist)
			fmt.Printf("🔗 [%d] %s", i+1, extractTitle(f.Title))
			if f.Emoticon != "" {
				fmt.Printf(" %s", f.Emoticon)
			}
			fmt.Printf(" (ID: %d) - Compartilhada", f.ID)

			if f.HasMyInvites {
				fmt.Printf(" 🎫")
			}
			fmt.Println()

			// Mostrar cor se definida
			if color, ok := f.GetColor(); ok {
				fmt.Printf("   🎨 Cor: %s\n", getColorName(color))
			}

			// Mostrar peers incluídos
			if len(f.IncludePeers) > 0 {
				fmt.Printf("   ✅ Incluídos: %d chats\n", len(f.IncludePeers))
			}

			// Mostrar peers fixados
			if len(f.PinnedPeers) > 0 {
				fmt.Printf("   📌 Fixados: %d chats\n", len(f.PinnedPeers))
			}

		case *tg.DialogFilterDefault:
			// Pasta padrão (Todos os chats) - apenas para usuários Premium
			fmt.Printf("⭐ [%d] Todos os Chats (Padrão)\n", i+1)
		}

		fmt.Println() // Linha em branco entre pastas
	}

	// Mostrar se tags estão habilitadas
	if dialogFilters.TagsEnabled {
		fmt.Println("🏷️  Tags de pastas estão habilitadas")
	}

	return nil
}

// extractTitle extrai o texto de um TextWithEntities
func extractTitle(title tg.TextWithEntities) string {
	return title.Text
}

// showFilterSettings mostra as configurações de filtro de uma pasta
func showFilterSettings(filter *tg.DialogFilter) {
	var settings []string

	if filter.Contacts {
		settings = append(settings, "👥 Contatos")
	}
	if filter.NonContacts {
		settings = append(settings, "👤 Não-contatos")
	}
	if filter.Groups {
		settings = append(settings, "👥 Grupos")
	}
	if filter.Broadcasts {
		settings = append(settings, "📢 Canais")
	}
	if filter.Bots {
		settings = append(settings, "🤖 Bots")
	}
	if filter.ExcludeMuted {
		settings = append(settings, "🔇 Excluir silenciados")
	}
	if filter.ExcludeRead {
		settings = append(settings, "✅ Excluir lidos")
	}
	if filter.ExcludeArchived {
		settings = append(settings, "📦 Excluir arquivados")
	}

	if len(settings) > 0 {
		fmt.Printf("   ⚙️  Filtros: %s\n", strings.Join(settings, ", "))
	}
}

// getColorName retorna o nome da cor baseado no índice
func getColorName(colorIndex int) string {
	colors := map[int]string{
		-1: "Oculta",
		0:  "Vermelho",
		1:  "Laranja",
		2:  "Violeta",
		3:  "Verde",
		4:  "Ciano",
		5:  "Azul",
		6:  "Rosa",
	}

	if name, exists := colors[colorIndex]; exists {
		return name
	}
	return fmt.Sprintf("Cor %d", colorIndex)
}

// truncateMessage trunca uma mensagem para um tamanho máximo
func truncateMessage(message string, maxLen int) string {
	if len(message) <= maxLen {
		return message
	}
	return message[:maxLen] + "..."
}

// formatDate converte timestamp Unix para formato legível
func formatDate(timestamp int) string {
	if timestamp == 0 {
		return "Data desconhecida"
	}
	// Converter timestamp Unix para time.Time e formatar
	// Por simplicidade, vamos apenas mostrar o timestamp por enquanto
	return fmt.Sprintf("%d", timestamp)
}

// List all dialogs (conversations) which includes channels and groups
// if err := listChannelsAndGroups(ctx, raw); err != nil {
// 	return errors.Wrap(err, "list channels and groups")
// }

// List all folders (dialog filters)
// 	📁 [2] Promoções (ID: 4)
//  	✅ Incluídos: 28 chats
// if err := listFolders(ctx, raw); err != nil {
// 	return errors.Wrap(err, "list folders")
// }

// list all channels from folder
// if err := listChannelsFromFolders(ctx, raw, 4); err != nil {
// 	return errors.Wrap(err, "list folders")
// }
