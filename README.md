# TTLCache - an in-memory cache with item expiration

This is a friendly fork of [jellydator/ttlcache](https://github.com/jellydator/ttlcache),
which adds context support and a context aware singleflight implementation.

## Features
- Simple API.
- Type parameters.
- Item expiration and automatic deletion.
- Automatic expiration time extension on each `Get` call.
- `Loader` interface that is used to load/lazily initialize missing cache
items.
- Subscription to cache events (insertion and eviction).
- Metrics.
- Configurability.

## Installation
```
go get github.com/silas/ttlcache
```

## Usage
The main type of `ttlcache` is `Cache`. It represents a single
in-memory data store.

To create a new instance of `ttlcache.Cache`, the `ttlcache.New()` function
should be called:
```go
func main() {
	cache := ttlcache.New[string, string]()
}
```

Note that by default, a new cache instance does not let any of its
items to expire or be automatically deleted. However, this feature
can be activated by passing a few additional options into the
`ttlcache.New()` function and calling the `cache.Start(ctx)` method:
```go
func main() {
	cache := ttlcache.New[string, string](
		ttlcache.WithTTL[string, string](30 * time.Minute),
	)

	go cache.Start(context.Background()) // starts automatic expired item deletion
}
```

Even though the `cache.Start(ctx)` method handles expired item deletion well,
there may be times when the system that uses `ttlcache` needs to determine
when to delete the expired items itself. For example, it may need to
delete them only when the resource load is at its lowest (e.g., after
midnight, when the number of users/HTTP requests drops). So, in situations
like these, instead of calling `cache.Start(ctx)`, the system could
periodically call `cache.DeleteExpired(ctx)`:
```go
func main() {
	cache := ttlcache.New[string, string](
		ttlcache.WithTTL[string, string](30 * time.Minute),
	)

	for {
		time.Sleep(4 * time.Hour)
		cache.DeleteExpired(context.Background())
	}
}
```

The data stored in `ttlcache.Cache` can be retrieved and updated with
`Set`, `Get`, `Delete`, etc. methods:
```go
func main() {
	cache := ttlcache.New[string, string](
		ttlcache.WithTTL[string, string](30 * time.Minute),
	)
	ctx := context.Background()

	// insert data
	cache.Set(ctx, "first", "value1", ttlcache.DefaultTTL)
	cache.Set(ctx, "second", "value2", ttlcache.NoTTL)
	cache.Set(ctx, "third", "value3", ttlcache.DefaultTTL)

	// retrieve data
	item := cache.Get(ctx, "first")
	fmt.Println(item.Value(), item.ExpiresAt())

	// delete data
	cache.Delete(ctx, "second")
	cache.DeleteExpired(ctx)
	cache.DeleteAll(ctx)
}
```

To subscribe to insertion and eviction events, `cache.OnInsertion()` and
`cache.OnEviction()` methods should be used:
```go
func main() {
	cache := ttlcache.New[string, string](
		ttlcache.WithTTL[string, string](30 * time.Minute),
		ttlcache.WithCapacity[string, string](300),
	)

	cache.OnInsertion(func(ctx context.Context, item *ttlcache.Item[string, string]) {
		fmt.Println(item.Value(), item.ExpiresAt())
	})
	cache.OnEviction(func(ctx context.Context, reason ttlcache.EvictionReason, item *ttlcache.Item[string, string]) {
		if reason == ttlcache.EvictionReasonCapacityReached {
			fmt.Println(item.Key(), item.Value())
		}
	})

	ctx := context.Background()
	cache.Set(ctx, "first", "value1", ttlcache.DefaultTTL)
	cache.DeleteAll(ctx)
}
```

To load data when the cache does not have it, a custom or
existing implementation of `ttlcache.Loader` can be used:
```go
func main() {
	var loader ttlcache.Loader[string, string] = ttlcache.LoaderFunc[string, string](
		func(ctx context.Context, c *ttlcache.Cache[string, string], key string) *ttlcache.Item[string, string] {
			// load from file/make an HTTP request
			item := c.Set(ctx, "key from file", "value from file", time.Minute)
			return item
		},
	)
	loader = ttlcache.SingleFlightLoader(loader)
	cache := ttlcache.New[string, string](
		ttlcache.WithLoader[string, string](loader),
	)

	item := cache.Get(context.Background(), "key from file")
}
```
