package apmlogrus_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"go.elastic.co/apm"
	"go.elastic.co/apm/model"
	"go.elastic.co/apm/module/apmlogrus"
	"go.elastic.co/apm/transport/transporttest"
)

func TestHook(t *testing.T) {
	tracer, transport := transporttest.NewRecorderTracer()
	defer tracer.Close()

	var buf bytes.Buffer
	logger := newLogger(&buf)
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
	assert.Zero(t, err0.ParentID)
	assert.Zero(t, err0.TraceID)
	assert.Zero(t, err0.TransactionID)
}

func TestHookTransactionTraceContext(t *testing.T) {
	tracer, transport := transporttest.NewRecorderTracer()
	defer tracer.Close()

	logger := newLogger(ioutil.Discard)
	logger.AddHook(&apmlogrus.Hook{Tracer: tracer})

	tx := tracer.StartTransaction("name", "type")
	ctx := apm.ContextWithTransaction(context.Background(), tx)
	span, ctx := apm.StartSpan(ctx, "name", "type")
	logger.WithFields(apmlogrus.TraceContext(ctx)).Errorf("¡hola, %s!", "mundo")
	span.End()
	tx.End()

	tracer.Flush(nil)
	payloads := transport.Payloads()
	assert.Len(t, payloads.Transactions, 1)
	assert.Len(t, payloads.Spans, 1)
	assert.Len(t, payloads.Errors, 1)

	err0 := payloads.Errors[0]
	assert.Equal(t, payloads.Spans[0].ID, err0.ParentID)
	assert.Equal(t, payloads.Transactions[0].TraceID, err0.TraceID)
	assert.Equal(t, payloads.Transactions[0].ID, err0.TransactionID)
}

func TestHookTracerClosed(t *testing.T) {
	tracer, _ := transporttest.NewRecorderTracer()
	tracer.Close() // close it straight away, hook should return immediately

	logger := newLogger(ioutil.Discard)
	logger.AddHook(&apmlogrus.Hook{Tracer: tracer})
	logger.Error("boom")
}

func newLogger(w io.Writer) *logrus.Logger {
	return &logrus.Logger{
		Out:       w,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
}
