package bench

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/silas/ttlcache"
)

var (
	ctx = context.Background()
)

func BenchmarkCacheSetWithoutTTL(b *testing.B) {
	cache := ttlcache.New[string, string]()

	for n := 0; n < b.N; n++ {
		cache.Set(ctx, fmt.Sprint(n%1000000), "value", ttlcache.NoTTL)
	}
}

func BenchmarkCacheSetWithGlobalTTL(b *testing.B) {
	cache := ttlcache.New[string, string](
		ttlcache.WithTTL[string, string](50 * time.Millisecond),
	)

	for n := 0; n < b.N; n++ {
		cache.Set(ctx, fmt.Sprint(n%1000000), "value", ttlcache.DefaultTTL)
	}
}
