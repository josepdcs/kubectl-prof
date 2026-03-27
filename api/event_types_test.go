package api

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseEvent(t *testing.T) {
	t.Run("Parse Error Event", func(t *testing.T) {
		eventStr := `{"type":"error","data":{"reason":"something went wrong"}}`
		event, err := ParseEvent(eventStr)
		assert.NoError(t, err)
		assert.IsType(t, &ErrorData{}, event)
		assert.Equal(t, "something went wrong", event.(*ErrorData).Reason)
	})

	t.Run("Parse Result Event", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		eventStr := fmt.Sprintf(`{"type":"result","data":{"time":"%s","result-type":"flamegraph","file":"result.svg"}}`, now.Format(time.RFC3339))
		event, err := ParseEvent(eventStr)
		assert.NoError(t, err)
		assert.IsType(t, &ResultData{}, event)
		assert.Equal(t, "result.svg", event.(*ResultData).File)
		assert.True(t, now.Equal(event.(*ResultData).Time.Local()))
	})

	t.Run("Parse Progress Event", func(t *testing.T) {
		eventStr := `{"type":"progress","data":{"stage":"started"}}`
		event, err := ParseEvent(eventStr)
		assert.NoError(t, err)
		assert.IsType(t, &ProgressData{}, event)
		assert.Equal(t, Started, event.(*ProgressData).Stage)
	})

	t.Run("Parse Notice Event", func(t *testing.T) {
		eventStr := `{"type":"notice","data":{"msg":"some notice"}}`
		event, err := ParseEvent(eventStr)
		assert.NoError(t, err)
		assert.IsType(t, &NoticeData{}, event)
		assert.Equal(t, "some notice", event.(*NoticeData).Msg)
	})

	t.Run("Parse Log Event", func(t *testing.T) {
		eventStr := `{"type":"log","data":{"level":"info","msg":"some log message"}}`
		event, err := ParseEvent(eventStr)
		assert.NoError(t, err)
		assert.IsType(t, &LogData{}, event)
		assert.Equal(t, "info", event.(*LogData).Level)
		assert.Equal(t, "some log message", event.(*LogData).Msg)
	})

	t.Run("Parse Invalid JSON", func(t *testing.T) {
		eventStr := `{"type":"error","data":{invalid}}`
		event, err := ParseEvent(eventStr)
		assert.Error(t, err)
		assert.Nil(t, event)
	})

	t.Run("Parse Unknown Event Type", func(t *testing.T) {
		eventStr := `{"type":"unknown","data":{"foo":"bar"}}`
		event, err := ParseEvent(eventStr)
		assert.NoError(t, err)
		assert.Nil(t, event)
	})

	t.Run("Concurrency race condition check", func(t *testing.T) {
		const goroutines = 20
		const iterations = 100
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer wg.Done()
				// Cada goroutine usa un tipo de evento y un valor diferente para verificar aislamiento
				var eventStr string
				var expectedReason string
				var expectedMsg string

				if id%2 == 0 {
					expectedReason = fmt.Sprintf("error-from-%d", id)
					eventStr = fmt.Sprintf(`{"type":"error","data":{"reason":"%s"}}`, expectedReason)
				} else {
					expectedMsg = fmt.Sprintf("log-from-%d", id)
					eventStr = fmt.Sprintf(`{"type":"log","data":{"level":"info","msg":"%s"}}`, expectedMsg)
				}

				for j := 0; j < iterations; j++ {
					event, err := ParseEvent(eventStr)
					assert.NoError(t, err)
					if id%2 == 0 {
						data, ok := event.(*ErrorData)
						assert.True(t, ok, "Expected *ErrorData")
						assert.Equal(t, expectedReason, data.Reason)
					} else {
						data, ok := event.(*LogData)
						assert.True(t, ok, "Expected *LogData")
						assert.Equal(t, expectedMsg, data.Msg)
					}
				}
			}(i)
		}
		wg.Wait()
	})
}
