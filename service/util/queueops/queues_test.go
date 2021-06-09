package queueops

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/tpillz-presents/service/store-api/store"
)

func TestSendStagingMessage(t *testing.T) {
	var tests = []struct {
		input Staging
		want  error
	}{
		{
			input: Staging{
				Order:       &store.Order{OrderID: "001"},
				Customer:    &store.Customer{UserID: "001"},
				Transaction: &store.Transaction{TransactionID: "001"},
			},
			want: nil,
		},
		{
			input: Staging{
				Order:       &store.Order{OrderID: "002"},
				Customer:    &store.Customer{UserID: "002"},
				Transaction: &store.Transaction{TransactionID: "002"},
			},
			want: nil,
		},
		{
			input: Staging{
				Order:       &store.Order{},
				Customer:    &store.Customer{},
				Transaction: &store.Transaction{},
			},
			want: nil,
		},
	}
	for _, test := range tests {
		msgID, err := SendStagingMessage(test.input)
		if err != test.want {
			t.Errorf("FAIL: %v", err)
		} else {
			t.Logf("message ID: %s", msgID)
		}
	}
}

// ~ 0.084s / op
func BenchmarkSendStagingMessage(b *testing.B) {
	stages := []Staging{}
	for i := 0; i < 100; i++ {
		id := strconv.Itoa(i)
		stage := Staging{
			Order:       &store.Order{OrderID: id},
			Customer:    &store.Customer{UserID: id},
			Transaction: &store.Transaction{TransactionID: id},
		}
		stages = append(stages, stage)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := SendStagingMessage(stages[i])
		if err != nil {
			b.Errorf("FAIL: %v", err)
		}
	}
}

func TestSendStripeTxStatusMessage(t *testing.T) {
	var tests = []struct {
		input StripeTxStatus
		want  error
	}{
		{
			input: StripeTxStatus{
				OrderID: "001",
			},
			want: nil,
		},
		{
			input: StripeTxStatus{
				OrderID: "001",
			},
			want: nil,
		},
		{
			input: StripeTxStatus{
				// nil
			},
			want: nil,
		},
	}
	for _, test := range tests {
		msgID, err := SendStripeTxStatusMessage(test.input)
		if err != test.want {
			t.Errorf("FAIL: %v", err)
		} else {
			t.Logf("message ID: %s", msgID)
		}
	}
}

// ~ .080 - .095s / op
func BenchmarkSendStripeTxStatusMessageA(b *testing.B) {
	statuses := []StripeTxStatus{}
	for i := 0; i < 100; i++ {
		id := strconv.Itoa(i)
		status := StripeTxStatus{
			OrderID:        id,
			StripeTxID:     id,
			StageMessageID: id,
			TxStatus:       "SUCCESS",
			TxMessage:      "Message ok",
		}
		statuses = append(statuses, status)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := SendStripeTxStatusMessage(statuses[i])
		if err != nil {
			b.Errorf("FAIL: %v", err)
		}
	}
}

// no significant difference in reduced message size
func BenchmarkSendStripeTxStatusMessageB(b *testing.B) {
	statuses := []StripeTxStatus{}
	for i := 0; i < 100; i++ {
		id := strconv.Itoa(i)
		status := StripeTxStatus{
			OrderID: id,
		}
		statuses = append(statuses, status)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := SendStripeTxStatusMessage(statuses[i])
		if err != nil {
			b.Errorf("FAIL: %v", err)
		}
	}
}

func TestPollStagingQueue(t *testing.T) {
	type testInput struct {
		input  string // message ID
		wantId string
		want   error
	}
	var tests = []testInput{}
	for i := 0; i < 30; i++ {
		id := strconv.Itoa(i)
		input := testInput{"", id, nil}
		stage := Staging{
			Order:       &store.Order{OrderID: id},
			Customer:    &store.Customer{UserID: id},
			Transaction: &store.Transaction{TransactionID: id},
		}
		msgId, err := SendStagingMessage(stage)
		if err != nil {
			t.Errorf("FAIL: %v", err)
		}
		input.input = msgId
		tests = append(tests, input)
	}
	for _, test := range tests {
		resp, err := PollStagingQueue(test.input)
		if err != test.want {
			t.Errorf("FAIL: %v", err)
		}
		if len(resp.Stages) == 0 {
			t.Errorf("FAIL: stage not found")
		}
		for _, stage := range resp.Stages {
			if stage.Order.OrderID != test.wantId {
				t.Errorf("FAIL: stage doesn't match input (%v)", test.input)
			}
		}
	}
}

type pollStagingTestInput struct {
	input  string // message ID
	wantId string
	want   error
}

func TestPollStagingQueueConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	var tests = []pollStagingTestInput{}
	for i := 0; i < 30; i++ {
		id := strconv.Itoa(i)
		input := pollStagingTestInput{"", id, nil}
		stage := Staging{
			Order:       &store.Order{OrderID: id},
			Customer:    &store.Customer{UserID: id},
			Transaction: &store.Transaction{TransactionID: id},
		}
		msgId, err := SendStagingMessage(stage)
		if err != nil {
			t.Errorf("FAIL: %v", err)
		}
		input.input = msgId
		tests = append(tests, input)
	}
	start := time.Now()
	for _, test := range tests {
		wg.Add(1)
		go pollConcurrent(test, t, &wg)
	}
	wg.Wait()
	fin := time.Since(start)
	t.Logf("goroutine time: %v", fin)
}

func pollConcurrent(test pollStagingTestInput, t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := PollStagingQueue(test.input)
	if err != test.want {
		t.Errorf("FAIL: %v", err)
	}
	if len(resp.Stages) == 0 {
		t.Errorf("FAIL: stage not found")
	}
	for _, stage := range resp.Stages {
		if stage.Order.OrderID != test.wantId {
			t.Errorf("FAIL: stage doesn't match input (%v)", test.input)
		} else {
			t.Logf("success: %v", test.wantId)
		}
	}
}
