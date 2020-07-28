package commandbus

import "time"

func NewSubscriptionsMap() *SubscriptionsMap {
	m := &SubscriptionsMap{
		subs: make(map[CommandType][]Subscription),
	}

	m.startSubscriptionHandler()

	return m
}

type SubscriptionsMap struct {
	subs map[CommandType][]Subscription
	createSub chan newSubscription
	sendCmd chan *Command
	deleteSub chan danglingSubscription
	countSub chan countSubscriptions
}

func (m *SubscriptionsMap) Create(cmdType CommandType) Subscription {
	sub := make(Subscription)
	m.createSub <- newSubscription{cmdType, sub}

	return sub
}

func (m *SubscriptionsMap) Send(cmd *Command) {
	m.sendCmd <- cmd
}
func (m *SubscriptionsMap) Count(cmdType CommandType) int {
	req := countSubscriptions{
		cmdType: cmdType,
		count: make(chan int),
	}

	m.countSub <- req

	return <- req.count
}

func (m *SubscriptionsMap) startSubscriptionHandler() {
	m.createSub = make(chan newSubscription)
	m.sendCmd = make(chan *Command)
	m.deleteSub = make(chan danglingSubscription, 1)
	m.countSub = make(chan countSubscriptions)

	go func() {
		for {
			select {
			case newSub := <- m.createSub:
				m.create(newSub.cmdType, newSub.sub)
			case cmd := <- m.sendCmd:
				m.send(cmd)
			case danglingSub := <- m.deleteSub:
				m.delete(danglingSub.cmdType, danglingSub.sub)
			case countReq := <- m.countSub:
				countReq.count <- len(m.subs[countReq.cmdType])
			}
		}
	}()
}

func (m *SubscriptionsMap) create(cmdType CommandType, sub Subscription) {
	m.subs[cmdType] = append(m.subs[cmdType], sub)
}

func (m *SubscriptionsMap) send(cmd *Command) {
	for _, s := range m.subs[cmd.Type] {
		select {
		case s <- cmd:
		case <- time.After(20 * time.Millisecond):
			m.deleteSub <- danglingSubscription{cmd.Type, s}
			continue
		}
	}
}

func (m *SubscriptionsMap) delete(cmdType CommandType, sub Subscription) {
	idx := -1
	subs := m.subs[cmdType]
	for i, s := range subs {
		if s == sub {
			idx = i
			break
		}
	}

	if idx == -1 {
		return
	}

	m.subs[cmdType] = append(subs[:idx], subs[idx + 1:]...)
}

type newSubscription struct {
	cmdType CommandType
	sub Subscription
}

type danglingSubscription struct {
	cmdType CommandType
	sub Subscription
}

type countSubscriptions struct {
	cmdType CommandType
	count chan int
}
