package supabase

import "encoding/json"

func UnmarshalJSON(b []byte, v any) error {
	return json.Unmarshal(b, v)
}
