package quest

import (
	"encoding/json"
	"errors"
	"github.com/merisho/quester"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type QuestDescriptions []QuestDescription

func (qd QuestDescriptions) buildQuesterMissions(out io.Writer) []quester.Mission {
	var missions []quester.Mission
	for i := 0; i < len(qd) - 1; i++ {
		m := qd[i].buildQuesterMission(out)
		m.Next = qd[i + 1].Name

		missions = append(missions, m)
	}

	return append(missions, qd[len(qd) - 1].buildQuesterMission(out))
}

type QuestDescription struct {
	Name string `json:"name"`
	MissionStartMessage string `json:"missionStartMessage"`
	MissionStartDelay time.Duration `json:"missionStartDelay"`
	MissionEndMessage   string `json:"missionEndMessage"`
	Task TaskDescription `json:"task"`
}

func (qd QuestDescription) buildQuesterMission(out io.Writer) quester.Mission {
	return quester.Mission{
		Name: qd.Name,
		Tasks: quester.Tasks{
			{
				Statement: qd.Task.Statement,
				Clue: qd.Task.Clue,
				Resolve: func(answer string) bool {
					return strings.ToLower(answer) == qd.Task.CorrectAnswer
				},
			},
		},
		Start: func() {
			if qd.MissionStartMessage != "" {
				time.Sleep(qd.MissionStartDelay * time.Second)
				if _, err := out.Write([]byte(qd.MissionStartMessage)); err != nil {
					log.Println(err)
				}
			}

			time.Sleep(qd.Task.StatementDelay * time.Second)

			if _, err := out.Write([]byte(qd.Task.Statement)); err != nil {
				log.Println(err)
			}
		},
		End: func() {
			if qd.MissionEndMessage == "" {
				return
			}

			_, err := out.Write([]byte(qd.MissionEndMessage))
			if err != nil {
				log.Println(err)
			}
		},
	}
}

type TaskDescription struct {
	Statement string `json:"statement"`
	StatementDelay time.Duration `json:"statementDelay"`
	Clue string `json:"clue"`
	CorrectAnswer string `json:"correctAnswer"`
}

func NewQuestFromFile(path string, out io.Writer) (*Quest, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var descr QuestDescriptions
	err = json.Unmarshal(b, &descr)
	if err != nil {
		return nil, errors.New("invalid JSON: " + err.Error())
	}

	q := constructQuest(descr, out)

	return q, nil
}

func constructQuest(descr QuestDescriptions, out io.Writer) *Quest {
	q := &Quest{
		qst: quester.NewQuest(),
		out: out,
	}

	missions := descr.buildQuesterMissions(out)
	for _, m := range missions {
		q.qst.AddMission(m)
	}

	return q
}

type Quest struct {
	out io.Writer
	qst *quester.Quest
}

func (q *Quest) MissionCount() int {
	return q.qst.Length()
}

func (q *Quest) Start() {
	q.qst.Start()
}

func (q *Quest) MissionName() string {
	return q.qst.Current().Name
}

func (q *Quest) CompleteMission() {
	q.qst.PassCurrent()
}

func (q *Quest) Answer(ans string) bool {
	res, _ := q.qst.ResolveCurrentTask(ans)
	return res
}

func (q *Quest) Finished() bool {
	return q.qst.IsFinished()
}

func (q *Quest) Clue() string {
	return q.qst.Current().Clue()
}
