package openfga

import (
	"context"
	"fmt"
	"strings"
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

// InvalidateForTuples invalidates all cached Check and ListObjects entries
// that could be affected by the given tuples. This should be called after
// WriteTuples or DeleteTuples to keep the cache consistent.
//
// For each tuple it invalidates:
//   - The exact Check cache entry (user + relation + object)
//   - The ListObjects cache for the user on the object's type with the tuple's relation
//   - The ListObjects cache for common computed relations (can_view, can_edit, can_admin, can_manage)
func InvalidateForTuples(ctx context.Context, rdb *redis.Client, tuples []Tuple) {
	if rdb == nil || len(tuples) == 0 {
		return
	}

	var keys []string
	seen := make(map[string]struct{})

	for _, t := range tuples {
		userID, objectType, objectID := parseTupleComponents(t)
		if userID == "" || objectType == "" {
			continue
		}

		if objectID != "" {
			k := checkCacheKey(userID, t.Relation, objectType, objectID)
			if _, ok := seen[k]; !ok {
				keys = append(keys, k)
				seen[k] = struct{}{}
			}
		}

		for _, rel := range affectedRelations(t.Relation, objectType) {
			k := listCacheKey(userID, rel, objectType)
			if _, ok := seen[k]; !ok {
				keys = append(keys, k)
				seen[k] = struct{}{}
			}
		}
	}

	for _, k := range keys {
		_ = rdb.Del(ctx, k)
	}
}

// parseTupleComponents extracts the bare userID, objectType, and objectID from a Tuple.
// Tuple.User is e.g. "user:abc" or "tenant:root"; Tuple.Object is e.g. "organization:xyz".
func parseTupleComponents(t Tuple) (userID, objectType, objectID string) {
	if i := strings.IndexByte(t.User, ':'); i >= 0 && strings.HasPrefix(t.User, "user:") {
		userID = t.User[i+1:]
	}
	if i := strings.IndexByte(t.Object, ':'); i >= 0 {
		objectType = t.Object[:i]
		objectID = t.Object[i+1:]
	}
	return
}

// affectedRelations returns the tuple's own relation plus computed relations
// that might be affected by a change to the given assignable relation.
func affectedRelations(relation, objectType string) []string {
	rels := []string{relation}

	computedByType := map[string][]string{
		"tenant":       {"can_view", "can_manage"},
		"organization": {"can_view", "can_manage", "can_manage_members"},
		"project":      {"can_view", "can_edit", "can_admin", "can_manage_members"},
	}
	if computed, ok := computedByType[objectType]; ok {
		rels = append(rels, computed...)
	}
	return rels
}

func checkCacheKey(userID, relation, objectType, objectID string) string {
	return fmt.Sprintf("authz:check:%s:%s:%s:%s", userID, relation, objectType, objectID)
}

func listCacheKey(userID, relation, objectType string) string {
	return fmt.Sprintf("authz:list:%s:%s:%s", userID, relation, objectType)
}

func boolStr(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
