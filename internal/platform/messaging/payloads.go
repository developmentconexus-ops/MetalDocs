package messaging

import (
	"encoding/json"
	"fmt"
)

func DecodePayload(eventType EventType, payloadJSON []byte) (Payload, error) {
	switch eventType {
	case EventTypePDFConvert:
		var payload PDFConvertPayload
		if err := json.Unmarshal(payloadJSON, &payload); err != nil {
			return nil, fmt.Errorf("decode pdf convert payload: %w", err)
		}
		return payload, nil
	default:
		var payload UnknownPayload
		if err := json.Unmarshal(payloadJSON, &payload); err != nil {
			return nil, fmt.Errorf("decode unknown payload: %w", err)
		}
		return payload, nil
	}
}

func PDFConvertPayloadFrom(event Event) (PDFConvertPayload, error) {
	payload, ok := event.Payload.(PDFConvertPayload)
	if !ok {
		return PDFConvertPayload{}, fmt.Errorf("event %s payload has type %T, want messaging.PDFConvertPayload", event.EventID, event.Payload)
	}
	return payload, nil
}
