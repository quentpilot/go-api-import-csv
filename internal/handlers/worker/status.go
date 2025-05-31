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

func (s *MessageProgressStore) Init(reqId string, total int64) {
	var p MessageProgress
	p.Total.Store(total)
	p.StartTime = time.Now()
	s.counter.Store(reqId, &p)
}

func (s *MessageProgressStore) Increment(reqId string, batch int64) {
	if val, ok := s.counter.Load(reqId); ok {
		if progress, ok := val.(*MessageProgress); ok {
			progress.Inserted.Add(batch)
			dur := time.Since(progress.StartTime)
			progress.Duration.Store(dur.Nanoseconds())
		}
	}
}

func (s *MessageProgressStore) Get(reqId string) (inserted int64, total int64, duration int64, ok bool) {
	if val, ok := s.counter.Load(reqId); ok {
		if progress, ok := val.(*MessageProgress); ok {
			return progress.Inserted.Load(), progress.Total.Load(), progress.Duration.Load(), true
		}
	}
	return 0, 0, 0, false
}

// Handler retrieves progress file infos from file request identifier
func (s *MessageProgressStore) Handler() http.Handler {
	r := gin.Default()
	r.GET("/upload/status/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		logger.Info("Call endpoint /upload/status", "uuid", uuid)

		if inserted, total, duration, ok := s.Get(uuid); ok {
			resp := &MessageProgressResponse{
				Total:      total,
				Inserted:   inserted,
				Percentile: utils.MathRound(float64(inserted)/float64(total)*100, 3),
				Status: func() string {
					if inserted == 0 {
						return "Scheduled"
					} else if inserted < total {
						return "Processing"
					}
					return "Completed"
				}(),
				Duration: time.Duration(duration).String(),
			}
			logger.Debug("Progress Found", "body", resp)
			c.JSON(http.StatusOK, resp)
		} else {
			logger.Error("Progress Not Found")
			c.JSON(http.StatusNotFound, gin.H{"message": "Progress not found"})
		}
	})
	return r
}
