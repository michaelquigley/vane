package model

import (
	"reflect"
	"testing"
)

func card(filename string, state State, created string) CardInput {
	return CardInput{Filename: filename, Title: filename, State: state, Created: created}
}

func laneFilenames(b Board, state State) []string {
	for _, l := range b.Lanes {
		if l.State == state {
			names := make([]string, len(l.Cards))
			for i, c := range l.Cards {
				names[i] = c.Filename
			}
			return names
		}
	}
	return nil
}

func rankedCount(b Board, state State) int {
	for _, l := range b.Lanes {
		if l.State == state {
			return l.RankedCount
		}
	}
	return -1
}

func TestComputeBoardDeterminism(t *testing.T) {
	tests := []struct {
		name             string
		cards            []CardInput
		order            map[State][]string
		wantDispositions map[State][]EntryDisposition
		wantLanes        map[State][]string
		wantRanked       map[State]int
	}{
		{
			name: "stale line never shadows a valid one",
			// x.md moved to building by hand; the stale researching entry is
			// discarded before first-occurrence-wins, so the building entry
			// ranks the card.
			cards: []CardInput{card("x.md", Building, "2026-07-01")},
			order: map[State][]string{
				Researching: {"x.md"},
				Building:    {"x.md"},
			},
			wantDispositions: map[State][]EntryDisposition{
				Researching: {EntryPrunable},
				Building:    {EntryActive},
			},
			wantLanes:  map[State][]string{Building: {"x.md"}, Researching: {}},
			wantRanked: map[State]int{Building: 1, Researching: 0},
		},
		{
			name: "duplicate filename first-wins after discard",
			cards: []CardInput{
				card("x.md", Researching, "2026-07-01"),
			},
			order: map[State][]string{
				Researching: {"gone.md", "x.md", "x.md"},
			},
			wantDispositions: map[State][]EntryDisposition{
				Researching: {EntryPrunable, EntryActive, EntryPrunable},
			},
			wantLanes:  map[State][]string{Researching: {"x.md"}},
			wantRanked: map[State]int{Researching: 1},
		},
		{
			name: "transient-unreadable retains rank",
			// x.md's state can't currently be read: its researching entry is
			// retained inert, not pruned, and the card degrades to inbox.
			cards: []CardInput{card("x.md", "", "2026-07-01")},
			order: map[State][]string{
				Researching: {"x.md"},
			},
			wantDispositions: map[State][]EntryDisposition{
				Researching: {EntryInert},
			},
			wantLanes:  map[State][]string{Researching: {}, Inbox: {"x.md"}},
			wantRanked: map[State]int{Researching: 0, Inbox: 0},
		},
		{
			name: "repaired state revives the retained rank",
			cards: []CardInput{card("x.md", Researching, "2026-07-01")},
			order: map[State][]string{
				Researching: {"x.md"},
			},
			wantDispositions: map[State][]EntryDisposition{
				Researching: {EntryActive},
			},
			wantLanes:  map[State][]string{Researching: {"x.md"}, Inbox: {}},
			wantRanked: map[State]int{Researching: 1},
		},
		{
			name: "malformed-created sorts after dated unranked",
			cards: []CardInput{
				card("bb.md", Inbox, "2026-07-02"),
				card("aa.md", Inbox, "2026-07-03"),
				card("00-undated.md", Inbox, ""),
			},
			order: map[State][]string{},
			wantLanes: map[State][]string{
				Inbox: {"bb.md", "aa.md", "00-undated.md"},
			},
			wantRanked: map[State]int{Inbox: 0},
		},
		{
			name: "unranked tail ties break by filename, undated sort by filename",
			cards: []CardInput{
				card("b.md", Horizon, "2026-07-01"),
				card("a.md", Horizon, "2026-07-01"),
				card("z-undated.md", Horizon, ""),
				card("m-undated.md", Horizon, ""),
			},
			order: map[State][]string{},
			wantLanes: map[State][]string{
				Horizon: {"a.md", "b.md", "m-undated.md", "z-undated.md"},
			},
			wantRanked: map[State]int{Horizon: 0},
		},
		{
			name: "inbox entry actively ranks an unreadable-state card; other lanes stay inert",
			cards: []CardInput{
				card("x.md", "", ""),
				card("y.md", Inbox, "2026-07-01"),
			},
			order: map[State][]string{
				Inbox:       {"x.md"},
				Researching: {"x.md"},
			},
			wantDispositions: map[State][]EntryDisposition{
				Inbox:       {EntryActive},
				Researching: {EntryInert},
			},
			wantLanes:  map[State][]string{Inbox: {"x.md", "y.md"}, Researching: {}},
			wantRanked: map[State]int{Inbox: 1, Researching: 0},
		},
		{
			name: "ranked prefix precedes unranked tail in entry order",
			cards: []CardInput{
				card("ranked-2.md", Researching, "2026-07-01"),
				card("ranked-1.md", Researching, "2026-07-02"),
				card("arrival.md", Researching, "2026-07-03"),
			},
			order: map[State][]string{
				Researching: {"ranked-1.md", "ranked-2.md"},
			},
			wantLanes: map[State][]string{
				Researching: {"ranked-1.md", "ranked-2.md", "arrival.md"},
			},
			wantRanked: map[State]int{Researching: 2},
		},
		{
			name: "lane mismatch discards only with positive evidence",
			// mismatched.md parses with a valid, different state: discard.
			// unreadable.md exists but can't be read: retain, inert.
			cards: []CardInput{
				card("mismatched.md", Building, "2026-07-01"),
				card("unreadable.md", "", "2026-07-01"),
			},
			order: map[State][]string{
				Horizon: {"mismatched.md", "unreadable.md", "gone.md"},
			},
			wantDispositions: map[State][]EntryDisposition{
				Horizon: {EntryPrunable, EntryInert, EntryPrunable},
			},
			wantLanes:  map[State][]string{Horizon: {}},
			wantRanked: map[State]int{Horizon: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := ComputeBoard(tt.cards, tt.order)
			for lane, want := range tt.wantDispositions {
				if got := b.Dispositions[lane]; !reflect.DeepEqual(got, want) {
					t.Errorf("dispositions[%s] = %v, want %v", lane, got, want)
				}
			}
			for lane, want := range tt.wantLanes {
				got := laneFilenames(b, lane)
				if len(got) == 0 && len(want) == 0 {
					continue
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("lane %s = %v, want %v", lane, got, want)
				}
			}
			for lane, want := range tt.wantRanked {
				if got := rankedCount(b, lane); got != want {
					t.Errorf("rankedCount[%s] = %d, want %d", lane, got, want)
				}
			}
		})
	}
}

func TestComputeBoardLanesInLifecycleOrder(t *testing.T) {
	b := ComputeBoard(nil, nil)
	if len(b.Lanes) != len(LaneOrder) {
		t.Fatalf("board has %d lanes, want %d", len(b.Lanes), len(LaneOrder))
	}
	for i, lane := range LaneOrder {
		if b.Lanes[i].State != lane {
			t.Errorf("lane[%d] = %q, want %q", i, b.Lanes[i].State, lane)
		}
	}
}
