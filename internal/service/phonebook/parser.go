package phonebook

import (
	"errors"
	"fmt"
	"go-csv-import/internal/logger"
	"go-csv-import/internal/model"
)

func (c *ContactUploader) combine(header []string, row []string) (map[string]string, error) {
	if len(header) != len(row) {
		return nil, errors.New("header and row slices mismatch")
	}

	r := make(map[string]string, len(header))
	for i, k := range header {
		r[k] = row[i]
	}

	logger.Trace("Row combined", "row", fmt.Sprintf("%#v", r))
	return r, nil
}

func (c *ContactUploader) createContactFromRow(file *FilePart, header []string, row []string) (*model.Contact, error) {
	r, err := c.combine(header, row)
	if err != nil {
		return &model.Contact{}, err
	}

	required := []string{"Phone", "Firstname", "Lastname"}
	for i := 0; i < len(required); i++ {
		key := required[i]
		if _, exists := r[key]; exists {
			continue
		} else {
			return &model.Contact{}, fmt.Errorf("columns <%s> is missing", key)
		}
	}

	return &model.Contact{
		ReqId:     file.Uuid,
		Phone:     r["Phone"],
		Firstname: r["Firstname"],
		Lastname:  r["Lastname"],
	}, nil
}

func (c *ContactUploader) handleBatchInsert(file *FilePart, batch *Batch, header []string, row []string, force bool) error {
	var err error
	if len(row) > 0 {
		c, err := c.createContactFromRow(file, header, row)
		if err != nil {
			return err
		}
		logger.Trace("Contact model created", "contact", fmt.Sprintf("%#v", c))

		batch.Append(c)
	}

	if batch.IsReached(c.HttpConfig.BatchInsert) || (force && batch.Length > 0) {
		//time.Sleep(6 * time.Second)
		logger.Trace("Batch insert contacts", "total", batch.Length, "force", force)
		err = c.Repository.InsertBatch(batch.Contacts)
		c.ProgressStore.Increment(file.Uuid, int64(batch.Length))
		batch.Reset()
	}

	return err
}
