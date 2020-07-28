package player

import (
	"github.com/merisho/quest/commandbus"
	"github.com/merisho/quest/quest"
	"io"
	"log"
	"time"
)

type Config struct {
	UserID string
	WrongAnswersForClue int
	IntroMessage string
	OutroMessage string
	OutroMessageDelay time.Duration
}

func NewPlayer(conf Config, cb *commandbus.CommandBus, q *quest.Quest, out io.Writer) *Player {
	onlyThisUser := func(command *commandbus.Command) bool {
		return command.UserID == conf.UserID
	}
	onlyOtherUsers := func(command *commandbus.Command) bool {
		return command.UserID != conf.UserID
	}

	answerSub := cb.FilterSubscribe("a", onlyThisUser)
	adminMsgs := cb.FilterSubscribe("adminmsg", onlyOtherUsers)

	p := &Player{
		cb: cb,
		userID: conf.UserID,
		answer: answerSub,
		adminMsgs: adminMsgs,
		quest: q,
		tries: make(map[string]int),
		triesForClue: conf.WrongAnswersForClue,
		out: out,
		destroy: make(chan struct{}),
		introMsg: conf.IntroMessage,
		outroMsg: conf.OutroMessage,
		outroMsgDelay: conf.OutroMessageDelay,
	}

	p.IntroMessage()

	q.Start()

	p.handleCommands()

	return p
}

type Player struct {
	cb *commandbus.CommandBus
	userID string
	answer chan *commandbus.Command
	adminMsgs chan *commandbus.Command
	quest *quest.Quest
	tries map[string]int
	triesForClue int
	out io.Writer
	destroy chan struct{}
	introMsg string
	outroMsg string
	outroMsgDelay time.Duration
}

func (p *Player) UserID() string {
	return p.userID
}

func (p *Player) handleCommands() {
	go func() {
		for {
			select {
			case a := <- p.answer:
				p.Answer(a)
			case msg := <- p.adminMsgs:
				p.Write(msg.Input)
			case <- p.destroy:
				return
			}
		}
	}()
}

func (p *Player) MissionName() string {
	return p.quest.MissionName()
}

func (p *Player) Answer(ans *commandbus.Command) {
	if p.Finished() {
		return
	}

	res := p.quest.Answer(ans.Input)
	if !res {
		p.tries[p.MissionName()]++
		if p.tries[p.MissionName()] == p.triesForClue {
			p.Write(p.quest.Clue())
		}
	}

	if p.Finished() {
		p.OutroMessage()
	}
}

func (p *Player) Write(s string) {
	_, err := p.out.Write([]byte(s))
	if err != nil {
		log.Println(err)
	}
}

func (p *Player) Finished() bool {
	return p.quest.Finished()
}

func (p *Player) Destroy() {
	close(p.destroy)
}

func (p *Player) IntroMessage() {
	if p.introMsg != "" {
		p.Write(p.introMsg)
	}
}

func (p *Player) OutroMessage() {
	if p.outroMsg != "" {
		time.Sleep(p.outroMsgDelay * time.Second)
		p.Write(p.outroMsg)
	}
}
