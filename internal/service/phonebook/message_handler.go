package phonebook

import (
	"context"
	"fmt"
	"go-csv-import/internal/amqp"
	"go-csv-import/internal/db"
	"go-csv-import/internal/logger"
	"strings"
	"time"

	rabbit "github.com/streadway/amqp"
)

// AMQP message type handled by phonebook service
type MessageType string

const (
	MessageTypeUpload MessageType = "upload"
	MessageTypeDelete MessageType = "delete"
)

type MessageHandlerFunc func(ctx context.Context, msg rabbit.Delivery) (ack bool, err error)

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
func (h *MessageHandler) Process(ctx context.Context, msg rabbit.Delivery) (ack bool, err error) {
	if f, ok := h.handlers[MessageType(msg.Type)]; ok {
		return f(ctx, msg)
	}
	return true, fmt.Errorf("MessageHandler not found for type [%s]", msg.Type)
}

func (p *PhonebookHandler) handleMessageInsertPhonebook(ctx context.Context, msg rabbit.Delivery) (ack bool, err error) {
	// Decode the message body into a FileMessage struct.
	var file *FileMessage
	message := amqp.NewJsonMessageDecoder(msg.Body)
	errd := message.Decode(&file)
	if errd != nil {
		logger.Error("Decode AMQP message", "body", msg.Body, "error", err, "type", fmt.Sprintf("%T", err))
		return true, fmt.Errorf("cannot decode AMQP message for FileMessage")
	}

	start := time.Now()
	logger.Info("Treating file", "file", file.FilePath)

	if erru := p.Uploader.Upload(ctx, file); erru != nil {
		p.ProgressStore.SetError(file.Uuid, erru)
		p.printTypedErrors(erru, file)
	} else {
		logger.Info("File successful treated", "file", file.FilePath, "time", time.Since(start))
	}

	file.Remove()
	return true, nil
}

func (p *PhonebookHandler) handleMessageDeletePhonebook(ctx context.Context, msg rabbit.Delivery) (ack bool, err error) {
	// Decode the message body into a FileMessage struct.
	var file *FileMessage
	message := amqp.NewJsonMessageDecoder(msg.Body)
	errd := message.Decode(&file)
	if errd != nil {
		logger.Error("Decode AMQP message", "body", msg.Body, "error", err, "type", fmt.Sprintf("%T", err))
		return true, fmt.Errorf("cannot decode AMQP message for FileMessage")
	}

	start := time.Now()
	logger.Info("Deleting contacts...", "uuid", file.Uuid)

	err = p.Uploader.Repository.DeleteByReqId(ctx, file.Uuid)
	if err != nil {
		// TODO: refactor to retry x times with msg headers. log retries and error type to get error with other than Contains
		if strings.Contains(err.Error(), "Lock wait timeout exceeded") && !msg.Redelivered {
			logger.Warn("Retrying after lock timeout", "uuid", file.Uuid)
			time.Sleep(2 * time.Second)
			return false, err
		}

		logger.Error("Cannot delete contacts", "error", err)
		return true, db.NewDbError(fmt.Errorf("cannot delete contacts: %w", err))
	}

	logger.Info("Contacts successful deleted ", "time", time.Since(start))
	return true, nil
}
