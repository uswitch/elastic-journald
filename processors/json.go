package processors

import (
	"encoding/json"
	"fmt"
)

type JsonFieldParser struct {
	SourceField string
	TargetField string
}

func NewJsonFieldParser(sourceField string, targetField string) *JsonFieldParser {
	return &JsonFieldParser{
		SourceField: sourceField,
		TargetField: targetField,
	}
}

func (p *JsonFieldParser) Process(entry LogEntry) {
	field, ok := entry[p.SourceField]
	if !ok {
		return
	}
	message := field.(string)
	if message[0] != '{' {
		return
	}

	jsonObj := make(map[string]interface{})
	err := json.Unmarshal([]byte(message), &jsonObj)
	if err != nil {
		return
	}

	jsonObjWithStringKeys := make(map[string]string)
	for key, val := range jsonObj {
		jsonObjWithStringKeys[key] = fmt.Sprintf("%v", val)
	}

	entry[p.TargetField] = jsonObjWithStringKeys
}
