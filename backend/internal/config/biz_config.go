package config

import "context"

// BizConfigReader exposes typed, read-only accessors over dynamic business config
// stored in the database. Modules depend on this interface so they can be tested
// without pulling in the concrete bizconfig module.
//
// Add typed methods here as new dynamic-config needs emerge. Prefer narrow
// methods (e.g. `FeatureXEnabled(ctx)`) over a generic Get/Set bag.
type BizConfigReader interface {
	String(ctx context.Context, key, fallback string) string
	Bool(ctx context.Context, key string, fallback bool) bool
	Int(ctx context.Context, key string, fallback int) int
}
