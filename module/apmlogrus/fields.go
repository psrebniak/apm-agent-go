package apmlogrus

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.elastic.co/apm"
)

const (
	// FieldTraceID is the field name for the trace ID.
	FieldTraceID = "trace.id"

	// FieldTransactionID is the field name for the transaction ID.
	FieldTransactionID = "transaction.id"

	// FieldSpanID is the field name for the span ID.
	FieldSpanID = "span.id"
)

// TraceContext returns a logrus.Fields containing the trace
// context of the transaction and span contained in ctx, if any.
func TraceContext(ctx context.Context) logrus.Fields {
	tx := apm.TransactionFromContext(ctx)
	if tx == nil {
		return nil
	}
	traceContext := tx.TraceContext()
	fields := logrus.Fields{
		FieldTraceID:       traceContext.Trace,
		FieldTransactionID: traceContext.Span,
	}
	if span := apm.SpanFromContext(ctx); span != nil {
		fields[FieldSpanID] = span.TraceContext().Span
	}
	return fields
}
