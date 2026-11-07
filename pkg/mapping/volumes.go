package mapping

import (
	"strconv"
)

// ValidateVolumeMap ensures volume data is coherent, sequential, starts at Vol 1, and strictly increases.
func ValidateVolumeMap(vols VolumeMap) bool {
	if len(vols) == 0 {
		return false
	}

	// 1. Volume 1 MUST exist and cannot end at chapter 0 or 1 (unless it is a known oneshot volume)
	v1, ok := vols["1"]
	if !ok || v1 < 2 {
		return false
	}

	// 2. Must be strictly sequential from volume 1 to N without gaps
	prevCh := v1
	for i := 2; i <= len(vols); i++ {
		ch, exists := vols[strconv.Itoa(i)]
		if !exists {
			return false // Gap in volume sequence (e.g. Vol 1, 2, 4 missing 3)
		}
		if ch <= prevCh {
			return false // Chapters must strictly increase
		}
		prevCh = ch
	}

	return true
}

// PruneOngoingVolumes strips the highest volume number from an ongoing series,
// because without a subsequent (N+1) volume or completion confirmation,
// we cannot guarantee whether the highest volume has finished or is still serializing.
func PruneOngoingVolumes(vols VolumeMap, isOngoing bool) VolumeMap {
	if !isOngoing || len(vols) == 0 {
		return vols
	}

	maxVol := 0
	for k := range vols {
		v, err := strconv.Atoi(k)
		if err == nil && v > maxVol {
			maxVol = v
		}
	}

	if maxVol <= 0 {
		return vols
	}

	pruned := make(VolumeMap)
	for k, v := range vols {
		vNum, err := strconv.Atoi(k)
		if err == nil && vNum != maxVol {
			pruned[k] = v
		}
	}

	if !ValidateVolumeMap(pruned) {
		return nil
	}

	return pruned
}
