package integration

import (
	"testing"
	"time"
)

// poll retries fn every 500ms until it returns true, or fails the test once
// timeout elapses. Go equivalent of Poller.WaitAsync in the C# reference:
// cross-module propagation goes through outbox -> event bus -> inbox, so the
// effect of a command isn't visible in another module until its outbox
// worker's next tick.
func poll(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		if fn() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("poll: condition not met within %s", timeout)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
