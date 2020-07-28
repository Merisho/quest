package commandbus

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseCommand(t *testing.T) {
	cb := NewCommandBus()

	cmd := cb.Parse("/test command arguments")

	assert.Equal(t, CommandType("test"), cmd.Type)
	assert.Equal(t, "command arguments", cmd.Input)
	assert.Equal(t, 2, len(cmd.Args))
	assert.Equal(t, "command", cmd.Args[0])
	assert.Equal(t, "arguments", cmd.Args[1])
}

func TestParseArbitraryCommand(t *testing.T) {
	cb := NewCommandBus()

	cmd := cb.Parse("some arbitrary text")

	assert.Equal(t, CommandType(""), cmd.Type)
	assert.Equal(t, "some arbitrary text", cmd.Input)
	assert.Equal(t, 0, len(cmd.Args))

	assert.NotPanics(t, func() {
		cmd := cb.Parse("")
		assert.Nil(t, cmd)
	})
}

func TestSubscribe(t *testing.T) {
	cb := NewCommandBus()

	sub1 := cb.Subscribe("test")
	sub2 := cb.Subscribe("test")

	time.AfterFunc(0, func() {
		cb.Publish("/test command", "user-id")
	})

	cmd := <- sub1
	assert.Equal(t, CommandType("test"), cmd.Type)
	assert.Equal(t, "command", cmd.Input)
	assert.Equal(t, "command", cmd.Args[0])
	assert.Equal(t, "user-id", cmd.UserID)

	cmd = <- sub2
	assert.Equal(t, CommandType("test"), cmd.Type)
	assert.Equal(t, "command", cmd.Input)
	assert.Equal(t, "command", cmd.Args[0])
	assert.Equal(t, "user-id", cmd.UserID)
}

func TestDanglingSubscription(t *testing.T) {
	cb := NewCommandBus()

	cb.Subscribe("test")
	sub2 := cb.Subscribe("test")

	time.AfterFunc(0, func() {
		cb.Publish("/test command", "user-id")
	})

	select {
	case c := <- sub2:
		assert.Equal(t, "user-id", c.UserID)
	case <- time.After(50 * time.Millisecond):
		assert.Fail(t, "sends to subscriptions are blocked")
	}

	assert.Equal(t, 1, cb.SubscriptionsCount("test"), "must remove first subscription which is dangling")
}

func TestFilterSubscribe(t *testing.T) {
	cb := NewCommandBus()

	sub := cb.FilterSubscribe("", func(cmd *Command) bool {
		return cmd.UserID == "user-1"
	})

	time.AfterFunc(0, func() {
		cb.Publish("arbitrary text", "user-2")
	})

	select {
	case <- sub:
		assert.Fail(t, "must NOT receive command due to filter")
	case <- time.After(50 * time.Millisecond):
	}

	time.AfterFunc(0, func() {
		cb.Publish("arbitrary text", "user-1")
	})

	select {
	case <- sub:
	case <- time.After(50 * time.Millisecond):
		assert.Fail(t, "must receive command which is passed by filter")
	}
}

func TestIgnoreEmptyCommand(t *testing.T) {
	cb := NewCommandBus()

	sub := cb.Subscribe("")

	time.AfterFunc(0, func() {
		assert.NotPanics(t, func() {
			cb.Publish("", "user-id")
		})
	})

	select {
	case <- sub:
		assert.Fail(t, "must not receive the empty command")
	case <- time.After(10 * time.Millisecond):
	}
}
