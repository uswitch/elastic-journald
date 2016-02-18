package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/uswitch/elastic-journald/journald"
	"github.com/uswitch/elastic-journald/processors"

	log "github.com/Sirupsen/logrus"
	"github.com/mattbaird/elastigo/lib"
)

type Config struct {
	Host        string
	IndexPrefix string
}

type Service struct {
	Config     *Config
	Journal    *journald.Journal
	Elastic    *elastigo.Conn
	Indexer    *elastigo.BulkIndexer
	Processors []processors.LogEntryProcessor
}

func NewService() *Service {
	config := &Config{
		Host:        *elasticHost,
		IndexPrefix: *indexPrefix,
	}

	journal, err := journald.New()
	if err != nil {
		panic("TODO")
	}

	elastic := elastigo.NewConn()
	indexer := elastic.NewBulkIndexerErrors(2, 30)
	indexer.BufferDelayMax = time.Duration(5) * time.Second
	indexer.BulkMaxDocs = 1000
	indexer.BulkMaxBuffer = 65536
	indexer.Sender = func(buf *bytes.Buffer) error {
		respJson, err := elastic.DoCommand("POST", "/_bulk", nil, buf)
		if err != nil {
			// TODO
			panic(fmt.Sprintf("Bulk error: \n%v", err))
		}

		response := struct {
			Took   int64 `json:"took"`
			Errors bool  `json:"errors"`
			Items  []struct {
				Index struct {
					Id    string `json:"_id"`
					Error string `json:"error"`
				} `json:"index"`
			} `json:"items"`
		}{}

		jsonErr := json.Unmarshal(respJson, &response)
		if jsonErr != nil {
			// TODO
			panic(jsonErr)
		}
		if response.Errors {
			for _, item := range response.Items {
				if item.Index.Error != "" {
					log.Warningf("elasticsearch reported errors on intake: %s", item.Index.Error)
				}
			}
		}

		docsStored := len(response.Items)
		lastStoredCursor := response.Items[docsStored-1].Index.Id
		ioutil.WriteFile(*cursorFile, []byte(lastStoredCursor), 0644)

		return err
	}

	service := &Service{
		Config:  config,
		Journal: journal,
		Elastic: elastic,
		Indexer: indexer,
		Processors: []processors.LogEntryProcessor{
			processors.NewDockerEnricher(),
			processors.NewJsonFieldParser("message", "parsed"),
			processors.NewStaticFields(map[string]interface{}{
				"ecs_cluster": "test",
			}),
		},
	}
	return service
}

func (s *Service) Run() error {
	//log.SetLevel(log.DebugLevel)

	s.Elastic.SetFromUrl(s.Config.Host)

	err := s.seekJournalFromCursorFile()
	if err != nil {
		return err
	}

	s.ProcessJournal()

	return nil
}

func (s *Service) seekJournalFromCursorFile() error {
	bytes, err := ioutil.ReadFile(*cursorFile)
	if err != nil {
		return err
	}

	if len(bytes) > 0 {
		cursor := string(bytes)
		err = s.Journal.Seek(cursor)
		if err != nil {
			return err
		}
	} else {
		err = s.Journal.SeekToTail()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ProcessJournal() error {
	s.Indexer.Start()
	defer s.Indexer.Stop()

	for {
		entry, err := s.Journal.Read()
		if err != nil {
			panic("TODO")
		}
		s.ProcessJournalEntry(entry)

		log.Debugf("Processed entry [%s] %v", entry.Time, entry.Fields)
	}
}

func (s *Service) ProcessJournalEntry(entry *journald.JournalEntry) {

	row := make(map[string]interface{})
	row["@timestamp"] = entry.Time.Format("2006-01-02T15:04:05.000Z")
	row["@version"] = "1"

	for key, val := range entry.Fields {
		switch key {
		case "cap_effective":
		case "cmdline":
		case "exe":
		case "machine_id":
		case "source_monotonic_timestamp":
		case "source_realtime_timestamp":
		case "syslog_facility":
		case "syslog_identifier":
		case "transport":
			continue
		default:
			row[key] = val
		}
	}

	for _, processor := range s.Processors {
		processor.Process(row)
	}

	indexName := fmt.Sprintf("%s-%s", s.Config.IndexPrefix, entry.Time.Format("2006.01.02"))
	message, _ := json.Marshal(row)

	//fmt.Printf("%s\n", indexName)
	//fmt.Printf("%s\n", message)

	log.Debugf("Indexing %s into %s\n'%s'", entry.Cursor, indexName, message)

	s.Indexer.Index(
		indexName,       // index
		"journald",      // type
		entry.Cursor,    // id
		"",              // parent
		"",              // ttl
		nil,             // date
		string(message), // content
	)

}
