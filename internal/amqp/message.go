package amqp

import "encoding/json"

/*
AmqpMessage defines the interface for AMQP messages.

It provides methods to encode and decode messages between AMQP producers and consumers.
*/
type AmqpMessage interface {
	Get() []byte                // Get the message as a byte slice
	Encode(any) ([]byte, error) // Encode the message to a valid AMQP bytes format
	Decode(any) error           // Decode the message from bytes to a wanted type
}

// JsonMessage implements the AmqpMessage interface for JSON formatted messages.
type JsonMessage struct {
	body []byte
}

// NewJsonMessage creates a new JsonMessage instance with an empty body
func NewJsonMessage() *JsonMessage {
	return &JsonMessage{}
}

// NewJsonMessageEncoder creates a new JsonMessage instance and encodes the provided body into JSON format
func NewJsonMessageEncoder(body any) (*JsonMessage, error) {
	m := &JsonMessage{}
	_, err := m.Encode(body)
	return m, err
}

// NewJsonMessageDecoder creates a new JsonMessage instance with the provided byte slice as its body
func NewJsonMessageDecoder(body []byte) *JsonMessage {
	return &JsonMessage{body: body}
}

// Get returns the body of the JsonMessage as a byte slice
func (m *JsonMessage) Get() []byte {
	return m.body
}

// Encode encodes the provided message into JSON format and stores it in the JsonMessage body
func (m *JsonMessage) Encode(message any) ([]byte, error) {
	body, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	m.body = body
	return []byte(m.body), nil
}

// Decode decodes the JsonMessage body into the provided decorated type
func (m *JsonMessage) Decode(decorated any) error {
	err := json.Unmarshal(m.body, &decorated)
	if err != nil {
		return err
	}
	return nil
}
