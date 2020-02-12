package consume

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jdextraze/go-gesclient/client"
	uuid "github.com/satori/go.uuid"

	"github.com/SpeedVan/go-common-eventstore/client/eventstore"
	"github.com/SpeedVan/go-common-faas/constant/httpconst"
	"github.com/SpeedVan/go-common-faas/struct/eventstruct"
	"github.com/SpeedVan/runtime-proxy/service"
	"github.com/SpeedVan/runtime-proxy/service/localhttpcall"
	"github.com/SpeedVan/runtime-proxy/worker/stage"
)

// StageImpl todo
type StageImpl struct {
	stage.Stage
	NextStage     stage.Stage
	Stream        string
	CallPort      string
	Client        *eventstore.Client
	LocalHTTPCall service.Call
	CloseFunc     func() error
}

// New todo
func New(stream, port string, httpclient *http.Client) (*StageImpl, error) {
	c, err := eventstore.New(stream, false, "tcp://admin:changeit@10.121.117.207:1113", "", true, false)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("PUT", "http://admin:changeit@10.121.117.207:2113/subscriptions/"+stream+"/Computer", nil)
	req.Header.Set("Content-Type", "application/json")
	res, err := httpclient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(res.StatusCode)
	}

	//注册常规事件
	c.Connected().Add(func(evt client.Event) error { log.Printf("Connected: %+v", evt); return nil })
	c.Disconnected().Add(func(evt client.Event) error { log.Printf("Disconnected: %+v", evt); return nil })
	c.Reconnecting().Add(func(evt client.Event) error { log.Printf("Reconnecting: %+v", evt); return nil })
	c.Closed().Add(func(evt client.Event) error { log.Fatalf("Connection closed: %+v", evt); return nil })
	c.ErrorOccurred().Add(func(evt client.Event) error { log.Printf("Error: %+v", evt); return nil })
	c.AuthenticationFailed().Add(func(evt client.Event) error { log.Printf("Auth failed: %+v", evt); return nil })

	return &StageImpl{
		Stream:        stream,
		CallPort:      port,
		Client:        c,
		LocalHTTPCall: localhttpcall.New(port, httpclient),
	}, nil
}

// Do todo
func (s *StageImpl) Do() error {
	task, err := s.Client.ConnectToPersistentSubscriptionAsync(s.Stream, "Computer", s.eventAppeared, subscriptionDropped, nil, 100, true)

	if err != nil {
		log.Printf("Error occured while subscribing to stream: %v", err)
	} else if err := task.Error(); err != nil {
		log.Printf("Error occured while waiting for result of subscribing to stream: %v", err)
	} else {
		sub := task.Result().(client.PersistentSubscription)
		log.Printf("SubscribeToStream result: %+v", sub)
		s.CloseFunc = func() error { return sub.Stop() }
	}
	return nil
}

// Close todo
func (s *StageImpl) Close() error {
	return s.CloseFunc()
}

// Next todo
func (s *StageImpl) Next() stage.Stage {
	return s.NextStage
}

func (s *StageImpl) eventAppeared(_ client.PersistentSubscription, e *client.ResolvedEvent) error {
	bs := e.Event().Data()
	id := e.Event().EventId().String()
	log.Printf("event received, id: %s", id)
	requestEventMetadata, err := eventstruct.FromHTTPRequestJSONBytes(e.Event().Metadata())
	if err != nil {
		return err
	}

	ctx := requestEventMetadata.Context
	reqHeader := requestEventMetadata.Header.Clone()
	traceID := fmt.Sprint(ctx[httpconst.TraceID])
	requestID := fmt.Sprint(ctx[httpconst.RequestID])
	reqHeader.Set(httpconst.TraceID, traceID)
	reqHeader.Set(httpconst.RequestID, requestID)

	statusCode, status, resHeader, resBody, err := s.LocalHTTPCall.Call(requestEventMetadata.Method, requestEventMetadata.Path, reqHeader, bytes.NewReader(bs))
	data := []byte{}
	if err != nil {
		data = []byte(err.Error())
	} else {
		data, _ = ioutil.ReadAll(resBody)
	}
	resMeta := &eventstruct.ResponseEventMetadata{
		Context:        ctx,
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

// Sink todo
func (s *StageImpl) Sink(streamName, eventType string, metadata, data []byte) {
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

func subscriptionDropped(_ client.PersistentSubscription, r client.SubscriptionDropReason, err error) error {
	log.Printf("subscription dropped: %s, %v", r, err)
	return nil
}
