package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var CacheApiUploadStatus = cache.New(1*time.Second, 1*time.Minute)
