package pubsub

import (
	"sync"
)

// APACHE LICENSE NOTIFICATION
// Portions of this code extracted from github.com/filecoin-project/go-data-transfer
// Copyright 2019. Protocol Labs

// SubscriberFn is a function that receives events from a dispatcher
type SubscriberFn[Event any] func(event Event) error

// Unsubscribe is a function returned from subscribe that can be used to terminate
// the subscription
type Unsubscribe func()

type subscriber[Event any] struct {
	fn  SubscriberFn[Event]
	key uint64
}

// PubSub is a simple emitter of data transfer events
type PubSub[Event any] struct {
	subscribersLk sync.RWMutex
	subscribers   []subscriber[Event]
	nextKey       uint64
}

// New returns a new PubSub
func New[Event any]() *PubSub[Event] {
	return &PubSub[Event]{}
}

// Subscribe adds the given subscriber to the list of subscribers for this Pubsub
func (ps *PubSub[Event]) Subscribe(subscriberFn SubscriberFn[Event]) Unsubscribe {
	ps.subscribersLk.Lock()
	subscriber := subscriber[Event]{subscriberFn, ps.nextKey}
	ps.nextKey++
	ps.subscribers = append(ps.subscribers, subscriber)
	ps.subscribersLk.Unlock()
	return ps.unsubscribeAt(subscriber)
}

// unsubscribeAt returns a function that removes an item from ps.subscribers. Does not preserve order.
// Subsequent, repeated calls to the func with the same Subscriber are a no-op.
func (ps *PubSub[Event]) unsubscribeAt(sub subscriber[Event]) Unsubscribe {
	return func() {
		ps.subscribersLk.Lock()
		defer ps.subscribersLk.Unlock()
		curLen := len(ps.subscribers)
		for i, el := range ps.subscribers {
			if sub.key == el.key {
				ps.subscribers[i] = ps.subscribers[curLen-1]
				ps.subscribers = ps.subscribers[:curLen-1]
				return
			}
		}
	}
}

// Publish publishes the given event and channel state to all subscribers
func (ps *PubSub[Event]) Publish(event Event) error {
	ps.subscribersLk.RLock()
	defer ps.subscribersLk.RUnlock()
	for _, sub := range ps.subscribers {
		err := sub.fn(event)
		if err != nil {
			return err
		}
	}
	return nil
}
