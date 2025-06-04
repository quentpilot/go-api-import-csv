package worker

import (
	"go-csv-import/internal/logger"
	"go-csv-import/internal/utils"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type MessageProgressStatusType string

const (
	StatusScheduled  MessageProgressStatusType = "Scheduled"
	StatusProcessing MessageProgressStatusType = "Processing"
	StatusCompleted  MessageProgressStatusType = "Completed"
	StatusError      MessageProgressStatusType = "Error"
)

// MessageProgressStore stores all progress file infos to deliver from API.
type MessageProgressStore struct {
	counter sync.Map
}

// MessageProgress stores current file progress infos.
type MessageProgress struct {
	Inserted  atomic.Int64
	Total     atomic.Int64
	Duration  atomic.Int64
	StartTime time.Time
	Error     error
}

// MessageProgressResponse is the interface contract
// between public and private API to transfert current file progress infos.
type MessageProgressResponse struct {
	Status     string  `json:"Status"`
	Total      int64   `json:"Total"`
	Inserted   int64   `json:"Inserted"`
	Percentile float64 `json:"Percentile"`
	Duration   string  `json:"Duration"`
}

func NewMessageProgressStore() *MessageProgressStore {
	return &MessageProgressStore{}
}

// Init sets start time processing file and total rows to insert
func (s *MessageProgressStore) Init(reqId string, total int64) {
	var p MessageProgress
	p.Total.Store(total)
	p.StartTime = time.Now()
	s.counter.Store(reqId, &p)
}

// Increment updates the total of inserted messages
func (s *MessageProgressStore) Increment(reqId string, batch int64) {
	if val, ok := s.counter.Load(reqId); ok {
		if progress, ok := val.(*MessageProgress); ok {
			progress.Inserted.Add(batch)
			dur := time.Since(progress.StartTime)
			progress.Duration.Store(dur.Nanoseconds())
		}
	}
}

// SetError stores last error to track status details
func (s *MessageProgressStore) SetError(reqId string, err error) {
	if val, ok := s.counter.Load(reqId); ok {
		if progress, ok := val.(*MessageProgress); ok {
			progress.Error = err
		}
	}
}

// Get retrieves file progress status from his identifier
func (s *MessageProgressStore) Get(reqId string) (inserted int64, total int64, duration int64, err error, ok bool) {
	if val, ok := s.counter.Load(reqId); ok {
		if progress, ok := val.(*MessageProgress); ok {
			return progress.Inserted.Load(), progress.Total.Load(), progress.Duration.Load(), progress.Error, true
		}
	}
	return 0, 0, 0, nil, false
}

// Handler retrieves progress file infos from file request identifier
func (s *MessageProgressStore) Handler() http.Handler {
	r := gin.Default()
	r.GET("/upload/status/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		logger.Info("Call endpoint /upload/status", "uuid", uuid)

		if inserted, total, duration, err, ok := s.Get(uuid); ok {
			resp := &MessageProgressResponse{
				Total:      total,
				Inserted:   inserted,
				Percentile: utils.MathRound(float64(inserted)/float64(total)*100, 3),
				Status:     s.getStatus(inserted, total, err),
				Duration:   time.Duration(duration).Round(time.Millisecond).String(),
			}

			statusCode := http.StatusOK
			if resp.Status == string(StatusError) {
				statusCode = http.StatusMultiStatus
				logger.Error("Progress Error Found", "error", err.Error())
				resp.Status += ": " + err.Error()
			}

			logger.Debug("Progress Found", "body", resp)
			c.JSON(statusCode, resp)
		} else {
			logger.Error("Progress Not Found")
			c.JSON(http.StatusNotFound, gin.H{"message": "Progress not found"})
		}
	})
	return r
}

// getStatus defines progress status as string following file progress state
func (s *MessageProgressStore) getStatus(inserted int64, total int64, err error) string {
	if err != nil {
		return string(StatusError)
	} else if inserted == 0 {
		return string(StatusScheduled)
	} else if inserted < total {
		return string(StatusProcessing)
	}
	return string(StatusCompleted)
}
