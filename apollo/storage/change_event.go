package storage

const (
	ADDED ConfigChageType = iota
	MODIFIED
	DELETED
)

type ConfigChageType int

type ChangeListener interface {
	OnChange(event *ChangeEvent)
	OnNewestChange(event *FullChangeEvent)
}

type baseChangeEvent struct {
	Namespace      string
	NotificationID int64
}

type ChangeEvent struct {
	baseChangeEvent
	Changes map[string]*ConfigChange
}

type ConfigChange struct {
	OldValue   interface{}
	NewValue   interface{}
	ChangeType ConfigChageType
}

type FullChangeEvent struct {
	baseChangeEvent
	Changes map[string]interface{}
}

func createModifyConfigChange(oldValue interface{}, newValue interface{}) *ConfigChange {
	return &ConfigChange{
		OldValue:   oldValue,
		NewValue:   newValue,
		ChangeType: MODIFIED,
	}
}

func createAddConfigChange(newValue interface{}) *ConfigChange {
	return &ConfigChange{
		NewValue:   newValue,
		ChangeType: ADDED,
	}
}

func createDeletedConfigChange(oldValue interface{}) *ConfigChange {
	return &ConfigChange{
		OldValue:   oldValue,
		ChangeType: DELETED,
	}
}

func createConfigChangeEvent(changes map[string]*ConfigChange, namespace string, notificationID int64) *ChangeEvent {
	c := &ChangeEvent{
		Changes: changes,
	}
	c.Namespace = namespace
	c.NotificationID = notificationID
	return c
}
