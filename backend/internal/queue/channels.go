package queue

// Redis Pub/Sub channel names. Group by sub-domain and prefix with the service name.
const (
	ChannelAudit                  = "app:audit"
	ChannelNotifications          = "app:notifications"
	ChannelSearchSync             = "app:search.sync"
	ChannelPermissionInvalidate   = "app:permission.invalidate"
	ChannelSubscriptionInvalidate = "app:subscription.invalidate"
	ChannelInviteEmail            = "app:invite.email"
	ChannelPasswordResetEmail     = "app:password_reset.email"
	ChannelUserWelcomeEmail       = "app:user.welcome.email"
)

// DefaultChannels returns channels the worker subscribes to by default.
func DefaultChannels() []string {
	return []string{
		ChannelAudit,
		ChannelNotifications,
		ChannelSearchSync,
		ChannelPermissionInvalidate,
		ChannelSubscriptionInvalidate,
		ChannelInviteEmail,
		ChannelPasswordResetEmail,
		ChannelUserWelcomeEmail,
	}
}
