package event

type EventWriter interface {
	Publish(eventType string, payload any) error
	Close() error
}
