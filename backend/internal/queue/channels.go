package queue

// Redis Pub/Sub channel names. Group by sub-domain and prefix with the service name.
const (
	ChannelAudit         = "app:audit"
	ChannelNotifications = "app:notifications"
	ChannelSearchSync    = "app:search.sync"
)

// DefaultChannels returns channels the worker subscribes to by default.
func DefaultChannels() []string {
	return []string{ChannelAudit, ChannelNotifications, ChannelSearchSync}
}
