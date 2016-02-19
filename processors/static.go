package processors

type StaticFields struct {
	Fields map[string]interface{}
}

func NewStaticFields(fields map[string]interface{}) *StaticFields {
	return &StaticFields{
		Fields: fields,
	}
}

func (p *StaticFields) Process(entry LogEntry) {
	for k, v := range p.Fields {
		entry[k] = v
	}
}
