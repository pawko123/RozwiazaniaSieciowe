package protocol

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeProbe        MessageType = "PROBE"
	MessageTypeNickConflict MessageType = "NICK_CONFLICT"
	MessageTypeJoin         MessageType = "JOIN"
	MessageTypeHealthcheck  MessageType = "HEALTHCHECK"
	MessageTypeChat         MessageType = "CHAT"
	MessageTypeLeave        MessageType = "LEAVE"
)

type Message struct {
	ID        string      `json:"id"`
	Nickname  string      `json:"nickname"`
	Type      MessageType `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Content   string      `json:"content,omitempty"`
}

func NewMessage(nickname string, messageType MessageType, content string) Message {
	return Message{
		ID:        fmt.Sprintf("%s-%s", nickname, uuid.NewString()),
		Nickname:  nickname,
		Type:      messageType,
		Timestamp: time.Now().Unix(),
		Content:   content,
	}
}

func SerializeMessage(message Message) ([]byte, error) {
	serializedBytes, serializationError := json.Marshal(message)
	if serializationError != nil {
		return nil, serializationError
	}
	return serializedBytes, nil
}

func DeserializeMessage(dataBuffer []byte) (Message, error) {
	var message Message
	deserializationError := json.Unmarshal(dataBuffer, &message)
	if deserializationError != nil {
		return Message{}, deserializationError
	}
	return message, nil
}
