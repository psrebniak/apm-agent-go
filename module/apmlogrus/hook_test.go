package apmlogrus_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"go.elastic.co/apm/model"
	"go.elastic.co/apm/module/apmlogrus"
	"go.elastic.co/apm/transport/transporttest"
)

func TestHook(t *testing.T) {
	tracer, transport := transporttest.NewRecorderTracer()
	defer tracer.Close()

	var buf bytes.Buffer
	logger := &logrus.Logger{
		Out:       &buf,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
	logger.AddHook(&apmlogrus.Hook{Tracer: tracer})
	logger.WithTime(time.Unix(0, 0).UTC()).Errorf("¡hola, %s!", "mundo")

	assert.Equal(t, `{"level":"error","msg":"¡hola, mundo!","time":"1970-01-01T00:00:00Z"}`+"\n", buf.String())

	tracer.Flush(nil)
	payloads := transport.Payloads()
	assert.Len(t, payloads.Errors, 1)

	err0 := payloads.Errors[0]
	assert.Equal(t, "¡hola, mundo!", err0.Log.Message)
	assert.Equal(t, "error", err0.Log.Level)
	assert.Equal(t, "", err0.Log.LoggerName)
	assert.Equal(t, "", err0.Log.ParamMessage)
	assert.Equal(t, "TestHook", err0.Culprit)
	assert.NotEmpty(t, err0.Log.Stacktrace)
	assert.Equal(t, model.Time(time.Unix(0, 0).UTC()), err0.Timestamp)
	assert.Zero(t, err0.ID) // XXX check if this is a problem
	assert.Zero(t, err0.ParentID)
	assert.Zero(t, err0.TraceID)
	assert.Zero(t, err0.TransactionID)
}
