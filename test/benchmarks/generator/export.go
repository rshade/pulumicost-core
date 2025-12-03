package generator

import (
	"encoding/json"
)

// ToJSON serializes a SyntheticPlan to JSON bytes.
func ToJSON(plan SyntheticPlan) ([]byte, error) {
	return json.Marshal(plan)
}

// ToJSONIndent serializes a SyntheticPlan to indented JSON bytes.
func ToJSONIndent(plan SyntheticPlan) ([]byte, error) {
	return json.MarshalIndent(plan, "", "  ")
}

// FromJSON deserializes JSON bytes to a SyntheticPlan.
func FromJSON(data []byte) (SyntheticPlan, error) {
	var plan SyntheticPlan
	err := json.Unmarshal(data, &plan)
	return plan, err
}
