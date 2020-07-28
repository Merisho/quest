package admin

import (
	"fmt"
	"github.com/merisho/quest/commandbus"
	"io"
	"log"
)

func NewAdmin(userID string, cb *commandbus.CommandBus, out io.Writer) *Admin {
	a := &Admin{
		destroy: make(chan struct{}),
		out: out,
	}
	onlyOtherUsers := func(c *commandbus.Command) bool {
		return c.UserID != userID
	}

	playerResponses := cb.FilterSubscribe("userres", onlyOtherUsers)
	playerMsgs := cb.FilterSubscribe("", onlyOtherUsers)
	playerAnswers := cb.FilterSubscribe("a", onlyOtherUsers)

	a.forwards = a.combineIntoForwards(playerResponses, playerMsgs, playerAnswers)
	a.handleForwards()

	return a
}

type Admin struct {
	forwards chan *commandbus.Command
	destroy chan struct{}
	out io.Writer
}

func (a *Admin) Destroy() {
	close(a.destroy)
}

func (a *Admin) Greeting() {
	a.write("Hello, admin")
}

func (a *Admin) combineIntoForwards(playerRes, playerMsgs, playerAns chan *commandbus.Command) chan *commandbus.Command {
	f := make(chan *commandbus.Command)

	go func() {
		for {
			select {
			case c := <- playerRes:
				f <- c
			case c := <- playerMsgs:
				f <- c
			case c := <- playerAns:
				f <- c
			case <- a.destroy:
				close(a.forwards)
				return
			}
		}
	}()

	return f
}

func (a *Admin) handleForwards() {
	go func() {
		for c := range a.forwards {
			msg := fmt.Sprintf("%s\n=====\n", c.ServiceData["senderName"])
			if c.Type == "a" {
				msg += "!Answer: "
			}

			msg += c.Input
			a.write(msg)
		}
	}()
}

func (a *Admin) write(msg string) {
	_, err := a.out.Write([]byte(msg))
	if err != nil {
		log.Println(err)
	}
}
