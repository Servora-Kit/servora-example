package openfga

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Servora-Kit/servora/pkg/redis"
)

const (
	DefaultCheckCacheTTL = 60 * time.Second
	DefaultListCacheTTL  = 10 * time.Minute
)

// CachedCheck is like Check but caches results in Redis.  If the Redis client
// is nil the call degrades to a plain Check.
func (c *Client) CachedCheck(ctx context.Context, rdb *redis.Client, ttl time.Duration,
	userID, relation, objectType, objectID string) (bool, error) {

	if rdb == nil {
		return c.Check(ctx, userID, relation, objectType, objectID)
	}

	key := checkCacheKey(userID, relation, objectType, objectID)

	cached, err := rdb.Get(ctx, key)
	if err == nil {
		return cached == "1", nil
	}

	allowed, err := c.Check(ctx, userID, relation, objectType, objectID)
	if err != nil {
		return false, err
	}

	_ = rdb.Set(ctx, key, boolStr(allowed), ttl)
	return allowed, nil
}

// CachedListObjects is like ListObjects but caches the full ID list in Redis.
// Subsequent calls within the TTL window return the cached result, avoiding
// repeated OpenFGA round-trips.  Returns all IDs; the caller is responsible
// for pagination.
func (c *Client) CachedListObjects(ctx context.Context, rdb *redis.Client, ttl time.Duration,
	userID, relation, objectType string) ([]string, error) {

	if rdb == nil {
		return c.ListObjects(ctx, userID, relation, objectType)
	}

	key := listCacheKey(userID, relation, objectType)

	members, err := rdb.SMembers(ctx, key)
	if err == nil && len(members) > 0 {
		return members, nil
	}

	ids, err := c.ListObjects(ctx, userID, relation, objectType)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return ids, nil
	}

	vals := make([]any, len(ids))
	for i, id := range ids {
		vals[i] = id
	}
	_ = rdb.SAdd(ctx, key, vals...)
	_ = rdb.Expire(ctx, key, ttl)

	return ids, nil
}

// InvalidateCheck removes a cached Check result.
func InvalidateCheck(ctx context.Context, rdb *redis.Client, userID, relation, objectType, objectID string) {
	if rdb == nil {
		return
	}
	_ = rdb.Del(ctx, checkCacheKey(userID, relation, objectType, objectID))
}

// InvalidateListObjects removes a cached ListObjects result.
func InvalidateListObjects(ctx context.Context, rdb *redis.Client, userID, relation, objectType string) {
	if rdb == nil {
		return
	}
	_ = rdb.Del(ctx, listCacheKey(userID, relation, objectType))
}

func checkCacheKey(userID, relation, objectType, objectID string) string {
	return fmt.Sprintf("authz:check:%s:%s:%s:%s", userID, relation, objectType, objectID)
}

func listCacheKey(userID, relation, objectType string) string {
	return fmt.Sprintf("authz:list:%s:%s:%s", userID, relation, objectType)
}

func boolStr(v bool) string { return strconv.FormatBool(v) }
