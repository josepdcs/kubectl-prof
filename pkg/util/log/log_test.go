package log

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	assert.NotNil(t, Default())
	assert.False(t, Default().printLogs)
}

func TestLogger_SetPrintLogs(t *testing.T) {
	l := New()
	require.False(t, l.printLogs)
	l.SetPrintLogs(true)
	require.True(t, l.printLogs)
}

func TestSetPrintLogs(t *testing.T) {
	require.False(t, PrintLogs())
	SetPrintLogs(true)
	require.True(t, PrintLogs())
}

func TestLogger_EventLn(t *testing.T) {
	type fields struct {
		printLogs bool
	}

	type args struct {
		eventType api.EventType
		data      interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Should only print message",
			fields: fields{},
			args: args{
				eventType: api.Progress,
				data:      &api.ProgressData{Time: time.Now(), Stage: api.Started},
			},
			wantErr: false,
		},
		{
			name:   "Should fail when not marshal",
			fields: fields{},
			args: args{
				eventType: api.Progress,
				data:      make(chan int),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			l := &Logger{
				printLogs: tt.fields.printLogs,
			}

			// When
			err := l.EventLn(tt.args.eventType, tt.args.data)

			// Then
			if err != nil && !tt.wantErr {
				t.Errorf("EventLn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventLn(t *testing.T) {
	assert.NoError(t, EventLn(api.Progress, &api.ProgressData{Time: time.Now(), Stage: api.Started}))
}

func TestLogger_ErrorLn(t *testing.T) {
	l := &Logger{}
	l.ErrorLn(fmt.Errorf("my error"))
}

func TestErrorLn(t *testing.T) {
	ErrorLn(fmt.Errorf("my error"))
}

func TestLogger_PrintLogLn(t *testing.T) {
	type args struct {
		level api.LogLevel
		msg   string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "should do nothing when empty message",
		},
		{
			name: "should print message",
			args: args{
				level: api.InfoLevel,
				msg:   "this is the message",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Logger{}
			l.PrintLogLn(tt.args.level, tt.args.msg)
		})
	}
}

func TestPrintLogLn(t *testing.T) {
	type args struct {
		level api.LogLevel
		msg   string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "should do nothing when empty message",
		},
		{
			name: "should print message",
			args: args{
				level: api.InfoLevel,
				msg:   "this is the message",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PrintLogLn(tt.args.level, tt.args.msg)
		})
	}
}

func TestInfoLogLn(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "should do nothing when empty message",
		},
		{
			name: "should print message",
			args: args{
				msg: "this is the message",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InfoLogLn(tt.args.msg)
		})
	}
}

func TestDebugLogLn(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "should do nothing when empty message",
		},
		{
			name: "should print message",
			args: args{
				msg: "this is the message",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DebugLogLn(tt.args.msg)
		})
	}
}

func TestWarningLogLn(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "should do nothing when empty message",
		},
		{
			name: "should print message",
			args: args{
				msg: "this is the message",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WarningLogLn(tt.args.msg)
		})
	}
}

func TestErrorLogLn(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "should do nothing when empty message",
		},
		{
			name: "should print message",
			args: args{
				msg: "this is the message",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ErrorLogLn(tt.args.msg)
		})
	}
}
