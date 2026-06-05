package notification

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/chainguard-sandbox/go-linear/v2/internal/testutil"
)

func TestNewNotificationCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewNotificationCommand(factory)

	if cmd.Use != "notification" {
		t.Errorf("Use = %q, want %q", cmd.Use, "notification")
	}
	if len(cmd.Commands()) == 0 {
		t.Error("Expected subcommands")
	}
}

func TestNewArchiveCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewArchiveCommand(factory)

	if !strings.HasPrefix(cmd.Use, "archive") {
		t.Errorf("Use = %q, want prefix archive", cmd.Use)
	}
}

func TestRunArchive(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewArchiveCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"notif-123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestNewUpdateCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewUpdateCommand(factory)

	if !strings.HasPrefix(cmd.Use, "update") {
		t.Errorf("Use = %q, want prefix update", cmd.Use)
	}
}

func TestRunUpdate(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("mark read", func(t *testing.T) {
		cmd := NewUpdateCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"notif-123", "--read"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

// TestRunUpdate_SnoozeUntilIsFuture is a regression for #74: snooze-until must
// be parsed in the future direction, so a relative input like "3d" produces a
// future timestamp rather than one 3 days in the past.
func TestRunUpdate_SnoozeUntilIsFuture(t *testing.T) {
	server, lastVars := testutil.MockServerCapture(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewUpdateCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"notif-123", "--snooze-until=3d"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	vars := lastVars()
	if vars == nil {
		t.Fatal("no variables captured")
	}
	input, ok := vars["input"].(map[string]any)
	if !ok {
		t.Fatalf("variables[input] = %T, want map", vars["input"])
	}
	raw, ok := input["snoozedUntilAt"].(string)
	if !ok {
		t.Fatalf("input[snoozedUntilAt] = %T, want string", input["snoozedUntilAt"])
	}
	got, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("snoozedUntilAt %q not RFC3339: %v", raw, err)
	}
	// Assert against a lower bound two days out (the input is "3d") rather than
	// time.Now(): this documents the expected magnitude and stays robust if the
	// fixture's duration ever shrinks toward minutes.
	if !got.After(time.Now().Add(2 * 24 * time.Hour)) {
		t.Errorf("snooze-until=3d sent %v, want a future time at least 2 days out (regression #74)", got)
	}
}

// TestRunUpdate_SnoozeUntilRejectsPast is the negative half of #74: because
// snooze-until now parses in the future direction, an explicitly past date must
// be rejected rather than silently accepted.
func TestRunUpdate_SnoozeUntilRejectsPast(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewUpdateCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"notif-123", "--snooze-until=2020-01-01"})
	if err := cmd.Execute(); err == nil {
		t.Error("snooze-until=2020-01-01 (a past date) should be rejected, got nil error")
	}
}

func TestNewSubscribeCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewSubscribeCommand(factory)

	if cmd.Use != "subscribe" {
		t.Errorf("Use = %q, want %q", cmd.Use, "subscribe")
	}
}

func TestRunSubscribe(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	t.Run("subscribe to project", func(t *testing.T) {
		cmd := NewSubscribeCommand(factory)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		// Use UUID format
		cmd.SetArgs([]string{"--project=00000000-0000-0000-0000-000000000001"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
	})
}

func TestNewUnsubscribeCommand(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)
	cmd := NewUnsubscribeCommand(factory)

	if !strings.HasPrefix(cmd.Use, "unsubscribe") {
		t.Errorf("Use = %q, want prefix unsubscribe", cmd.Use)
	}
}

func TestRunUnsubscribe(t *testing.T) {
	server := testutil.MockServer(t, defaultHandlers())
	defer server.Close()
	factory := testutil.TestFactory(t, server.URL)

	cmd := NewUnsubscribeCommand(factory)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"sub-123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}
