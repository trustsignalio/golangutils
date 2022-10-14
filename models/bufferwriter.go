package models

import (
	"context"
	"encoding/json"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/trustsignalio/go-lang-utils/messaging"
)

var mu sync.Mutex

const maxBufferDocs = 400

// BufferWriter struct holds the elements docs and ids till they are
// flushed to database either manually using a timed event or when
// the count of the resource exceeds a certain threshold value
type BufferWriter struct {
	Docs          []interface{}
	Table         string
	Host          string
	Count         int
	MessageClient *messaging.Message

	Conn *mongo.Database
}

// NewBufferWriter creates a writer object for all the subsequents calls
func NewBufferWriter() *BufferWriter {
	w := BufferWriter{}
	return &w
}

// InsertDocs method appends the click to the w.docs slice and
// increments the click count by creating a read write lock to
// handle concurrent writes to the same object
func (w *BufferWriter) InsertDocs(d interface{}) {
	mu.Lock()
	w.Docs = append(w.Docs, d)
	w.Count++
	mu.Unlock()

	w.BulkInsert(false)
}

func (w *BufferWriter) retryInsert(docs []interface{}) {
	if w.MessageClient == nil || len(docs) == 0 {
		return
	}
	var jsonMap = make(map[string]interface{})
	jsonMap["table"] = w.Table
	jsonMap["host"] = w.Host
	jsonMap["docs"] = docs
	var jsonStr, jsonErr = json.Marshal(jsonMap)
	if jsonErr == nil {
		w.MessageClient.Send(jsonStr)
	}
}

func (w *BufferWriter) insertInDb(docs []interface{}) {
	if len(docs) == 0 {
		return
	}
	var ordered = false
	opts := &options.InsertManyOptions{
		Ordered: &ordered,
	}
	_, err := w.Conn.Collection(w.Table).InsertMany(context.Background(), docs, opts)
	if err != nil {
		w.retryInsert(docs)
	}
}

// BulkInsert checks the docs count and if the count exceeds the
// max buffer count then the documents are flushed to the database
// and all the counters are reset
func (w *BufferWriter) BulkInsert(override bool) {
	mu.Lock()
	var objCount = w.Count
	mu.Unlock()
	if override || objCount > maxBufferDocs {
		mu.Lock()
		var docs = w.Docs
		w.Docs = nil
		w.Count = 0
		mu.Unlock()

		if len(docs) == 0 { // since the number of docs could be zero because of override flag
			return
		}
		w.insertInDb(docs)
	}
}

// Flush manually flushes the data to database overriding the checks
func (w *BufferWriter) Flush() {
	w.BulkInsert(true)
}
