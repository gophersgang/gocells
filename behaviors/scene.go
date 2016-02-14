// Tideland Go Cells - Behaviors - Scene
//
// Copyright (C) 2015-2016 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package behaviors

//--------------------
// IMPORTS
//--------------------

import (
	"github.com/tideland/gocells/cells"
)

//--------------------
// SCENE BEHAVIOR
//--------------------

// sceneBehavior stores events in scenes.
type sceneBehavior struct {
	ctx cells.Context
}

// NewSceneBehavior creates a scene behavior that stores the payload
// of an event using the topic of an event as key in its scene. This
// way external code can wait for this topic as flag and fetch the
// value. It's not intended to use it as a standard behvior, even if
// it works. Instead it can be used in testing scenarios.
func NewSceneBehavior() cells.Behavior {
	return &sceneBehavior{}
}

// Init implements the Behavior interface.
func (b *sceneBehavior) Init(ctx cells.Context) error {
	b.ctx = ctx
	return nil
}

// Terminate implements the Behavior interface.
func (b *sceneBehavior) Terminate() error {
	return nil
}

// ProcessEvent stores and flags the event in the event scene.
// So other code parts using the same scene can wait for the signal.
func (b *sceneBehavior) ProcessEvent(event cells.Event) error {
	scn := event.Scene()
	if scn != nil {
		err := scn.StoreAndFlag(event.Topic(), event.Payload())
		if err != nil {
			return err
		}
	}
	return nil
}

// Recover implements the Behavior interface.
func (b *sceneBehavior) Recover(err interface{}) error {
	return nil
}

// EOF
