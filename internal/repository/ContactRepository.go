package repository

import (
	"go-csv-import/internal/db"
	"go-csv-import/internal/logger"
	"go-csv-import/internal/model"
)

type ContactRepository interface {
	Insert(contact *model.Contact) error
	InsertBatch(contacts []*model.Contact) error
	Truncate() error
}

type contactRepository struct{}

func NewContactRepository() *contactRepository {
	return &contactRepository{}
}

func (r *contactRepository) Insert(c *model.Contact) error {
	return db.DB.Create(c).Error
}

func (r *contactRepository) InsertBatch(c []*model.Contact) error {
	return db.DB.Create(c).Error
}

func (r *contactRepository) Truncate() error {
	logger.Trace("Truncating contacts table...")
	err := db.DB.Exec("TRUNCATE TABLE contacts").Error
	logger.Trace("...Contacts table trucated")
	return err
}
