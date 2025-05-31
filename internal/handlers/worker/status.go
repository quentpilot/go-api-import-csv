package worker

import (
	"go-csv-import/internal/utils"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type MessageProgressStore struct {
	counter sync.Map
}

type MessageProgress struct {
	Inserted atomic.Int64
	Total    atomic.Int64
}

type MessageProgressResponse struct {
	Status     string  `json:"Status"`
	Total      int64   `json:"Total"`
	Inserted   int64   `json:"Inserted"`
	Percentile float64 `json:"Percentile"`
}

func NewMessageProgressStore() *MessageProgressStore {
	return &MessageProgressStore{}
}

func (s *MessageProgressStore) Init(reqId string, total int64) {
	var p MessageProgress
	p.Total.Store(total)
	s.counter.Store(reqId, &p)
}

func (s *MessageProgressStore) Increment(reqId string, batch int64) {
	if val, ok := s.counter.Load(reqId); ok {
		if progress, ok := val.(*MessageProgress); ok {
			progress.Inserted.Add(batch)
		}
	}
}

func (s *MessageProgressStore) Get(reqId string) (inserted int64, total int64, ok bool) {
	if val, ok := s.counter.Load(reqId); ok {
		if progress, ok := val.(*MessageProgress); ok {
			return progress.Inserted.Load(), progress.Total.Load(), true
		}
	}
	return 0, 0, false
}

func (s *MessageProgressStore) Handler() http.Handler {
	r := gin.Default()
	r.GET("/upload/status/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		if inserted, total, ok := s.Get(uuid); ok {
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
			}
			c.JSON(http.StatusOK, resp)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Progress not found"})
		}
	})
	return r
}
