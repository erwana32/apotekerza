package store

import (
	"context"
	"fmt"

	"github.com/erwana32/apotekerza/internal/identity/store/gen"
)

// ListAccessGroups returns one page of access groups ordered by name. Groups
// are global, so there is no company filter. The caller passes PageSize one
// larger than the page it shows, to learn whether another page exists, and
// AfterName (the last name of the previous page) for the next page.
func (q *Queries) ListAccessGroups(ctx context.Context, params gen.ListAccessGroupsParams) ([]gen.IdentityAccessGroup, error) {
	groups, err := q.Queries.ListAccessGroups(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("store: list access groups: %w", err)
	}
	return groups, nil
}
