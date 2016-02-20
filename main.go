package main

import (
	log "github.com/Sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	elasticHost  = kingpin.Flag("host", "The ElasticSearch hosts to forward requests to").Required().String()
	cursorFile   = kingpin.Flag("cursor", "The file to keep cursor state between runs").Default("elastic-journald.cursor").String()
	indexPrefix  = kingpin.Flag("index-prefix", "The index prefix to use").Default("logs-test").String()
	staticFields = kingpin.Flag("static-fields", "Static fields to add to all log entries").StringMap()
)

func main() {
	kingpin.Parse()
	service := NewService()

	err := service.Run()
	if err != nil {
		log.Fatalf("%s", err)
	}
}
