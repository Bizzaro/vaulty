package events

import (
	"fmt"
	"time"
)

func TimedEventLog(invocation time.Time, description string, events EventStore) {
	elapsed := time.Since(invocation)
	took := fmt.Sprintf("\nTook: %.2fs", elapsed.Seconds())
	event := (description)
	events.NewEvent(event, took)
}
