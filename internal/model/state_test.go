package model

import "testing"

func TestParseState(t *testing.T) {
	for _, lane := range LaneOrder {
		got, ok := ParseState(string(lane))
		if !ok || got != lane {
			t.Errorf("ParseState(%q) = %q, %v; want %q, true", lane, got, ok, lane)
		}
	}
	for _, s := range []string{"", "Inbox", "shipped", "in-progress"} {
		if _, ok := ParseState(s); ok {
			t.Errorf("ParseState(%q) accepted; want invalid", s)
		}
	}
}

func TestLaneOrder(t *testing.T) {
	want := []State{Inbox, Horizon, Researching, Building, Evaluating, Done, Dropped}
	if len(LaneOrder) != len(want) {
		t.Fatalf("LaneOrder has %d lanes, want %d", len(LaneOrder), len(want))
	}
	for i, lane := range want {
		if LaneOrder[i] != lane {
			t.Errorf("LaneOrder[%d] = %q, want %q", i, LaneOrder[i], lane)
		}
	}
}
