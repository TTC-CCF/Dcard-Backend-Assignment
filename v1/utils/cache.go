package utils

import (
	"encore.dev/storage/cache"
)

var Cluster = cache.NewCluster("backend", cache.ClusterConfig{
	EvictionPolicy: cache.AllKeysLRU,
})
