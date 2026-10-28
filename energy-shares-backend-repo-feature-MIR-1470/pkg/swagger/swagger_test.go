package swagger

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddNullableProperties(t *testing.T) {
	input := []byte(`{
  "swagger": "2.0",
  "example": 9223372036854775807,
  "definitions": {
    "ResponseItem": {
      "type": "object",
      "properties": {
        "status": {
          "type": "string"
        }
      }
    }
  }
}`)

	gotJSON, err := addNullableProperties(input, []NullableProperty{
		{Definition: "ResponseItem", Property: "status"},
	})
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(gotJSON, &got))
	definitions := got["definitions"].(map[string]interface{})
	definition := definitions["ResponseItem"].(map[string]interface{})
	properties := definition["properties"].(map[string]interface{})
	status := properties["status"].(map[string]interface{})
	assert.Equal(t, true, status["x-nullable"])
	assert.Contains(t, string(gotJSON), `"example": 9223372036854775807`)
}

func TestAddNullableProperties_NotFound(t *testing.T) {
	input := []byte(`{"definitions":{"ResponseItem":{"properties":{}}}}`)

	_, err := addNullableProperties(input, []NullableProperty{
		{Definition: "ResponseItem", Property: "status"},
	})
	require.EqualError(t, err, "swagger property not found: ResponseItem.status")
}
