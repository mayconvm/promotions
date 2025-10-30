# Como Listar Channels e Groups usando gotd/td

Este documento explica como usar a biblioteca `gotd/td` para listar channels e grupos no Telegram.

## Conceitos Importantes

### Tipos de Conversas no Telegram:
1. **User** - Conversa privada com usuário
2. **Chat** - Group legado (limitado a 200 membros)
3. **Channel** - Pode ser:
   - **Broadcast Channel** - Canal onde apenas admins postam
   - **Supergroup** - Grupo com até 200.000 membros

## Métodos para Listar Channels e Groups

### 1. Método Principal - Usando Query Builder

```go
func listChannelsAndGroups(ctx context.Context, raw *tg.Client) error {
    // Usar query builder para obter todos os diálogos
    iter := query.GetDialogs(raw).Iter()
    
    for iter.Next(ctx) {
        elem := iter.Value()
        
        // Pular diálogos deletados
        if elem.Deleted() {
            continue
        }
        
        switch peer := elem.Peer.(type) {
        case *tg.InputPeerChannel:
            // Channel ou Supergroup
            handleChannelOrSupergroup(ctx, raw, peer)
        case *tg.InputPeerChat:
            // Group legado
            handleLegacyGroup(ctx, raw, peer)
        case *tg.InputPeerUser:
            // Conversa privada - pular
            continue
        }
    }
    
    return iter.Err()
}
```

### 2. Identificando o Tipo Exato

Para distinguir entre **Channel** (broadcast) e **Supergroup**:

```go
func handleChannelOrSupergroup(ctx context.Context, raw *tg.Client, peer *tg.InputPeerChannel) {
    inputChannel := &tg.InputChannel{
        ChannelID:  peer.ChannelID,
        AccessHash: peer.AccessHash,
    }
    
    channels, err := raw.ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})
    if err != nil {
        return
    }
    
    for _, chat := range channels.GetChats() {
        if channel, ok := chat.(*tg.Channel); ok {
            if channel.Broadcast {
                // É um Channel (broadcast)
                fmt.Printf("📢 Channel: %s\n", channel.Title)
            } else {
                // É um Supergroup
                fmt.Printf("👥 Supergroup: %s\n", channel.Title)
            }
        }
    }
}
```

### 3. Obter Informações Detalhadas

```go
func getChannelDetails(ctx context.Context, raw *tg.Client, channelID, accessHash int64) {
    inputChannel := &tg.InputChannel{
        ChannelID:  channelID,
        AccessHash: accessHash,
    }
    
    // Informações básicas
    channels, _ := raw.ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})
    
    // Informações completas
    fullChannel, _ := raw.ChannelsGetFullChannel(ctx, inputChannel)
}
```

## Estrutura dos Dados

### Channel/Supergroup (`*tg.Channel`)
- `ID` - ID único do channel
- `AccessHash` - Hash necessário para acessar
- `Title` - Nome do channel
- `Username` - Username público (opcional)
- `Broadcast` - true se for channel, false se for supergroup
- `Megagroup` - true se for supergroup
- `ParticipantsCount` - Número de membros

### Chat Legado (`*tg.Chat`)
- `ID` - ID único do chat
- `Title` - Nome do grupo
- `ParticipantsCount` - Número de membros

## Exemplo de Uso Completo

```go
package main

import (
    "context"
    "fmt"
    "github.com/gotd/td/telegram"
    "github.com/gotd/td/telegram/query"
    "github.com/gotd/td/tg"
)

func main() {
    client := telegram.NewClient(appID, appHash, options)
    
    client.Run(context.Background(), func(ctx context.Context) error {
        raw := tg.NewClient(client)
        
        // Listar todos os channels e grupos
        return listChannelsAndGroups(ctx, raw)
    })
}
```

## Métodos Alternativos

### Usando MessagesGetDialogs diretamente

```go
func listUsingDirectAPI(ctx context.Context, raw *tg.Client) error {
    dialogs, err := raw.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
        Limit:      100,
        OffsetPeer: &tg.InputPeerEmpty{},
    })
    if err != nil {
        return err
    }
    
    // Processar dialogs...
    return nil
}
```

### Filtros Específicos

```go
// Apenas channels (broadcast)
func listChannelsOnly(ctx context.Context, raw *tg.Client) error {
    // Implementação que filtra apenas channels com Broadcast = true
}

// Apenas supergroups
func listSupergroupsOnly(ctx context.Context, raw *tg.Client) error {
    // Implementação que filtra apenas channels com Broadcast = false
}

// Apenas groups legados
func listLegacyGroupsOnly(ctx context.Context, raw *tg.Client) error {
    // Implementação que processa apenas InputPeerChat
}
```

## Tratamento de Erros

Sempre trate erros adequadamente:

```go
if err := iter.Err(); err != nil {
    return fmt.Errorf("erro ao iterar diálogos: %w", err)
}
```

## Performance

- Use paginação para grandes quantidades de diálogos
- Cache informações quando possível
- Use batch requests quando apropriado

## Limitações

- Alguns channels privados podem não ser acessíveis
- Rate limiting pode afetar requests em massa
- AccessHash é necessário para acessar channels

---

# Como Listar Folders (Pastas) usando gotd/td

## Conceitos Importantes sobre Folders

### Tipos de Folders no Telegram:
1. **DialogFilter** - Pasta normal criada pelo usuário
2. **DialogFilterChatlist** - Pasta compartilhada (importada)
3. **DialogFilterDefault** - Pasta "Todos os Chats" (apenas Premium)

