package mapping

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"gopkg.in/yaml.v3"
)

// IDVal represents a service identifier which can be either a single integer
// or a list of integers (for split database entries). 0 represents unmapped.
type IDVal []int

// First returns the first non-zero ID, or 0 if none.
func (id IDVal) First() int {
	for _, v := range id {
		if v > 0 {
			return v
		}
	}
	return 0
}

// Has returns true if the specified target ID is present in IDVal.
func (id IDVal) Has(target int) bool {
	for _, v := range id {
		if v == target {
			return true
		}
	}
	return false
}

// AllNonZero returns all positive integers in this ID value.
func (id IDVal) AllNonZero() []int {
	var res []int
	for _, v := range id {
		if v > 0 {
			res = append(res, v)
		}
	}
	return res
}

// UnmarshalYAML handles both a single integer (e.g. 12345) and a list ([123, 456]).
func (id *IDVal) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		val, err := strconv.Atoi(node.Value)
		if err != nil {
			return fmt.Errorf("invalid integer ID '%s': %w", node.Value, err)
		}
		*id = IDVal{val}
		return nil
	}

	if node.Kind == yaml.SequenceNode {
		var list []int
		for _, child := range node.Content {
			val, err := strconv.Atoi(child.Value)
			if err != nil {
				return fmt.Errorf("invalid integer ID '%s' in list: %w", child.Value, err)
			}
			list = append(list, val)
		}
		*id = list
		return nil
	}

	return fmt.Errorf("unexpected YAML node kind for ID: %d", node.Kind)
}

// MarshalYAML serializes single IDs as integers and multiple IDs as sequences.
func (id IDVal) MarshalYAML() (interface{}, error) {
	if len(id) == 0 {
		return 0, nil
	}
	if len(id) == 1 {
		return id[0], nil
	}
	return []int(id), nil
}

// UnmarshalJSON handles both single integers and arrays of integers.
func (id *IDVal) UnmarshalJSON(data []byte) error {
	var single int
	if err := json.Unmarshal(data, &single); err == nil {
		*id = IDVal{single}
		return nil
	}

	var list []int
	if err := json.Unmarshal(data, &list); err == nil {
		*id = list
		return nil
	}

	return fmt.Errorf("failed to unmarshal ID value from JSON: %s", string(data))
}

// MarshalJSON outputs a single int if length <= 1, else array.
func (id IDVal) MarshalJSON() ([]byte, error) {
	if len(id) == 0 {
		return json.Marshal(0)
	}
	if len(id) == 1 {
		return json.Marshal(id[0])
	}
	return json.Marshal([]int(id))
}

// VolumeMap maps volume numbers (e.g. "1") to their ending chapter numbers (e.g. 7).
type VolumeMap map[string]int

// MarshalYAML serializes VolumeMap with keys sorted numerically (1, 2, 3 ... 10).
func (vm VolumeMap) MarshalYAML() (interface{}, error) {
	if len(vm) == 0 {
		return nil, nil
	}
	type pair struct {
		keyInt int
		keyStr string
		val    int
	}
	var pairs []pair
	for k, v := range vm {
		ki, _ := strconv.Atoi(k)
		pairs = append(pairs, pair{keyInt: ki, keyStr: k, val: v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].keyInt != pairs[j].keyInt {
			return pairs[i].keyInt < pairs[j].keyInt
		}
		return pairs[i].keyStr < pairs[j].keyStr
	})

	node := &yaml.Node{
		Kind: yaml.MappingNode,
	}
	for _, p := range pairs {
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: p.keyStr},
			&yaml.Node{Kind: yaml.ScalarNode, Value: strconv.Itoa(p.val)},
		)
	}
	return node, nil
}

// Entry represents a single mapping entry between databases.
type Entry struct {
	Title     string    `json:"title" yaml:"title"`
	Type      string    `json:"type" yaml:"type"` // manga, manhwa, manhua, novel, oneshot
	MALID     IDVal     `json:"mal_id" yaml:"mal_id"`
	AniListID IDVal     `json:"anilist_id" yaml:"anilist_id"`
	KitsuID   IDVal     `json:"kitsu_id" yaml:"kitsu_id"`
	Volumes   VolumeMap `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	Synonyms  []string  `json:"synonyms,omitempty" yaml:"synonyms,omitempty"`
}

// LeanEntry represents a clean lightweight entry without heavy synonyms.
type LeanEntry struct {
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	MALID     IDVal     `json:"mal_id"`
	AniListID IDVal     `json:"anilist_id"`
	KitsuID   IDVal     `json:"kitsu_id"`
	Volumes   VolumeMap `json:"volumes,omitempty"`
}

// ToLean converts an Entry to a LeanEntry.
func (e Entry) ToLean() LeanEntry {
	return LeanEntry{
		Title:     e.Title,
		Type:      e.Type,
		MALID:     e.MALID,
		AniListID: e.AniListID,
		KitsuID:   e.KitsuID,
		Volumes:   e.Volumes,
	}
}

// AllowedTypes defines valid content types for the dataset.
var AllowedTypes = map[string]bool{
	"manga":   true,
	"manhwa":  true,
	"manhua":  true,
	"novel":   true,
	"oneshot": true,
}
