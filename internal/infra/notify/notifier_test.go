package notify

import (
	"testing"
	"wood-hot-monitor/pkg/types"
)

func TestMeetsMinImportance(t *testing.T) {
	cases := []struct {
		importance types.Importance
		min        types.Importance
		want       bool
	}{
		{types.ImportanceHigh, types.ImportanceHigh, true},
		{types.ImportanceUrgent, types.ImportanceHigh, true},
		{types.ImportanceMedium, types.ImportanceHigh, false},
		{types.ImportanceLow, types.ImportanceLow, true},
		{types.ImportanceHigh, types.ImportanceHigh, true},    // empty min defaults to high
		{types.ImportanceMedium, types.ImportanceHigh, false}, // empty min defaults to high
		{types.ImportanceUrgent, types.ImportanceUrgent, true},
		{types.ImportanceHigh, types.ImportanceUrgent, false},
		{types.ImportanceLow, types.ImportanceHigh, false},
	}
	for _, tc := range cases {
		got := MeetsMinImportance(tc.importance, tc.min)
		if got != tc.want {
			t.Errorf("MeetsMinImportance(%q, %q) = %v, want %v", tc.importance, tc.min, got, tc.want)
		}
	}
}
