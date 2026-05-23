package coursemembersdb

import (
	"encoding/json"
	"fmt"
)

type Member struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Age   int64  `json:"age"`
}

type Members []Member

func (m *Members) Scan(src any) error {
	if src == nil {
		*m = nil
		return nil
	}

	var b []byte

	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("members: unexpected type %T", src)
	}

	return json.Unmarshal(b, m)
}
