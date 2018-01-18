package config

import (
	"time"
)

// ReqCacheTime reqcache持续时间
var ReqCacheTime = 24 * time.Hour

var ClickCacheTime = time.Hour * 24 * 31
