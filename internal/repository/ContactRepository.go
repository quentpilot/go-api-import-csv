package repository

import (
	"context"
	"go-csv-import/internal/db"
	"go-csv-import/internal/logger"
	"go-csv-import/internal/model"

	"gorm.io/hints"
)

type Repository interface {
	Insert(contact *model.Contact) error
	InsertBatch(ctx context.Context, contacts []*model.Contact) error
	Truncate() error
}

type ContactRepository struct{}

func NewContactRepository() *ContactRepository {
	return &ContactRepository{}
}

func (r *ContactRepository) Insert(c *model.Contact) error {
	return db.DB.Create(c).Error
}

func (r *ContactRepository) InsertBatch(ctx context.Context, c []*model.Contact) error {
	return db.DB.Clauses(hints.IgnoreIndex("idx_req_id")).Create(c).Error
}

func (r *ContactRepository) Truncate() error {
	logger.Trace("Truncating contacts table...")
	err := db.DB.Exec("TRUNCATE TABLE contacts").Error
	logger.Trace("...Contacts table trucated")
	return err
}

func (r *ContactRepository) CountByReqId(reqId string) (int, error) {
	var count int64
	err := db.DB.Where("req_id = ?", reqId).Model(&model.Contact{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *ContactRepository) DeleteByReqId(ctx context.Context, reqId string) error {
	return db.DB.
		WithContext(ctx).
		Where("req_id = ?", reqId).
		Delete(&model.Contact{}).
		Error
}
