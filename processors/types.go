package processors

type LogEntry map[string]interface{}

type LogEntryProcessor interface {
	Process(entry LogEntry)
}