## Método Principal - Listar Folders

```go
func listFolders(ctx context.Context, raw *tg.Client) error {
    // Buscar todos os dialog filters (folders)
    dialogFilters, err := raw.MessagesGetDialogFilters(ctx)
    if err != nil {
        return err
    }
    
    filters := dialogFilters.GetFilters()
    
    for _, filter := range filters {
        switch f := filter.(type) {
        case *tg.DialogFilter:
            // Pasta normal
            fmt.Printf("📁 %s (ID: %d)\n", f.Title.Text, f.ID)
        case *tg.DialogFilterChatlist:
            // Pasta compartilhada
            fmt.Printf("🔗 %s (ID: %d) - Compartilhada\n", f.Title.Text, f.ID)
        case *tg.DialogFilterDefault:
            // Pasta padrão (Premium)
            fmt.Println("⭐ Todos os Chats (Padrão)")
        }
    }
    
    return nil
}
```

## Estrutura dos Dados de Folders

### DialogFilter (Pasta Normal)
- `ID` - ID único da pasta
- `Title` - Nome da pasta (TextWithEntities)
- `Emoticon` - Emoji da pasta
- `Contacts` - Incluir contatos
- `NonContacts` - Incluir não-contatos  
- `Groups` - Incluir grupos
- `Broadcasts` - Incluir canais
- `Bots` - Incluir bots
- `ExcludeMuted` - Excluir silenciados
- `ExcludeRead` - Excluir lidos
- `ExcludeArchived` - Excluir arquivados
- `IncludePeers` - Chats específicos incluídos
- `ExcludePeers` - Chats específicos excluídos
- `PinnedPeers` - Chats fixados na pasta

### DialogFilterChatlist (Pasta Compartilhada)
- `ID` - ID único da pasta
- `Title` - Nome da pasta
- `Emoticon` - Emoji da pasta
- `Color` - Cor da tag da pasta (-1 a 6)
- `HasMyInvites` - Tem convites próprios
- `IncludePeers` - Chats incluídos
- `PinnedPeers` - Chats fixados

## Exemplos de Uso

### Listar Folders Simples
```go
func listFoldersSimple(ctx context.Context, raw *tg.Client) error {
    dialogFilters, err := raw.MessagesGetDialogFilters(ctx)
    if err != nil {
        return err
    }
    
    for i, filter := range dialogFilters.GetFilters() {
        switch f := filter.(type) {
        case *tg.DialogFilter:
            fmt.Printf("%d. 📁 %s\n", i+1, f.Title.Text)
        case *tg.DialogFilterChatlist:
            fmt.Printf("%d. 🔗 %s (Compartilhada)\n", i+1, f.Title.Text)
        }
    }
    
    return nil
}
```

### Obter Pastas Sugeridas
```go
func getSuggestedFolders(ctx context.Context, raw *tg.Client) error {
    suggested, err := raw.MessagesGetSuggestedDialogFilters(ctx)
    if err != nil {
        return err
    }
    
    for i, suggestion := range suggested {
        filter := suggestion.Filter.(*tg.DialogFilter)
        fmt.Printf("%d. 📁 %s - %s\n", i+1, filter.Title.Text, suggestion.Description)
    }
    
    return nil
}
```

### Criar Nova Pasta
```go
func createFolder(ctx context.Context, raw *tg.Client, title, emoticon string, folderID int) error {
    newFilter := &tg.DialogFilter{
        ID:          folderID,
        Title:       tg.TextWithEntities{Text: title},
        Emoticon:    emoticon,
        Contacts:    true,
        Groups:      true,
        Broadcasts:  true,
        IncludePeers: make([]tg.InputPeerClass, 0),
        ExcludePeers: make([]tg.InputPeerClass, 0),
        PinnedPeers:  make([]tg.InputPeerClass, 0),
    }
    
    success, err := raw.MessagesUpdateDialogFilter(ctx, &tg.MessagesUpdateDialogFilterRequest{
        ID:     folderID,
        Filter: newFilter,
    })
    
    return err
}
```

### Deletar Pasta
```go
func deleteFolder(ctx context.Context, raw *tg.Client, folderID int) error {
    // Para deletar, omitir o campo Filter
    success, err := raw.MessagesUpdateDialogFilter(ctx, &tg.MessagesUpdateDialogFilterRequest{
        ID: folderID,
        // Filter omitido para deletar
    })
    
    return err
}
```

### Reordenar Pastas
```go
func reorderFolders(ctx context.Context, raw *tg.Client, order []int) error {
    success, err := raw.MessagesUpdateDialogFiltersOrder(ctx, order)
    return err
}
```

## Cores das Tags (Pastas Premium)

As cores das tags vão de -1 a 6:
- `-1` - Oculta (não mostra tag)
- `0` - Vermelho
- `1` - Laranja
- `2` - Violeta
- `3` - Verde
- `4` - Ciano
- `5` - Azul
- `6` - Rosa

## API Methods para Folders

- `messages.getDialogFilters` - Lista folders existentes
- `messages.getSuggestedDialogFilters` - Lista folders sugeridas
- `messages.updateDialogFilter` - Criar/editar/deletar folder
- `messages.updateDialogFiltersOrder` - Reordenar folders
- `messages.toggleDialogFilterTags` - Habilitar/desabilitar tags (Business)