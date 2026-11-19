package peer

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"Lab3/protocol"
)

var ErrorNicknameTaken = errors.New("podany pseudonim jest juz zajety w sieci")

type NodeConfiguration struct {
	Nickname           string
	MulticastAddress   string
	HealthcheckPeriod  time.Duration
	UserTimeoutPeriod  time.Duration
	ProbeTimeoutPeriod time.Duration
	DedupCacheTTL      time.Duration
}

type Node struct {
	configuration        NodeConfiguration
	listenerConnection   *net.UDPConn
	senderConnection     *net.UDPConn
	messageCache         *cache.Cache
	presenceManager      *PresenceManager
	incomingConflictChan chan struct{}
	stopChannel          chan struct{}
	closeOnce            sync.Once

	OnChatMessageReceived func(message protocol.Message)
	OnSystemEventReceived func(systemEventMessage string)
	OnUserListChanged     func(activeUsers []string)
}

func NewNode(configuration NodeConfiguration) (*Node, error) {
	resolvedAddress, resolveError := net.ResolveUDPAddr("udp4", configuration.MulticastAddress)
	if resolveError != nil {
		return nil, fmt.Errorf("blad parsowania adresu multicast: %w", resolveError)
	}

	listenerConnection, listenError := net.ListenMulticastUDP("udp4", nil, resolvedAddress)
	if listenError != nil {
		return nil, fmt.Errorf("blad nasluchiwania multicast: %w", listenError)
	}

	senderConnection, dialError := net.DialUDP("udp4", nil, resolvedAddress)
	if dialError != nil {
		listenerConnection.Close()
		return nil, fmt.Errorf("blad tworzenia gniazda nadawczego: %w", dialError)
	}

	var messageCache = cache.New(configuration.DedupCacheTTL, 10*time.Second)

	var node = &Node{
		configuration:        configuration,
		listenerConnection:   listenerConnection,
		senderConnection:     senderConnection,
		messageCache:         messageCache,
		incomingConflictChan: make(chan struct{}, 10),
		stopChannel:          make(chan struct{}),
	}

	var presenceManager = NewPresenceManager(
		configuration.Nickname,
		configuration.UserTimeoutPeriod,
		1*time.Second,
		node.handleUserJoined,
		node.handleUserTimeout,
		node.handleUserLeft,
	)
	node.presenceManager = presenceManager

	return node, nil
}

func (node *Node) ProbeAndJoin() error {
	go node.runNetworkReadLoop()

	var probeMessage = protocol.NewMessage(node.configuration.Nickname, protocol.MessageTypeProbe, "")
	node.IsDuplicateOrAdd(probeMessage.ID)

	sendingError := node.sendRawMessage(probeMessage)
	if sendingError != nil {
		return fmt.Errorf("blad wysylania pakietu probe: %w", sendingError)
	}

	var timer = time.NewTimer(node.configuration.ProbeTimeoutPeriod)
	defer timer.Stop()

	select {
	case <-node.incomingConflictChan:
		return ErrorNicknameTaken
	case <-timer.C:
	}

	var joinMessage = protocol.NewMessage(node.configuration.Nickname, protocol.MessageTypeJoin, "")
	node.IsDuplicateOrAdd(joinMessage.ID)

	sendingJoinError := node.sendRawMessage(joinMessage)
	if sendingJoinError != nil {
		return fmt.Errorf("blad wysylania pakietu join: %w", sendingJoinError)
	}

	go node.runHealthcheckLoop()

	return nil
}

