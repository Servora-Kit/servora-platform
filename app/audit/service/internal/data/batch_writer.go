package data

import (
	"context"
	"sync"
	"time"

	auditv1 "github.com/Servora-Kit/servora/api/gen/go/servora/audit/v1"
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	"github.com/Servora-Kit/servora/infra/broker"
	"github.com/Servora-Kit/servora/obs/logging"
	"google.golang.org/protobuf/encoding/protojson"
)

const flushOnStopTimeout = 10 * time.Second

// pendingEvent bundles a deserialized event with its Kafka event handle for Ack/Nack.
type pendingEvent struct {
	event    *auditv1.AuditEvent
	kafkaEvt broker.Event
}

// BatchWriter buffers AuditEvent records and flushes them to ClickHouse in batches.
type BatchWriter struct {
	data      *Data
	log       *logger.Helper
	batchSize int
	interval  time.Duration

	mu     sync.Mutex
	buffer []pendingEvent

	flushCh  chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

// NewBatchWriter creates a new BatchWriter using config from conf.App.Audit.
func NewBatchWriter(d *Data, appCfg *conf.App, l logger.Logger) *BatchWriter {
	batchSize := 100
	interval := time.Second

	if appCfg != nil && appCfg.Audit != nil {
		if appCfg.Audit.ConsumerBatchSize > 0 {
			batchSize = int(appCfg.Audit.ConsumerBatchSize)
		}
		if appCfg.Audit.ConsumerFlushInterval != nil {
			d := appCfg.Audit.ConsumerFlushInterval.AsDuration()
			if d > 0 {
				interval = d
			}
		}
	}

	return &BatchWriter{
		data:      d,
		log:       logger.For(l, "batch_writer/data/audit"),
		batchSize: batchSize,
		interval:  interval,
		buffer:    make([]pendingEvent, 0, batchSize),
		flushCh:   make(chan struct{}, 1),
		done:      make(chan struct{}),
	}
}

// Start begins the background timer flush loop.
func (w *BatchWriter) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.flush(ctx)
			case <-w.flushCh:
				w.flush(ctx)
			case <-w.done:
				// Use an independent context for the final flush so that a
				// cancelled Start ctx doesn't silently drop the last batch.
				stopCtx, cancel := context.WithTimeout(context.Background(), flushOnStopTimeout)
				defer cancel()
				w.flush(stopCtx)
				return
			}
		}
	}()
}

// Stop signals the flush loop to stop and flushes remaining events.
// Safe to call multiple times.
func (w *BatchWriter) Stop() {
	w.stopOnce.Do(func() { close(w.done) })
}

// Add appends an event to the buffer. If the buffer reaches batchSize, triggers an immediate flush.
func (w *BatchWriter) Add(evt *auditv1.AuditEvent, kafkaEvt broker.Event) {
	w.mu.Lock()
	w.buffer = append(w.buffer, pendingEvent{event: evt, kafkaEvt: kafkaEvt})
	shouldFlush := len(w.buffer) >= w.batchSize
	w.mu.Unlock()

	if shouldFlush {
		select {
		case w.flushCh <- struct{}{}:
		default:
		}
	}
}

// flush writes all buffered events to ClickHouse and Ack/Nack their Kafka handles.
func (w *BatchWriter) flush(ctx context.Context) {
	w.mu.Lock()
	if len(w.buffer) == 0 {
		w.mu.Unlock()
		return
	}
	batch := w.buffer
	w.buffer = make([]pendingEvent, 0, w.batchSize)
	w.mu.Unlock()

	if w.data.ClickHouse() == nil {
		// ClickHouse not configured; ack events so they don't block
		for _, p := range batch {
			if p.kafkaEvt != nil {
				_ = p.kafkaEvt.Ack()
			}
		}
		return
	}

	chBatch, err := w.data.ClickHouse().PrepareBatch(ctx, "INSERT INTO audit_events")
	if err != nil {
		w.log.Warnf("failed to prepare batch: %v", err)
		w.nackAll(batch)
		return
	}

	for _, p := range batch {
		detail := detailJSON(p.event)
		e := p.event

		occurredAt := e.OccurredAt.AsTime()

		var actorID, actorType, actorDisplayName string
		if e.Actor != nil {
			actorID = e.Actor.Id
			actorType = e.Actor.Type
			actorDisplayName = e.Actor.DisplayName
		}

		var targetType, targetID, targetName string
		if e.Target != nil {
			targetType = e.Target.Type
			targetID = e.Target.Id
			targetName = e.Target.Name
		}

		var success bool
		var errorCode, errorMessage string
		if e.Result != nil {
			success = e.Result.Success
			errorCode = e.Result.ErrorCode
			errorMessage = e.Result.ErrorMessage
		}

		if err := chBatch.Append(
			e.EventId,
			e.EventType.String(),
			e.EventVersion,
			occurredAt,
			e.Service,
			e.Operation,
			actorID,
			actorType,
			actorDisplayName,
			targetType,
			targetID,
			targetName,
			success,
			errorCode,
			errorMessage,
			e.TraceId,
			e.RequestId,
			detail,
		); err != nil {
			w.log.Warnf("append failed for event %s, aborting batch: %v", e.EventId, err)
			_ = chBatch.Abort()
			w.nackAll(batch)
			return
		}
	}

	if err := chBatch.Send(); err != nil {
		w.log.Warnf("failed to send batch: %v", err)
		w.nackAll(batch)
		return
	}

	w.log.Infof("flushed %d events to ClickHouse", len(batch))
	w.ackAll(batch)
}

func (w *BatchWriter) ackAll(batch []pendingEvent) {
	for _, p := range batch {
		if p.kafkaEvt != nil {
			if err := p.kafkaEvt.Ack(); err != nil {
				w.log.Warnf("failed to ack event: %v", err)
			}
		}
	}
}

func (w *BatchWriter) nackAll(batch []pendingEvent) {
	for _, p := range batch {
		if p.kafkaEvt != nil {
			if err := p.kafkaEvt.Nack(); err != nil {
				w.log.Warnf("failed to nack event: %v", err)
			}
		}
	}
}

// detailJSON serializes the oneof detail field to a JSON string using protojson.
func detailJSON(e *auditv1.AuditEvent) string {
	if e == nil {
		return "{}"
	}
	switch d := e.Detail.(type) {
	case *auditv1.AuditEvent_AuthnDetail:
		b, err := protojson.Marshal(d.AuthnDetail)
		if err != nil {
			return "{}"
		}
		return string(b)
	case *auditv1.AuditEvent_AuthzDetail:
		b, err := protojson.Marshal(d.AuthzDetail)
		if err != nil {
			return "{}"
		}
		return string(b)
	case *auditv1.AuditEvent_TupleMutationDetail:
		b, err := protojson.Marshal(d.TupleMutationDetail)
		if err != nil {
			return "{}"
		}
		return string(b)
	case *auditv1.AuditEvent_ResourceMutationDetail:
		b, err := protojson.Marshal(d.ResourceMutationDetail)
		if err != nil {
			return "{}"
		}
		return string(b)
	default:
		return "{}"
	}
}
