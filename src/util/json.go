package util

import "encoding/json"

func ToJson(v interface{}) []byte {
	result, _ := json.Marshal(v)
	return result
}
