package storage

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/f0resee/stdlib/apollo/component/log"
)

const (
	fmtInvalidKey = "invalid key format for key %s"
)

var (
	ErrNilListener = errors.New("nil listener")
)

type Event struct {
	EventType ConfigChageType
	Key       string
	Value     interface{}
}

type Listener interface {
	Event(evnet *Event)
}

type Dispatcher struct {
	listeners map[string][]Listener
}

func UseEventDispatch() *Dispatcher {
	eventDispatch := new(Dispatcher)
	eventDispatch.listeners = make(map[string][]Listener)
	return eventDispatch
}

func (d *Dispatcher) RegisterListener(listenerObject Listener, keys ...string) error {
	log.Infof("start add key %v add listener", keys)
	if listenerObject == nil {
		return ErrNilListener
	}

	for _, key := range keys {
		if invalidKey(key) {
			return fmt.Errorf(fmtInvalidKey, key)
		}

		listenerList, ok := d.listeners[key]
		if !ok {
			d.listeners[key] = make([]Listener, 0)
		}

		for _, listener := range listenerList {
			if listener == listenerObject {
				log.Infof("key %s has listener", key)
				return nil
			}
		}
		listenerList = append(listenerList, listenerObject)
		d.listeners[key] = listenerList
	}
	return nil
}

func invalidKey(key string) bool {
	_, err := regexp.Compile(key)
	return err != nil
}

func (d *Dispatcher) UnRegisterListener(listenerObj Listener, keys ...string) error {
	if listenerObj == nil {
		return ErrNilListener
	}

	for _, key := range keys {
		listenerList, ok := d.listeners[key]
		if !ok {
			continue
		}

		newListenerList := make([]Listener, 0)
		for _, listener := range listenerList {
			if listener == listenerObj {
				continue
			}
			newListenerList = append(newListenerList, listener)
		}
		d.listeners[key] = newListenerList
	}
	return nil
}

func (d *Dispatcher) OnChange(changeEvent *ChangeEvent) {
	if changeEvent == nil {
		return
	}
	log.Infof("get change event for namespace %s", changeEvent.Namespace)
	for key, event := range changeEvent.Changes {
		d.dispatchEvent(key, event)
	}
}

func (d *Dispatcher) OnNewestChange(event *FullChangeEvent) {}

func (d *Dispatcher) dispatchEvent(eventKey string, event *ConfigChange) {
	for regKey, listenerList := range d.listeners {
		matched, err := regexp.MatchString(regKey, eventKey)
		if err != nil {
			log.Errorf("regular expression for key %s, error: %v", eventKey, err)
			continue
		}
		if matched {
			for _, listener := range listenerList {
				log.Infof("event generated for %s key %s", regKey, eventKey)
				go listener.Event(convertToEvent(eventKey, event))
			}
		}
	}
}

func convertToEvent(key string, event *ConfigChange) *Event {
	e := &Event{
		EventType: event.ChangeType,
		Key:       key,
	}
	switch event.ChangeType {
	case ADDED:
		e.Value = event.NewValue
	case MODIFIED:
		e.Value = event.NewValue
	case DELETED:
		e.Value = event.OldValue
	}
	return e
}
