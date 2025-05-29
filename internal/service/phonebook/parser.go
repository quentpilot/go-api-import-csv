package phonebook

import (
	"errors"
	"fmt"
	"go-csv-import/internal/model"
	"log/slog"
)

func (c *ContactUploader) combine(header []string, row []string) (map[string]string, error) {
	if len(header) != len(row) {
		return nil, errors.New("header and row slices mismatch")
	}

	r := make(map[string]string, len(header))
	for i, k := range header {
		r[k] = row[i]
	}

	return r, nil
}

func (c *ContactUploader) createContactFromRow(header []string, row []string) (*model.Contact, error) {
	r, err := c.combine(header, row)
	if err != nil {
		return &model.Contact{}, err
	}
	slog.Debug("Combine row result", "combine", fmt.Sprintf("%#v", r))

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
		Phone:     r["Phone"],
		Firstname: r["Firstname"],
		Lastname:  r["Lastname"],
	}, nil
}

func (c *ContactUploader) handleBatchInsert(batch *Batch, header []string, row []string, force bool) error {
	var err error
	if len(row) > 0 {
		c, err := c.createContactFromRow(header, row)
		if err != nil {
			return err
		}
		slog.Debug("Model created", "contact", fmt.Sprintf("%#v", c))

		batch.Append(c)
	}

	if batch.IsReached(c.HttpConfig.BatchInsert) || (force && batch.Length > 0) {
		//time.Sleep(6 * time.Second)
		slog.Debug("Batch insert contacts", "total", batch.Length, "force", force)
		err = c.Repository.InsertBatch(batch.Contacts)
		batch.Reset()
	}

	return err
}
