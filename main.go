package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/merisho/quest/admin"
	"github.com/merisho/quest/commandbus"
	"github.com/merisho/quest/player"
	"github.com/merisho/quest/quest"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

func NewTGBotOut(id int64, bot *tgbotapi.BotAPI, cb *commandbus.CommandBus) *TGBotOut {
	return &TGBotOut{
		id: id,
		bot: bot,
		cb: cb,
	}
}

type TGBotOut struct {
	id int64
	bot *tgbotapi.BotAPI
	cb *commandbus.CommandBus
}

func (t *TGBotOut) Write(b []byte) (int, error) {
	msg := tgbotapi.NewMessage(t.id, string(b))
	_, err := t.bot.Send(msg)
	if err != nil {
		return 0, err
	}

	botName := fmt.Sprintf("%s %s", t.bot.Self.FirstName, t.bot.Self.LastName)
	t.cb.Publish("/userres " + string(b), strconv.FormatInt(t.id, 10), [2]string{"senderName", botName})

	return len(b), nil
}

func main() {
	if len(os.Args) < 2 {
		panic("please supply Telegram bot token as the first argument")
	}
	tgBotToken := os.Args[1]
	bot, err := tgbotapi.NewBotAPI(tgBotToken)
	if err != nil {
		panic(err) // You should add better error handling than this!
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	var p *player.Player
	var a *admin.Admin
	commands := commandbus.NewCommandBus()
	for update := range updates {
		msg := update.Message
		if msg == nil {
			continue
		}

		id := msg.Chat.ID

		if msg.IsCommand() {
			if msg.Command() == "philadelphia" {
				if p != nil {
					p.Destroy()
				}

				out := NewTGBotOut(id, bot, commands)
				p = initPlayer(id, commands, out)
				continue
			}

			if msg.Command() == "adminsecret" {
				if a != nil {
					a.Destroy()
				}

				out := NewTGBotOut(id, bot, commands)
				a = admin.NewAdmin(strconv.FormatInt(id, 10), commands, out)
				a.Greeting()
				continue
			}
		}

		fullName := fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName)
		commands.Publish(update.Message.Text, strconv.FormatInt(id, 10), [2]string{"senderName", fullName})
	}
}

type PlayerConfig struct {
	IntroMessage string `json:"introMessage"`
	OutroMessage string `json:"outroMessage"`
	OutroMessageDelay time.Duration `json:"outroMessageDelay"`
	WrongAnswersForClue int `json:"wrongAnswersForClue"`
}

func initPlayer(id int64, cb *commandbus.CommandBus, out *TGBotOut) *player.Player {
	parsedConf := parsePlayerConfig("./config.json")

	userID := strconv.FormatInt(id, 10)
	conf := player.Config{
		UserID: userID,
		WrongAnswersForClue: parsedConf.WrongAnswersForClue,
		IntroMessage: parsedConf.IntroMessage,
		OutroMessage: parsedConf.OutroMessage,
		OutroMessageDelay: parsedConf.OutroMessageDelay,
	}

	q, err := quest.NewQuestFromFile("./quest.json", out)
	if err != nil {
		log.Println(err)
		return nil
	}

	return player.NewPlayer(conf, cb, q, out)
}

func parsePlayerConfig(path string) PlayerConfig {
	var msgs PlayerConfig

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("could not find messages file:", err)
		return msgs
	}

	err = json.Unmarshal(b, &msgs)
	if err != nil {
		log.Println("could not parse messages:", err)
		return msgs
	}

	return msgs
}
