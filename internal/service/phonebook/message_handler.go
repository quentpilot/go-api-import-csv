package phonebook

import (
	"context"
	"fmt"
	"go-csv-import/internal/amqp"
	"go-csv-import/internal/db"
	"go-csv-import/internal/logger"
	"time"

	rabbit "github.com/streadway/amqp"
)

// AMQP message type handled by phonebook service
type MessageType string

const (
	MessageTypeUpload MessageType = "upload"
	MessageTypeDelete MessageType = "delete"
)

type MessageHandlerFunc func(ctx context.Context, body []byte) error

// MessageHandler stores functions strategies to handle an AMQP message
type MessageHandler struct {
	/*
		handlers maps AMQP message type with accosiated function to launch.

			ctx is the context to propagate
			body is the AMQP message body
	*/
	handlers map[MessageType]MessageHandlerFunc
}

// Initialise map of type and associated strategies
func (p *PhonebookHandler) NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		handlers: map[MessageType]MessageHandlerFunc{
			MessageTypeUpload: p.handleMessageInsertPhonebook,
			MessageTypeDelete: p.handleMessageDeletePhonebook,
		},
	}
}

// Process runs the function associated to AMQP message type
func (h *MessageHandler) Process(ctx context.Context, msg rabbit.Delivery) error {
	if f, ok := h.handlers[MessageType(msg.Type)]; ok {
		return f(ctx, msg.Body)
	}
	return fmt.Errorf("MessageHandler not found for type [%s]", msg.Type)
}

func (p *PhonebookHandler) handleMessageInsertPhonebook(ctx context.Context, body []byte) error {
	// Decode the message body into a FileMessage struct.
	var file *FileMessage
	message := amqp.NewJsonMessageDecoder(body)
	err := message.Decode(&file)
	if err != nil {
		logger.Error("Decode AMQP message", "body", body, "error", err, "type", fmt.Sprintf("%T", err))
		return fmt.Errorf("cannot decode AMQP message for FileMessage")
	}

	start := time.Now()
	logger.Info("Treating file", "file", file.FilePath)

	if err := p.Uploader.Upload(ctx, file); err != nil {
		p.ProgressStore.SetError(file.Uuid, err)
		p.printTypedErrors(err, file)
	} else {
		logger.Info("File successful treated", "file", file.FilePath, "time", time.Since(start))
	}

	file.Remove()
	return nil
}

func (p *PhonebookHandler) handleMessageDeletePhonebook(ctx context.Context, body []byte) error {
	// Decode the message body into a FileMessage struct.
	var file *FileMessage
	message := amqp.NewJsonMessageDecoder(body)
	err := message.Decode(&file)
	if err != nil {
		logger.Error("Decode AMQP message", "body", body, "error", err, "type", fmt.Sprintf("%T", err))
		return fmt.Errorf("cannot decode AMQP message for FileMessage")
	}

	start := time.Now()
	logger.Info("Deleting contacts...")

	err = p.Uploader.Repository.DeleteByReqId(file.Uuid)
	if err != nil {
		logger.Error("Cannot delete contacts", "error", err)
		return db.NewDbError(fmt.Errorf("cannot delete contacts: %w", err))
	}

	logger.Info("Contacts successful deleted ", "time", time.Since(start))
	return nil
}
