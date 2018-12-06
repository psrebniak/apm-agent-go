package apmlogrus_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmlogrus"
	"go.elastic.co/apm/transport/transporttest"
)

func TestTraceContext(t *testing.T) {
	tracer, _ := transporttest.NewRecorderTracer()
	defer tracer.Close()

	var buf bytes.Buffer
	logger := newLogger(&buf)
	logger.AddHook(&apmlogrus.Hook{Tracer: tracer})

	tx := tracer.StartTransaction("name", "type")
	ctx := apm.ContextWithTransaction(context.Background(), tx)
	defer tx.End()

	span, ctx := apm.StartSpan(ctx, "name", "type")
	defer span.End()

	traceContext := tx.TraceContext()
	spanID := span.TraceContext().Span

	logger.WithTime(time.Unix(0, 0).UTC()).WithFields(apmlogrus.TraceContext(ctx)).Debug("beep")
	assert.Equal(t,
		fmt.Sprintf(
			`{"level":"debug","msg":"beep","span.id":"%x","time":"1970-01-01T00:00:00Z","trace.id":"%x","transaction.id":"%x"}`+"\n",
			spanID[:], traceContext.Trace[:], traceContext.Span[:],
		),
		buf.String(),
	)
}
