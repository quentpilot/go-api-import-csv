package phonebook

import (
	"go-csv-import/internal/logger"
	"go-csv-import/internal/model"
)

// BatchHandler defines the interface for handling batches of data
type BatchHandler interface {
	Reset()  // Resets the current batch infos
	Append() // Adds item to the current batch
}

// Batch represents the current batch of data to insert. It implements the BatchHandler interface.
type Batch struct {
	Contacts []*model.Contact // Current rows ready to be batch inserted
	Length   uint             // Number of Contacts rows
}

func NewBatch() *Batch {
	logger.Trace("Creating a new batch")
	return &Batch{}
}

func (b *Batch) Reset() {
	b.Contacts = []*model.Contact{}
	b.Length = 0
	logger.Trace("Batch reset")
}

func (b *Batch) Append(c *model.Contact) {
	b.Contacts = append(b.Contacts, c)
	b.Length++
	logger.Trace("Contact appended to batch")
}

func (b *Batch) IsReached(maxBatch uint) bool {
	return b.Length == maxBatch
}
