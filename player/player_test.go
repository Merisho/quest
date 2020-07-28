package player

import (
	"bytes"
	"github.com/merisho/quest/commandbus"
	"github.com/merisho/quest/quest"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

const (
	userID = "user-id"
	questFile = "./test-quest.json"
)

var conf = Config{
	UserID: userID,
	WrongAnswersForClue: 3,
	IntroMessage: "intro message",
	OutroMessage: "outro message",
}

func TestHandleAnswer(t *testing.T) {
	cb := commandbus.NewCommandBus()
	buf := bytes.NewBuffer(nil)
	q, _ := quest.NewQuestFromFile(questFile, buf)

	player := NewPlayer(conf, cb, q, buf)
	time.Sleep(1)

	assert.Equal(t, "Mission 1", player.MissionName())

	cb.Publish("/a answer 1", userID)
	time.Sleep(1)

	assert.Equal(t, "Mission 2", player.MissionName())

	cb.Publish("/a answer 2", userID)
	time.Sleep(1)

	assert.True(t, player.Finished())
}

func TestClueAfter3WrongAnswers(t *testing.T) {
	cb := commandbus.NewCommandBus()
	buf := bytes.NewBuffer(nil)

	q, _ := quest.NewQuestFromFile(questFile, buf)
	NewPlayer(conf, cb, q, buf)

	cb.Publish("/a 111", userID)
	time.Sleep(1)
	cb.Publish("/a 111", userID)
	time.Sleep(1)
	cb.Publish("/a 111", userID)
	time.Sleep(1)

	str := "clue 1"
	assert.True(t, strings.Contains(buf.String(), str), buf.String())
}

func TestDestroyPlayer(t *testing.T) {
	cb := commandbus.NewCommandBus()
	buf := bytes.NewBuffer(nil)

	q, _ := quest.NewQuestFromFile(questFile, buf)
	p := NewPlayer(conf, cb, q, buf)

	time.Sleep(1)

	m := p.MissionName()

	p.Destroy()
	time.Sleep(1)

	cb.Publish("/a answer 1", userID)
	time.Sleep(1)

	assert.Equal(t, m, p.MissionName(), "must NOT handle anything after it is destroyed")
}

func TestIntroMessage(t *testing.T) {
	cb := commandbus.NewCommandBus()
	buf := bytes.NewBuffer(nil)

	q, _ := quest.NewQuestFromFile(questFile, buf)
	NewPlayer(conf, cb, q, buf)

	assert.True(t, strings.Contains(buf.String(), conf.IntroMessage))
}

func TestOutroMessage(t *testing.T) {
	cb := commandbus.NewCommandBus()
	buf := bytes.NewBuffer(nil)

	q, _ := quest.NewQuestFromFile(questFile, buf)
	NewPlayer(conf, cb, q, buf)

	cb.Publish("/a answer 1", userID)
	time.Sleep(1)
	cb.Publish("/a answer 2", userID)
	time.Sleep(1)

	assert.True(t, strings.Contains(buf.String(), conf.OutroMessage))
}

func TestAdminMessages(t *testing.T) {
	cb := commandbus.NewCommandBus()
	buf := bytes.NewBuffer(nil)

	q, _ := quest.NewQuestFromFile(questFile, buf)
	NewPlayer(conf, cb, q, buf)
	time.Sleep(1)

	cb.Publish("/adminmsg Hello from admin", "admin-id")
	time.Sleep(1)

	assert.True(t, strings.Contains(buf.String(), "Hello from admin"))
}
