package quest

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockOut struct {
	outs [][]byte
}

func (o *MockOut) Write(b []byte) (int, error) {
	o.outs = append(o.outs, b)
	return len(b), nil
}

func (o *MockOut) Last() string {
	if len(o.outs) == 0 {
		return ""
	}

	return string(o.outs[len(o.outs) - 1])
}

func (o *MockOut) OffsetLast(offset int) string {
	if len(o.outs) - 1 - offset < 0 {
		return ""
	}

	return string(o.outs[len(o.outs) - 1 - offset])
}

func (o *MockOut) First() string {
	return string(o.outs[0])
}

func TestNonexistentFile(t *testing.T) {
	q, err := NewQuestFromFile("./non/existent/file.json", nil)
	assert.Error(t, err)
	assert.Nil(t, q)
}

func TestInvalidJSONFile(t *testing.T) {
	q, err := NewQuestFromFile("./invalid.json", nil)
	assert.Error(t, err)
	assert.Nil(t, q)
}

func TestConstructsMissionsFromJSONFile(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Fatal(e)
		}
	}()

	q, err := NewQuestFromFile("./quest.json", &MockOut{})
	assert.NoError(t, err)
	assert.NotNil(t, q)

	assert.Equal(t, 2, q.MissionCount())
}

func TestCorrectMissionsLinking(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Fatal(e)
		}
	}()

	q, _ := NewQuestFromFile("./quest.json", &MockOut{})

	q.Start()
	assert.Equal(t, "Mission 1", q.MissionName())

	q.CompleteMission()
	assert.Equal(t, "Mission 2", q.MissionName())

	q.CompleteMission()
	assert.True(t, q.Finished())
}

func TestOutputStartMessage(t *testing.T) {
	out := &MockOut{}
	q, _ := NewQuestFromFile("./quest.json", out)

	q.Start()

	assert.Equal(t, "Welcome to Mission 1", out.First())
	assert.Equal(t, "2 + 2", out.Last())

	q.Answer("four")

	assert.Equal(t, "4 + 4", out.Last())
}

func TestOutputEndMessage(t *testing.T) {
	out := &MockOut{}
	q, _ := NewQuestFromFile("./quest.json", out)

	q.Start()
	q.Answer("four")

	assert.Equal(t, "Mission 1 successfully completed", out.OffsetLast(2))
}

func TestAnswer(t *testing.T) {
	q, _ := NewQuestFromFile("./quest.json", &MockOut{})

	q.Start()

	assert.False(t, q.Answer("3"))
	assert.True(t, q.Answer("Four"))
	assert.Equal(t, "Mission 2", q.MissionName())
}
