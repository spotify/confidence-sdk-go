package confidence

import (
	"testing"
	"time"
)

func TestAPIConfig_WithResolveTimeout(t *testing.T) {
	config := NewAPIConfig("test-key")

	// Verify default timeout
	if config.ResolveTimeout != 10000*time.Millisecond {
		t.Errorf("Expected default ResolveTimeout to be 10000ms, got %v", config.ResolveTimeout)
	}

	// Test setting custom timeout
	customTimeout := 5 * time.Second
	config.WithResolveTimeout(customTimeout)

	if config.ResolveTimeout != customTimeout {
		t.Errorf("Expected ResolveTimeout to be %v, got %v", customTimeout, config.ResolveTimeout)
	}
}

func TestAPIConfig_WithResolveTimeout_Chaining(t *testing.T) {
	customTimeout := 3 * time.Second
	config := NewAPIConfig("test-key").WithResolveTimeout(customTimeout)

	if config.ResolveTimeout != customTimeout {
		t.Errorf("Expected ResolveTimeout to be %v, got %v", customTimeout, config.ResolveTimeout)
	}
}

func TestAPIConfig_WithResolveTimeout_MultipleUpdates(t *testing.T) {
	config := NewAPIConfig("test-key")

	// Update timeout multiple times
	firstTimeout := 2 * time.Second
	secondTimeout := 7 * time.Second

	config.WithResolveTimeout(firstTimeout)
	if config.ResolveTimeout != firstTimeout {
		t.Errorf("Expected ResolveTimeout to be %v after first update, got %v", firstTimeout, config.ResolveTimeout)
	}

	config.WithResolveTimeout(secondTimeout)
	if config.ResolveTimeout != secondTimeout {
		t.Errorf("Expected ResolveTimeout to be %v after second update, got %v", secondTimeout, config.ResolveTimeout)
	}
}
