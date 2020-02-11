package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/jdextraze/go-gesclient/client"
	uuid "github.com/satori/go.uuid"

	"github.com/SpeedVan/go-common-eventstore/client/eventstore"
	"github.com/SpeedVan/runtime-proxy/service"
)

// CallConsumer todo
type CallConsumer struct {
	Client        *eventstore.Client
	LocalHTTPCall service.Call
	StreamName    string
	SubStopFunc   func()
}

// New todo
func New(name string, endpoint string, localhttpcall service.Call, StreamName string) (*CallConsumer, error) {
	c, err := eventstore.New(name, false, endpoint, "", true, false)
	if err != nil {
		return nil, err
	}

	//注册常规事件
	c.Connected().Add(func(evt client.Event) error { log.Printf("Connected: %+v", evt); return nil })
	c.Disconnected().Add(func(evt client.Event) error { log.Printf("Disconnected: %+v", evt); return nil })
	c.Reconnecting().Add(func(evt client.Event) error { log.Printf("Reconnecting: %+v", evt); return nil })
	c.Closed().Add(func(evt client.Event) error { log.Fatalf("Connection closed: %+v", evt); return nil })
	c.ErrorOccurred().Add(func(evt client.Event) error { log.Printf("Error: %+v", evt); return nil })
	c.AuthenticationFailed().Add(func(evt client.Event) error { log.Printf("Auth failed: %+v", evt); return nil })

	return &CallConsumer{
		Client:        c,
		LocalHTTPCall: localhttpcall,
		StreamName:    StreamName,
	}, nil
}

// Subscribe todo
func (s *CallConsumer) Subscribe() {
	fmt.Printf("all streams: %v", s.StreamName)
	// for _, sn := range s.StreamName {
	// 	stream := "fission-edit+" + sn
	// 	fmt.Printf("stream: %v", stream)
	// 	s.Source(stream)
	// }
	s.Source(s.StreamName)
}

// Source todo
func (s *CallConsumer) Source(streamName string) {
	task, err := s.Client.ConnectToPersistentSubscriptionAsync(streamName, "script_1", s.eventAppeared, subscriptionDropped, nil, 100, true)

	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else if err := task.Error(); err != nil {
		log.Printf("Error occured while waiting for result of subscribing to stream: %v", err)
	} else {
		sub := task.Result().(client.PersistentSubscription)
		log.Printf("SubscribeToStream result: %+v", sub)
		s.SubStopFunc = func() { sub.Stop() }
	}
}

// Sink todo
func (s *CallConsumer) Sink(streamName, eventType string, metadata, data []byte) {
	id := uuid.Must(uuid.NewV4())
	evt := client.NewEventData(id, eventType, true, data, metadata)
	log.Printf("event sent, id: %s", id)
	task, err := s.Client.AppendToStreamAsync(streamName, client.ExpectedVersion_Any, []*client.EventData{evt}, nil)
	if err != nil {
		log.Printf("Error occured while appending to stream: %v", err)
	} else if err := task.Error(); err != nil {
		log.Printf("Error occured while waiting for result of appending to stream: %v", err)
	} else {
		result := task.Result().(*client.WriteResult)
		log.Printf("<- %+v", result)
	}
}

// Close close client
func (s *CallConsumer) Close() error {
	s.SubStopFunc()
	err := s.Client.Close()
	time.Sleep(10 * time.Millisecond)
	return err
}

// RequestEventMetadata todo
type RequestEventMetadata struct {
	ResponseStreamName string                 `json:"responseStreamName"`
	Context            map[string]interface{} `json:"context"`
	Method             string                 `json:"method"`
	Path               string                 `json:"path"`
	Header             http.Header            `json:"header"`
}

// ResponseEventMetadata todo
type ResponseEventMetadata struct {
	Context        map[string]interface{}
	RequestEventID string
	StatusCode     int
	Status         string
	Header         http.Header
}

func (s *CallConsumer) eventAppeared(_ client.PersistentSubscription, e *client.ResolvedEvent) error {
	bs := e.Event().Data()
	id := e.Event().EventId().String()
	log.Printf("event received, id: %s", id)
	requestEventMetadata := &RequestEventMetadata{}
	err := json.Unmarshal(e.Event().Metadata(), requestEventMetadata)
	if err != nil {
		return err
	}

	ctx := requestEventMetadata.Context
	reqHeader := requestEventMetadata.Header.Clone()
	reqHeader.Set("X-Trace-Id", fmt.Sprint(ctx["X-Trace-Id"]))
	reqHeader.Set("X-Request-Id", fmt.Sprint(ctx["X-Request-Id"]))

	statusCode, status, resHeader, resBody, err := s.LocalHTTPCall.Call(requestEventMetadata.Method, requestEventMetadata.Path, reqHeader, bytes.NewReader(bs))
	data := []byte{}
	if err != nil {
		data = []byte(err.Error())
	} else {
		data, _ = ioutil.ReadAll(resBody)
	}
	resMeta := &ResponseEventMetadata{
		Context:        requestEventMetadata.Context,
		RequestEventID: id,
		StatusCode:     statusCode,
		Status:         status,
		Header:         resHeader,
	}

	metadata, _ := json.Marshal(resMeta)

	responseStreamName := requestEventMetadata.ResponseStreamName
	log.Printf("event received, responseStreamName: %s", responseStreamName)
	if responseStreamName == "" {
		responseStreamName = "DefaultResult"
	}

	s.Sink(responseStreamName, "EventResult", metadata, data)
	return nil
}

// func (s *CallConsumer) handleRequest(meta *RequestEventMetadata, bs)

func subscriptionDropped(_ client.PersistentSubscription, r client.SubscriptionDropReason, err error) error {
	log.Printf("subscription dropped: %s, %v", r, err)
	return nil
}
