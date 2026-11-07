package mapping

import (
	"testing"
)

func TestValidateVolumeMap(t *testing.T) {
	tests := []struct {
		name     string
		vols     VolumeMap
		expected bool
	}{
		{
			name:     "Valid sequential volumes",
			vols:     VolumeMap{"1": 7, "2": 16, "3": 25},
			expected: true,
		},
		{
			name:     "Valid single multi-chapter volume",
			vols:     VolumeMap{"1": 6},
			expected: true,
		},
		{
			name:     "Empty volume map",
			vols:     VolumeMap{},
			expected: false,
		},
		{
			name:     "Missing volume 1",
			vols:     VolumeMap{"2": 15, "3": 25},
			expected: false,
		},
		{
			name:     "Volume 1 ends at chapter 0",
			vols:     VolumeMap{"1": 0, "2": 10},
			expected: false,
		},
		{
			name:     "Volume 1 ends at chapter 1",
			vols:     VolumeMap{"1": 1, "2": 10},
			expected: false,
		},
		{
			name:     "Gap in sequence (missing volume 2)",
			vols:     VolumeMap{"1": 7, "3": 25},
			expected: false,
		},
		{
			name:     "Non-ascending chapter boundaries",
			vols:     VolumeMap{"1": 10, "2": 8},
			expected: false,
		},
		{
			name:     "Equal chapter boundaries",
			vols:     VolumeMap{"1": 10, "2": 10},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateVolumeMap(tc.vols)
			if got != tc.expected {
				t.Errorf("ValidateVolumeMap(%+v) = %v; want %v", tc.vols, got, tc.expected)
			}
		})
	}
}

func TestPruneOngoingVolumes(t *testing.T) {
	tests := []struct {
		name      string
		vols      VolumeMap
		isOngoing bool
		expected  VolumeMap
	}{
		{
			name:      "Completed series keeps all volumes",
			vols:      VolumeMap{"1": 7, "2": 15, "3": 24},
			isOngoing: false,
			expected:  VolumeMap{"1": 7, "2": 15, "3": 24},
		},
		{
			name:      "Ongoing series prunes unclosed final volume",
			vols:      VolumeMap{"1": 7, "2": 15, "3": 24},
			isOngoing: true,
			expected:  VolumeMap{"1": 7, "2": 15},
		},
		{
			name:      "Ongoing series with only 1 volume prunes to nil (cannot verify closure)",
			vols:      VolumeMap{"1": 7},
			isOngoing: true,
			expected:  nil,
		},
		{
			name:      "Empty map returns empty",
			vols:      VolumeMap{},
			isOngoing: true,
			expected:  VolumeMap{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := PruneOngoingVolumes(tc.vols, tc.isOngoing)
			if len(got) != len(tc.expected) {
				t.Fatalf("PruneOngoingVolumes(%+v, %v) len = %d, want %d", tc.vols, tc.isOngoing, len(got), len(tc.expected))
			}
			for k, v := range tc.expected {
				if got[k] != v {
					t.Errorf("PruneOngoingVolumes(%+v, %v)[%s] = %d, want %d", tc.vols, tc.isOngoing, k, got[k], v)
				}
			}
		})
	}
}
