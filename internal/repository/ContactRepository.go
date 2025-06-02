package repository

import (
	"context"
	"go-csv-import/internal/db"
	"go-csv-import/internal/logger"
	"go-csv-import/internal/model"
)

type ContactRepository interface {
	Insert(contact *model.Contact) error
	InsertBatch(ctx context.Context, contacts []*model.Contact) error
	Truncate() error
	CountByReqId(reqId string) (int, error)
	DeleteByReqId(ctx context.Context, reqId string) error
}

type contactRepository struct{}

func NewContactRepository() *contactRepository {
	return &contactRepository{}
}

func (r *contactRepository) Insert(c *model.Contact) error {
	return db.DB.Create(c).Error
}

func (r *contactRepository) InsertBatch(ctx context.Context, c []*model.Contact) error {
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
	err := db.DB.Where("req_id = ?", reqId).Model(&model.Contact{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *contactRepository) DeleteByReqId(ctx context.Context, reqId string) error {
	return db.DB.
		WithContext(ctx).
		Where("req_id = ?", reqId).
		Delete(&model.Contact{}).
		Error
}
