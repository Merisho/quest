package commandbus

import (
	"strings"
)

type CommandType string
type Subscription chan *Command

func NewCommandBus() *CommandBus {
	return &CommandBus{
		subs: NewSubscriptionsMap(),
	}
}

type CommandBus struct {
	subs *SubscriptionsMap
}

func (cb *CommandBus) SubscriptionsCount(cmdType CommandType) int {
	return cb.subs.Count(cmdType)
}

func (cb *CommandBus) Subscribe(cmdType CommandType) chan *Command {
	sub := cb.subs.Create(cmdType)
	return sub
}

func (cb *CommandBus) FilterSubscribe(cmdType CommandType, filter func(*Command) bool) chan *Command {
	filteredSub := make(chan *Command)
	sub := cb.Subscribe(cmdType)

	go func() {
		for cmd := range sub {
			if filter(cmd) {
				filteredSub <- cmd
			}
		}
	}()

	return filteredSub
}

func (cb *CommandBus) Publish(text, userID string, serviceData ...[2]string) {
	cmd := cb.Parse(text)
	if cmd == nil {
		return
	}

	cmd.UserID = userID

	data := make(map[string]string)
	for _, sd := range serviceData {
		data[sd[0]] = sd[1]
	}

	cmd.ServiceData = data

	cb.subs.Send(cmd)
}

func (cb *CommandBus) Parse(text string) *Command {
	if text == "" {
		return nil
	}

	if cb.isNotCommand(text) {
		return &Command{
			Input: text,
		}
	}

	strs := strings.Split(text, " ")
	t := strings.Replace(strs[0], "/", "", 1)
	args := strs[1:]
	input := strings.Join(args, " ")

	return &Command{
		Type: CommandType(t),
		Input: input,
		Args: args,
	}
}

func (cb *CommandBus) isNotCommand(text string) bool {
	return text == "" || text[0] != '/'
}

type Command struct {
	Type CommandType
	Input string
	Args []string
	UserID string
	ServiceData map[string]string
}
