package apmlogrus_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.elastic.co/apm"
	"go.elastic.co/apm/apmtest"
	"go.elastic.co/apm/module/apmlogrus"
)

func TestTraceContext(t *testing.T) {
	tx := apmtest.DiscardTracer.StartTransaction("name", "type")
	ctx := apm.ContextWithTransaction(context.Background(), tx)
	defer tx.End()
	traceContext := tx.TraceContext()

	span, ctx := apm.StartSpan(ctx, "name", "type")
	defer span.End()
	spanID := span.TraceContext().Span

	assert.Equal(t,
		fmt.Sprintf(
			`{"level":"debug","msg":"beep","span.id":"%x","time":"1970-01-01T00:00:00Z","trace.id":"%x","transaction.id":"%x"}`+"\n",
			spanID[:], traceContext.Trace[:], traceContext.Span[:],
		),
		logTraceContext(ctx),
	)
}

func TestTraceContextNoSpan(t *testing.T) {
	tx := apmtest.DiscardTracer.StartTransaction("name", "type")
	ctx := apm.ContextWithTransaction(context.Background(), tx)
	defer tx.End()
	traceContext := tx.TraceContext()

	assert.Equal(t,
		fmt.Sprintf(
			`{"level":"debug","msg":"beep","time":"1970-01-01T00:00:00Z","trace.id":"%x","transaction.id":"%x"}`+"\n",
			traceContext.Trace[:], traceContext.Span[:],
		),
		logTraceContext(ctx),
	)
}

func TestTraceContextEmpty(t *testing.T) {
	// apmlogrus.TraceContext will return nil if the context does not contain a transaction.
	output := logTraceContext(context.Background())
	assert.Equal(t, `{"level":"debug","msg":"beep","time":"1970-01-01T00:00:00Z"}`+"\n", output)
}

func logTraceContext(ctx context.Context) string {
	var buf bytes.Buffer
	logger := newLogger(&buf)
	logger.WithTime(time.Unix(0, 0).UTC()).WithFields(apmlogrus.TraceContext(ctx)).Debug("beep")
	return buf.String()
}
