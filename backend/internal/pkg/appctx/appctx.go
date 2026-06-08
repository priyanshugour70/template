// Package appctx provides typed accessors for request-scoped values that flow
// from auth middleware → context.Context → repositories. Keeps stringly keys
// out of consumer code and provides a single source of truth.
package appctx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

const (
	keyUser ctxKey = iota + 1
	keyTenant
	keyOrg
	keyEmail
	keyIP
	keyUserAgent
	keyRoles
	keyJTI
	keyMembership
	keySuperAdmin
)

// Principal is the resolved identity for the current request.
type Principal struct {
	UserID         uuid.UUID
	TenantID       uuid.UUID
	OrganizationID uuid.UUID
	MembershipID   uuid.UUID
	Email          string
	IP             string
	UserAgent      string
	Roles          []string
	JTI            string
	IsSuperAdmin   bool
}

func (p Principal) IsZero() bool { return p.UserID == uuid.Nil }

func With(ctx context.Context, p Principal) context.Context {
	ctx = context.WithValue(ctx, keyUser, p.UserID)
	ctx = context.WithValue(ctx, keyTenant, p.TenantID)
	ctx = context.WithValue(ctx, keyOrg, p.OrganizationID)
	ctx = context.WithValue(ctx, keyMembership, p.MembershipID)
	ctx = context.WithValue(ctx, keyEmail, p.Email)
	ctx = context.WithValue(ctx, keyIP, p.IP)
	ctx = context.WithValue(ctx, keyUserAgent, p.UserAgent)
	ctx = context.WithValue(ctx, keyRoles, p.Roles)
	ctx = context.WithValue(ctx, keyJTI, p.JTI)
	ctx = context.WithValue(ctx, keySuperAdmin, p.IsSuperAdmin)
	return ctx
}

func UserID(ctx context.Context) uuid.UUID {
	v, _ := ctx.Value(keyUser).(uuid.UUID)
	return v
}

func TenantID(ctx context.Context) uuid.UUID {
	v, _ := ctx.Value(keyTenant).(uuid.UUID)
	return v
}

func OrganizationID(ctx context.Context) uuid.UUID {
	v, _ := ctx.Value(keyOrg).(uuid.UUID)
	return v
}

func MembershipID(ctx context.Context) uuid.UUID {
	v, _ := ctx.Value(keyMembership).(uuid.UUID)
	return v
}

func Email(ctx context.Context) string {
	v, _ := ctx.Value(keyEmail).(string)
	return v
}

func IP(ctx context.Context) string {
	v, _ := ctx.Value(keyIP).(string)
	return v
}

func UserAgent(ctx context.Context) string {
	v, _ := ctx.Value(keyUserAgent).(string)
	return v
}

func Roles(ctx context.Context) []string {
	v, _ := ctx.Value(keyRoles).([]string)
	return v
}

func JTI(ctx context.Context) string {
	v, _ := ctx.Value(keyJTI).(string)
	return v
}

func IsSuperAdmin(ctx context.Context) bool {
	v, _ := ctx.Value(keySuperAdmin).(bool)
	return v
}

// Principal returns the full principal reconstructed from ctx.
func PrincipalFrom(ctx context.Context) Principal {
	return Principal{
		UserID:         UserID(ctx),
		TenantID:       TenantID(ctx),
		OrganizationID: OrganizationID(ctx),
		MembershipID:   MembershipID(ctx),
		Email:          Email(ctx),
		IP:             IP(ctx),
		UserAgent:      UserAgent(ctx),
		Roles:          Roles(ctx),
		JTI:            JTI(ctx),
		IsSuperAdmin:   IsSuperAdmin(ctx),
	}
}
