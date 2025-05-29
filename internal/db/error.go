package db

// DbError represents an error that occurs during database operations.
type DbError struct {
	Err error
}

func NewDbError(err error) *DbError {
	return &DbError{
		Err: err,
	}
}

func (e *DbError) Error() string {
	return "database error: " + e.Err.Error()
}
func (e *DbError) Unwrap() error {
	return e.Err
}
