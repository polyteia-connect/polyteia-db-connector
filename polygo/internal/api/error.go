package api

import (
	"encoding/json"
	"fmt"
)

type Error struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Details []map[string]any `json:"details"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("polygo api error: code -> %d, message -> %s, details -> %v", e.Code, e.Message, e.Details)
}

func getError(result map[string]any) error {
	if result == nil {
		return nil
	}

	if val, ok := result["error"]; ok {
		// unmarshal error
		raw, err := json.Marshal(val)
		if err != nil {
			return err
		}

		var apiErr Error
		if err := json.Unmarshal(raw, &apiErr); err != nil {
			return err
		}

		return &apiErr
	}

	return nil
}
