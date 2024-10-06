package proxy

import (
	"testing"
)

// Ensures that the metrics prefix is correctly generated from the given endpoint
func TestFromResourceToMetricsPrefix(t *testing.T) {
	// First, test basic expansion of a simple path
	expected := "home_"
	actual := FromResourceToMetricsPrefix("/home")
	if actual != expected {
		t.Fatalf(`Expected prefix "%s" but got "%s"`, expected, actual)
	}

	// Next, test a nested path
	expected = "home_about_"
	actual = FromResourceToMetricsPrefix("/home/about")
	if actual != expected {
		t.Fatalf(`Expected prefix "%s" but got "%s"`, expected, actual)
	}

	// Root edge case
	expected = "ROOT_"
	actual = FromResourceToMetricsPrefix("/")
	if actual != expected {
		t.Fatalf(`Expected prefix "%s" but got "%s"`, expected, actual)
	}
}