func (node *Node) SendChatMessage(chatMessageContent string) error {
	if chatMessageContent == "" {
		return nil
	}

	var chatMessage = protocol.NewMessage(
		node.configuration.Nickname,
		protocol.MessageTypeChat,
		chatMessageContent,
	)

	node.IsDuplicateOrAdd(chatMessage.ID)

	for range 1 {
		_ = node.sendRawMessage(chatMessage)
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

func (node *Node) LeaveAndClose() {
	node.closeOnce.Do(func() {
		close(node.stopChannel)

		var leaveMessage = protocol.NewMessage(node.configuration.Nickname, protocol.MessageTypeLeave, "")
		node.IsDuplicateOrAdd(leaveMessage.ID)
		_ = node.sendRawMessage(leaveMessage)

		node.presenceManager.Stop()

		if node.listenerConnection != nil {
			_ = node.listenerConnection.Close()
		}
		if node.senderConnection != nil {
			_ = node.senderConnection.Close()
		}
	})
}

func (node *Node) GetActiveUsers() []string {
	return node.presenceManager.GetActiveUsers()
}

func (node *Node) IsDuplicateOrAdd(messageIdentifier string) bool {
	addingError := node.messageCache.Add(messageIdentifier, struct{}{}, cache.DefaultExpiration)
	return addingError != nil
}

func (node *Node) sendRawMessage(message protocol.Message) error {
	serializedPayload, serializationError := protocol.SerializeMessage(message)
	if serializationError != nil {
		return serializationError
	}

	_, writeError := node.senderConnection.Write(serializedPayload)
	return writeError
}

func (node *Node) runNetworkReadLoop() {
	var packetBuffer = make([]byte, 65535)

	for {
		select {
		case <-node.stopChannel:
			return
		default:
		}

		bytesReadCount, _, readError := node.listenerConnection.ReadFrom(packetBuffer)
		if readError != nil {
			select {
			case <-node.stopChannel:
				return
			default:
				continue
			}
		}

		if bytesReadCount == 0 {
			continue
		}

		receivedMessage, deserializationError := protocol.DeserializeMessage(packetBuffer[:bytesReadCount])
		if deserializationError != nil {
			continue
		}

		if node.IsDuplicateOrAdd(receivedMessage.ID) {
			continue
		}
		node.processIncomingMessage(receivedMessage)
	}
}

func (node *Node) processIncomingMessage(receivedMessage protocol.Message) {
	switch receivedMessage.Type {
	case protocol.MessageTypeProbe:
		if receivedMessage.Nickname == node.configuration.Nickname {
			var conflictResponse = protocol.NewMessage(
				node.configuration.Nickname,
				protocol.MessageTypeNickConflict,
				"",
			)
			node.IsDuplicateOrAdd(conflictResponse.ID)
			_ = node.sendRawMessage(conflictResponse)
		}

	case protocol.MessageTypeNickConflict:
		if receivedMessage.Nickname == node.configuration.Nickname {
			select {
			case node.incomingConflictChan <- struct{}{}:
			default:
			}
		}

	case protocol.MessageTypeJoin:
		node.presenceManager.RecordActivity(receivedMessage.Nickname)
		node.notifyUserListUpdated()

		var healthcheckReply = protocol.NewMessage(node.configuration.Nickname, protocol.MessageTypeHealthcheck, "")
		node.IsDuplicateOrAdd(healthcheckReply.ID)
		_ = node.sendRawMessage(healthcheckReply)

	case protocol.MessageTypeHealthcheck:
		hasJoinedNew := node.presenceManager.RecordActivity(receivedMessage.Nickname)
		if hasJoinedNew {
			node.notifyUserListUpdated()
		}

	case protocol.MessageTypeChat:
		if receivedMessage.Nickname == node.configuration.Nickname {
			return
		}
		node.presenceManager.RecordActivity(receivedMessage.Nickname)
		if node.OnChatMessageReceived != nil {
			node.OnChatMessageReceived(receivedMessage)
		}
	case protocol.MessageTypeLeave:
		node.presenceManager.RemoveUser(receivedMessage.Nickname)
		node.notifyUserListUpdated()
	}
}

func (node *Node) runHealthcheckLoop() {
	var ticker = time.NewTicker(node.configuration.HealthcheckPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-node.stopChannel:
			return
		case <-ticker.C:
			var healthcheckMessage = protocol.NewMessage(node.configuration.Nickname, protocol.MessageTypeHealthcheck, "")
			node.IsDuplicateOrAdd(healthcheckMessage.ID)
			_ = node.sendRawMessage(healthcheckMessage)
		}
	}
}

func (node *Node) handleUserJoined(userNickname string) {
	node.notifyUserListUpdated()
	if node.OnSystemEventReceived != nil {
		node.OnSystemEventReceived(fmt.Sprintf("Uzytkownik %s dolaczyl do pokoju", userNickname))
	}
}

func (node *Node) handleUserTimeout(userNickname string) {
	node.notifyUserListUpdated()
	if node.OnSystemEventReceived != nil {
		node.OnSystemEventReceived(fmt.Sprintf("Uzytkownik %s rozlaczyl sie (timeout)", userNickname))
	}
}

func (node *Node) handleUserLeft(userNickname string) {
	node.notifyUserListUpdated()
	if node.OnSystemEventReceived != nil {
		node.OnSystemEventReceived(fmt.Sprintf("Uzytkownik %s opuscil czat", userNickname))
	}
}

func (node *Node) notifyUserListUpdated() {
	if node.OnUserListChanged != nil {
		var activeUsers = node.presenceManager.GetActiveUsers()
		node.OnUserListChanged(activeUsers)
	}
}
