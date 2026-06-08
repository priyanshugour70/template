// Package pagination provides a uniform query-param shape for list endpoints.
package pagination

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	DefaultLimit = 25
	MaxLimit     = 200
)

type Params struct {
	Page   int
	Limit  int
	Offset int
	Sort   string // e.g. "-created_at" → ORDER BY created_at DESC
	Search string
}

func FromGin(c *gin.Context) Params {
	page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(strings.TrimSpace(c.Query("limit")))
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return Params{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
		Sort:   strings.TrimSpace(c.Query("sort")),
		Search: strings.TrimSpace(c.Query("q")),
	}
}

// SortClause translates "-field" / "field" into a safe ORDER BY clause. Only
// columns in allowed are accepted; everything else falls back to defaultCol.
func (p Params) SortClause(allowed map[string]bool, defaultCol string) string {
	s := p.Sort
	if s == "" {
		return defaultCol + " DESC"
	}
	dir := "ASC"
	if strings.HasPrefix(s, "-") {
		dir = "DESC"
		s = strings.TrimPrefix(s, "-")
	}
	if !allowed[s] {
		return defaultCol + " DESC"
	}
	return s + " " + dir
}
