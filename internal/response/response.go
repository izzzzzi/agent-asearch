package response

import "encoding/json"

type OK map[string]any

type Error struct {
	OK      bool   `json:"ok"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Hint    string `json:"hint,omitempty"`
}

func Marshal(v any) ([]byte, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return append(body, '\n'), nil
}

func MarshalError(code, message, hint string) ([]byte, error) {
	return Marshal(Error{OK: false, Code: code, Message: message, Hint: hint})
}
