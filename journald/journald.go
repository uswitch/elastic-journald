package journald

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unsafe"
)

// #include <stdio.h>
// #include <string.h>
// #include <systemd/sd-journal.h>
// #cgo pkg-config: libsystemd
import "C"

type Journal struct {
	Handle *C.sd_journal
}

type JournalEntry struct {
	Cursor string
	Time   time.Time
	Fields map[string]interface{}
}

func New() (*Journal, error) {
	j := &Journal{}
	r := C.sd_journal_open(&j.Handle, C.SD_JOURNAL_LOCAL_ONLY)
	if r < 0 {
		return nil, errors.New(fmt.Sprintf("failed to open journal: %s", C.strerror(-r)))
	}
	return j, nil
}

func (j *Journal) Close() {
	C.sd_journal_close(j.Handle)
}

func (j *Journal) Seek(cursor string) error {
	r := C.sd_journal_seek_cursor(j.Handle, C.CString(cursor))
	if r < 0 {
		return errors.New(fmt.Sprintf("failed to seek journal: %s", C.strerror(-r)))
	}
	r = C.sd_journal_next_skip(j.Handle, 1)
	if r < 0 {
		return errors.New(fmt.Sprintf("failed to skip current journal entry: %s", C.strerror(-r)))
	}
	return nil
}

func (j *Journal) SeekToTail() error {
	r := C.sd_journal_seek_tail(j.Handle)
	if r < 0 {
		return errors.New(fmt.Sprintf("failed to seek to tail of journal: %s", C.strerror(-r)))
	}
	// r = C.sd_journal_next_skip(j.Handle, 1)
	// if r < 0 {
	// 	return errors.New(fmt.Sprintf("failed to skip current journal entry: %s", C.strerror(-r)))
	// }
	return nil
}

func (j *Journal) Read() (*JournalEntry, error) {
	for {
		r := C.sd_journal_next(j.Handle)
		if r < 0 {
			return nil, errors.New(fmt.Sprintf("failed to iterate to next entry: %s", C.strerror(-r)))
		}
		if r == 0 {
			r = C.sd_journal_wait(j.Handle, 1000000)
			if r < 0 {
				return nil, errors.New(fmt.Sprintf("failed to wait for changes: %s", C.strerror(-r)))
			}
			continue
		}
		return j.readEntry()
	}
}

func (j *Journal) readEntry() (*JournalEntry, error) {
	var cursor *C.char
	r := C.sd_journal_get_cursor(j.Handle, &cursor)
	if r < 0 {
		return nil, errors.New(fmt.Sprintf("failed to get cursor: %s", C.strerror(-r)))
	}

	var realtime C.uint64_t
	r = C.sd_journal_get_realtime_usec(j.Handle, &realtime)
	if r < 0 {
		return nil, errors.New(fmt.Sprintf("failed to get realtime timestamp: %s", C.strerror(-r)))
	}
	time := time.Unix(int64(realtime/1000000), int64(realtime%1000000000)).UTC()

	fields := make(map[string]interface{})
	err := j.readFields(fields)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read fields: %s", err))
	}

	return &JournalEntry{
		Cursor: C.GoString(cursor),
		Time:   time,
		Fields: fields,
	}, nil
}

func (j *Journal) readFields(row map[string]interface{}) error {
	var length C.size_t
	var cData *C.char

	for C.sd_journal_restart_data(j.Handle); C.sd_journal_enumerate_data(j.Handle, (*unsafe.Pointer)(unsafe.Pointer(&cData)), &length) > 0; {
		data := C.GoString(cData)
		parts := strings.SplitN(data, "=", 2)
		key := strings.ToLower(parts[0])
		value := parts[1]
		row[strings.TrimPrefix(key, "_")] = value
	}

	return nil
}
