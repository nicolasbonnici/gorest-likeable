package likeable

import (
	"context"
	"errors"
	"fmt"

	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/query"
)

// LikeService exposes read-optimized access to like data. Counts and
// membership are resolved by the database so callers never materialize
// individual rows just to tally them, and list views resolve their whole
// page of objects in a single round-trip instead of one query per item.
type LikeService struct {
	db   database.Database
	crud *crud.CRUD[Like]
}

func NewLikeService(db database.Database) *LikeService {
	return &LikeService{
		db:   db,
		crud: crud.New[Like](db),
	}
}

func (s *LikeService) GetByID(ctx context.Context, id string) (*Like, error) {
	return s.crud.GetByID(ctx, id)
}

// Count returns the number of likes for a single object using a COUNT
// aggregate, which lets the database answer from the (likeable, likeable_id)
// index without transferring any rows.
func (s *LikeService) Count(ctx context.Context, likeableType, likeableID string) (int64, error) {
	q, args, err := query.New(s.db.Dialect()).
		Select("COUNT(*)").
		From(likesTable).
		Where(query.Eq("likeable", likeableType)).
		Where(query.Eq("likeable_id", likeableID)).
		Build()
	if err != nil {
		return 0, fmt.Errorf("build count query: %w", err)
	}

	var count int64
	if err := s.db.QueryRow(ctx, q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// CountBatch resolves like counts for a whole list of objects in one grouped
// aggregate query, avoiding the N+1 pattern of counting each object
// separately. Every requested id is present in the result; objects with no
// likes map to zero.
func (s *LikeService) CountBatch(ctx context.Context, likeableType string, likeableIDs []string) (map[string]int64, error) {
	counts := make(map[string]int64, len(likeableIDs))
	for _, id := range likeableIDs {
		counts[id] = 0
	}
	if len(likeableIDs) == 0 {
		return counts, nil
	}

	q, args, err := query.New(s.db.Dialect()).
		Select("likeable_id", "COUNT(*)").
		From(likesTable).
		Where(query.Eq("likeable", likeableType)).
		Where(query.In("likeable_id", toAnySlice(likeableIDs)...)).
		GroupBy("likeable_id").
		Build()
	if err != nil {
		return nil, fmt.Errorf("build batch count query: %w", err)
	}

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var count int64
		if err := rows.Scan(&id, &count); err != nil {
			return nil, err
		}
		counts[id] = count
	}
	return counts, rows.Err()
}

// LikedByBatch reports, for a whole list of objects, which ones the given user
// has already liked. A single IN query replaces one existence check per
// object. Every requested id is present in the result; an empty likerID yields
// an all-false map without touching the database.
func (s *LikeService) LikedByBatch(ctx context.Context, likerID, likeableType string, likeableIDs []string) (map[string]bool, error) {
	liked := make(map[string]bool, len(likeableIDs))
	for _, id := range likeableIDs {
		liked[id] = false
	}
	if likerID == "" || len(likeableIDs) == 0 {
		return liked, nil
	}

	q, args, err := query.New(s.db.Dialect()).
		Select("likeable_id").
		Distinct().
		From(likesTable).
		Where(query.Eq("liker_id", likerID)).
		Where(query.Eq("likeable", likeableType)).
		Where(query.In("likeable_id", toAnySlice(likeableIDs)...)).
		Build()
	if err != nil {
		return nil, fmt.Errorf("build liked-state query: %w", err)
	}

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		liked[id] = true
	}
	return liked, rows.Err()
}

var errInvalidIDType = errors.New("invalid ID type")

var likesTable = Like{}.TableName()

func toAnySlice(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}
