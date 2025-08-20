package trigger

import (
	"encoding/json"
	"fmt"
)

type Source string

const (
	HttpSource      Source = "http"
	TimerSource     Source = "timer"
	LocalSource     Source = "local"
	UnknownSource   Source = "unknown"
	NotParsedSource Source = "not parsed"
)

// LocalEvent represents a locally generated event with a single body field in JSON format.
type LocalEvent struct {
	Body string `json:"body"`
}

// TimerEvent represents the structure of an event from a Yandex Cloud timer trigger.
type TimerEvent struct {
	Details struct {
		TriggerID string `json:"trigger_id"`
	} `json:"details"`
}

// HTTPEvent represents the structure of an event from a Yandex Cloud HTTP trigger.
type HTTPEvent struct {
	HTTPMethod string            `json:"httpMethod"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Url        string            `json:"url"`
}

// DetectType determines the type of trigger that invoked the function (timer or HTTP).
func DetectType(event interface{}) string {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return string(NotParsedSource)
	}

	// Local Event
	var localEvent LocalEvent
	err = json.Unmarshal(eventBytes, &localEvent)
	if err == nil && localEvent.Body != "" && localEvent.Body == string(LocalSource) {
		return string(LocalSource)
	}

	// TimerEvent
	var timerEvent TimerEvent
	err = json.Unmarshal(eventBytes, &timerEvent)
	if err == nil && timerEvent.Details.TriggerID != "" {
		return fmt.Sprintf("%s: %s", TimerSource, timerEvent.Details.TriggerID)
	}

	// HTTPEvent
	var httpEvent HTTPEvent
	err = json.Unmarshal(eventBytes, &httpEvent)
	if err == nil && httpEvent.HTTPMethod != "" {
		return string(HttpSource)
	}

	// Default
	return string(UnknownSource)
}
