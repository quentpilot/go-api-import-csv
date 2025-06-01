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
	CountByReqId(reqId string) (int, error)
	DeleteByReqId(reqId string) error
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

func (r *contactRepository) CountByReqId(reqId string) (int, error) {
	var count int64
	err := db.DB.Model(&model.Contact{}).Where("req_id = ?", reqId).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *contactRepository) DeleteByReqId(reqId string) error {
	logger.Trace("Delete contacts by req_id...", "req_id", reqId)
	return db.DB.Where("req_id = ?", reqId).Delete(&model.Contact{}).Error
}
