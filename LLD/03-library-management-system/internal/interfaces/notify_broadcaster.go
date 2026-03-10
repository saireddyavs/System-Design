package interfaces

// NotifyBroadcaster broadcasts notifications to all registered channels (Observer Subject)
// Used by services to decouple from concrete NotificationManager (DIP)
type NotifyBroadcaster interface {
	NotifyAll(payload NotificationPayload)
}
