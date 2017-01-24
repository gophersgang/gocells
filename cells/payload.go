// Tideland Go Cells - Payload
//
// Copyright (C) 2010-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package cells

//--------------------
// IMPORTS
//--------------------

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

//--------------------
// PAYLOAD
//--------------------

// PayloadValues is intended to set and get the information
// of a payload as bulk.
type PayloadValues map[string]interface{}

// Payload is a write-once/read-multiple container for the
// transport of additional information with events. In case
// one item is a reference type it's in the responsibility
// of the users to avoid concurrent changes of their values.
type Payload interface {
	fmt.Stringer

	// Len returns the number of values.
	Len() int

	// Get returns one of the payload values.
	Get(key string) (interface{}, bool)

	// GetBool returns one of the payload values
	// as bool. If it's no bool false is returned.
	GetBool(key string) (bool, bool)

	// GetInt returns one of the payload values
	// as int. If it's no int false is returned.
	GetInt(key string) (int, bool)

	// GetFloat64 returns one of the payload values
	// as float64. If it's no float64 false is returned.
	GetFloat64(key string) (float64, bool)

	// GetString returns one of the payload values
	// as string. If it's no string false is returned.
	GetString(key string) (string, bool)

	// GetTime returns one of the payload values
	// as time.Time. If it's no time false is returned.
	GetTime(key string) (time.Time, bool)

	// GetDuration returns one of the payload values as
	// time.Duration. If it's no duration false is returned.
	GetDuration(key string) (time.Duration, bool)

	// GetWaiter returns a payload waiter to be used
	// for answering with a payload.
	GetWaiter(key string) (PayloadWaiter, bool)

	// Keys return all keys of the payload.
	Keys() []string

	// Do iterates a function over all keys and values.
	Do(f func(key string, value interface{}) error) error

	// Apply creates a new payload containing the values
	// of this one and the passed values. Allowed are
	// PayloadValues, map[string]interface{}, and any
	// other single value. The latter will be stored
	// with the cells.DefaultPayload key. Values of this
	// payload are overwritten by those which are passed
	// if they share the key.
	Apply(values interface{}) Payload
}

// payload implements the Payload interface.
type payload struct {
	values PayloadValues
}

// NewPayload creates a new payload containing the passed
// values. In case of a Payload this is used directly, in
// case of a PayloadValues or a map[string]interface{} their
// content is used, and when passing any other type the
// value is stored with the key cells.DefaultPayload.
func NewPayload(values interface{}) Payload {
	if p, ok := values.(Payload); ok {
		return p
	}
	p := &payload{
		values: PayloadValues{},
	}
	switch vs := values.(type) {
	case PayloadValues:
		for key, value := range vs {
			p.values[key] = value
		}
	case map[string]interface{}:
		for key, value := range vs {
			p.values[key] = value
		}
	default:
		p.values[DefaultPayload] = values
	}
	return p
}

// Len implementes the Payload interface.
func (p *payload) Len() int {
	return len(p.values)
}

// Get implementes the Payload interface.
func (p *payload) Get(key string) (interface{}, bool) {
	value, ok := p.values[key]
	return value, ok
}

// GetBool implementes the Payload interface.
func (p *payload) GetBool(key string) (bool, bool) {
	raw, ok := p.Get(key)
	if !ok {
		return false, ok
	}
	value, ok := raw.(bool)
	return value, ok
}

// GetInt implementes the Payload interface.
func (p *payload) GetInt(key string) (int, bool) {
	raw, ok := p.Get(key)
	if !ok {
		return 0, ok
	}
	value, ok := raw.(int)
	return value, ok
}

// GetFloat64 implementes the Payload interface.
func (p *payload) GetFloat64(key string) (float64, bool) {
	raw, ok := p.Get(key)
	if !ok {
		return 0.0, ok
	}
	value, ok := raw.(float64)
	return value, ok
}

// GetString implementes the Payload interface.
func (p *payload) GetString(key string) (string, bool) {
	raw, ok := p.Get(key)
	if !ok {
		return "", ok
	}
	value, ok := raw.(string)
	return value, ok
}

// GetTime implementes the Payload interface.
func (p *payload) GetTime(key string) (time.Time, bool) {
	raw, ok := p.Get(key)
	if !ok {
		return time.Time{}, ok
	}
	value, ok := raw.(time.Time)
	return value, ok
}

// GetDuration implementes the Payload interface.
func (p *payload) GetDuration(key string) (time.Duration, bool) {
	raw, ok := p.Get(key)
	if !ok {
		return time.Duration(0), ok
	}
	value, ok := raw.(time.Duration)
	return value, ok
}

// GetWaiter implements the Payload interface.
func (p *payload) GetWaiter(key string) (PayloadWaiter, bool) {
	raw, ok := p.Get(key)
	if !ok {
		return nil, ok
	}
	value, ok := raw.(PayloadWaiter)
	return value, ok
}

// Keys is specified on the Payload interface.
func (p *payload) Keys() []string {
	keys := []string{}
	for key := range p.values {
		keys = append(keys, key)
	}
	return keys
}

// Do implementes the Payload interface.
func (p *payload) Do(f func(key string, value interface{}) error) error {
	for key, value := range p.values {
		if err := f(key, value); err != nil {
			return err
		}
	}
	return nil
}

// Apply implementes the Payload interface.
func (p *payload) Apply(values interface{}) Payload {
	applied := &payload{
		values: PayloadValues{},
	}
	for key, value := range p.values {
		applied.values[key] = value
	}
	switch vs := values.(type) {
	case Payload:
		vs.Do(func(key string, value interface{}) error {
			applied.values[key] = value
			return nil
		})
	case PayloadValues:
		for key, value := range vs {
			applied.values[key] = value
		}
	case map[string]interface{}:
		for key, value := range vs {
			applied.values[key] = value
		}
	default:
		applied.values[DefaultPayload] = values
	}
	return applied
}

// String returns the payload represented as string.
func (p *payload) String() string {
	ps := []string{}
	for key, value := range p.values {
		ps = append(ps, fmt.Sprintf("<%q: %v>", key, value))
	}
	return strings.Join(ps, ", ")
}

//--------------------
// PAYLOAD WAITER
//--------------------

// PayloadWaiter can be sent with an event as payload.
// Once a payload is set by a behavior Wait() continues
// and returns it.
type PayloadWaiter interface {
	// Set sets the payload somebody is waiting for.
	Set(p Payload)

	// Wait waits until the payload is set. A deadline
	// or timeout set by the context may cancel the
	// waiting.
	Wait(ctx context.Context) (Payload, error)
}

// payloadWaiter implements the PayloadWaiter interface.
type payloadWaiter struct {
	payloadc chan Payload
	once     sync.Once
}

// NewPayloadWaiter creates a new waiter for a payload
// returned by a behavior.
func NewPayloadWaiter() PayloadWaiter {
	return &payloadWaiter{
		payloadc: make(chan Payload, 1),
	}
}

// Set implements the PayloadWaiter interface.
func (w *payloadWaiter) Set(p Payload) {
	w.once.Do(func() {
		w.payloadc <- p
	})
}

// Wait implements the PayloadWaiter interface.
func (w *payloadWaiter) Wait(ctx context.Context) (Payload, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		select {
		case pl := <-w.payloadc:
			return pl, nil
		case <-ctx.Done():
			err := ctx.Err()
			return nil, err
		}
	}
}

// EOF
