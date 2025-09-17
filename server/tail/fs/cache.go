package fs

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	// defaultExpiration is the default time an item will be kept in the cache.
	defaultExpiration = 3 * time.Hour
	// cleanupInterval is the frequency at which the cache will be purged of expired items.
	cleanupInterval = 1 * time.Hour

	// alegCache stores aleg records, keyed by aleg.UUID.
	alegCache = cache.New(defaultExpiration, cleanupInterval)
	// blegCache stores bleg records, keyed by bleg.Originator.
	blegCache = cache.New(defaultExpiration, cleanupInterval)
)
