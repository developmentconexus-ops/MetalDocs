package application

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"metaldocs/internal/modules/audit/domain"
)

// utf8BOM lets Excel auto-detect UTF-8 when opening the CSV.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

func render(format domain.ExportFormat, events []domain.Event) ([]byte, error) {
	switch format {
	case domain.ExportFormatCSV:
		return renderCSV(events)
	case domain.ExportFormatJSONL:
		return renderJSONL(events)
	default:
		return nil, ErrInvalidFormat
	}
}

func renderCSV(events []domain.Event) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write(utf8BOM)
	w := csv.NewWriter(&buf)
	header := []string{"occurred_at", "actor_id", "action", "resource_type", "resource_id", "trace_id", "payload_json"}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("csv header: %w", err)
	}
	for _, e := range events {
		row := []string{
			e.OccurredAt.UTC().Format(time.RFC3339Nano),
			e.ActorID,
			e.Action,
			e.ResourceType,
			e.ResourceID,
			e.TraceID,
			canonicalPayload(e.PayloadJSON),
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("csv row %s: %w", e.ID, err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("csv flush: %w", err)
	}
	return buf.Bytes(), nil
}

func renderJSONL(events []domain.Event) ([]byte, error) {
	var buf bytes.Buffer
	for _, e := range events {
		payload := map[string]any{}
		if raw := strings.TrimSpace(e.PayloadJSON); raw != "" {
			if err := json.Unmarshal([]byte(raw), &payload); err != nil {
				payload = map[string]any{"raw": raw}
			}
		}
		obj := map[string]any{
			"id":           e.ID,
			"occurredAt":   e.OccurredAt.UTC().Format(time.RFC3339Nano),
			"actorId":      e.ActorID,
			"action":       e.Action,
			"resourceType": e.ResourceType,
			"resourceId":   e.ResourceID,
			"traceId":      e.TraceID,
			"payload":      payload,
		}
		encoded, err := json.Marshal(obj)
		if err != nil {
			return nil, fmt.Errorf("jsonl marshal %s: %w", e.ID, err)
		}
		buf.Write(encoded)
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}

func canonicalPayload(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "{}"
	}
	return raw
}
