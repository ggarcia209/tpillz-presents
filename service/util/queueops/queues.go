package queueops

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-aws/go-sqs/gosqs"
	"github.com/tpillz-presents/service/store-api/store"
)

// StagingFifoQueue contains the queue name of the order staging queue.
// Messages sent to this queue are store.Order, store.Customer, & store.Transaction
// objects awaiting processing pending receipt of a StripeTxStatus message.
const StagingFifoQueue = "staging-queue.fifo"

// StripeTxStatusFifoQueue contains the name of the Stripe Transaction Status queue.
// Messages sent to this queue are used to confirm the successful completion of
// Stripe charges before processing the objects sent to the Staging queue.
const StripeTxStatusFifoQueue = "stripe-tx-status.fifo"

// StripeTxStatusSuccess contains the status code for successful stripe transactions.
const StripeTxStatusSuccess = "STRIPE_TX_SUCCESS"

// StripeTxStatusFail contains the status code for failed stripe transactions.
const StripeTxStatusFail = "STRIPE_TX_FAIL"

// ValidStripeTxStatus contains the valid values for Stripe transaction statuses.
var ValidStripeTxStatus = map[string]bool{
	StripeTxStatusFail:    true,
	StripeTxStatusSuccess: true,
}

// Staging contains pkg store objects to be staged in the StagingQueue, which are processed
// on receipt of a StripeTxStatus message.
type Staging struct {
	Order       *store.Order       `json:"order"`
	Customer    *store.Customer    `json:"customer"`
	Transaction *store.Transaction `json:"transaction"`
}

// StripeTxStatus contains message info to send to the StripeTxStatus fifo queue.
// Objects staged in the Staging fifo queue are processed on receipt of this message
type StripeTxStatus struct {
	OrderID        string `json:"order_id"`
	StripeTxID     string `json:"stripe_tx_id"`
	StageMessageID string `json:"stage_message_id"`
	TxStatus       string `json:"tx_status"`
	TxMessage      string `json:"tx_message"`
}

// return type for PollStagingQueue
type stagingPollResponse struct {
	Stages         []Staging `json:"stages"`
	MessageIDs     []string  `json:"message_ids"`
	ReceiptHandles []string  `json:"receipt_handles"`
}

// return type for PollStripeTxStatusQueue
type stripeTxStatusPollResponse struct {
	Statuses       []StripeTxStatus `json:"statuses"`
	MessageIDs     []string         `json:"message_ids"`
	ReceiptHandles []string         `json:"receipt_handles"`
}

// InitSesh wraps the gosqs.InitSesh() method.
func InitSesh() interface{} {
	return gosqs.InitSesh()
}

// GetQueueURL wraps the gosqs.GetQueueURL method.
func GetQueueURL(svc interface{}, queueName string) (string, error) {
	url, err := gosqs.GetQueueURL(svc, queueName)
	if err != nil {
		log.Printf("GetQueueURL failed: %v", err)
		return "", err
	}
	return url, nil
}

// SendStagingMessage sends a Staging object to the Staging Fifo queue
// and returns the message ID.
func SendStagingMessage(svc interface{}, url string, stage Staging) (string, error) {
	// re-encode to JSON
	json, err := json.Marshal(stage)
	if err != nil {
		log.Printf("SendStagingMessage failed: %v", err)
		return "", err
	}

	options := gosqs.SendMsgOptions{
		DelaySeconds:            gosqs.SendMsgDefault.DelaySeconds,
		MessageAttributes:       nil,
		MessageBody:             string(json),
		MessageDeduplicationId:  gosqs.GenerateDedupeID(url),
		MessageGroupId:          gosqs.GenerateDedupeID(url),
		MessageSystemAttributes: nil,
		QueueURL:                url,
	}

	resp, err := gosqs.SendMessage(svc, options)
	if err != nil {
		log.Printf("SendStagingMessage failed: %v", err)
		return "", err
	}
	return resp.MessageId, nil
}

// SendStripeTxStatusMessage sends a StripeTxStatus object to the StripeTxStatus Fifo queue
// and returns the message ID.
func SendStripeTxStatusMessage(svc interface{}, url string, status StripeTxStatus) (string, error) {
	// re-encode to JSON
	json, err := json.Marshal(status)
	if err != nil {
		log.Printf("SendStripeTxStatusMessage failed: %v", err)
		return "", err
	}

	options := gosqs.SendMsgOptions{
		DelaySeconds:            gosqs.SendMsgDefault.DelaySeconds,
		MessageAttributes:       nil,
		MessageBody:             string(json),
		MessageDeduplicationId:  gosqs.GenerateDedupeID(url),
		MessageGroupId:          gosqs.GenerateDedupeID(url),
		MessageSystemAttributes: nil,
		QueueURL:                url,
	}

	resp, err := gosqs.SendMessage(svc, options)
	if err != nil {
		log.Printf("SendStripeTxStatusMessage failed: %v", err)
		return "", err
	}
	return resp.MessageId, nil
}

// add error handling for duplicate queue creation
func PollStagingQueue(svc interface{}, url string) (stagingPollResponse, error) {
	resp := stagingPollResponse{}
	idSet := make(map[string]bool) // check for duplicates

	// set receive message options
	options := gosqs.RecMsgOptions{
		AttributeNames:          gosqs.RecMsgDefault.AttributeNames,
		MaxNumberOfMessages:     int64(10),
		MessageAttributeNames:   gosqs.RecMsgDefault.MessageAttributeNames,
		QueueURL:                url,
		ReceiveRequestAttemptId: gosqs.RecMsgDefault.ReceiveRequestAttemptId,
		VisibilityTimeout:       int64(10),
		WaitTimeSeconds:         int64(0),
	}

	// poll for messages with exponential backoff for errors & empty responses
	retries := 0
	maxRetries := 4
	backoff := 1000.0
	for {
		// receive messages from queue
		msgs, err := gosqs.ReceiveMessage(svc, options)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("PollStagingQueue failed: %v -- max retries exceeded", err)
				return resp, err
			}
			log.Printf("PollStagingQueue failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}
		// retry up to 2 times if no messages received w/o error response
		if len(msgs) == 0 {
			if retries > 1 {
				log.Printf("no messages received - max retries exceeded")
				return resp, nil
			}
			log.Printf("no messages received - retrying...")
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		} else {
			log.Printf("messages received: %v", len(msgs))
		}

		for _, msg := range msgs {
			// get stage info from json in message body
			stage := Staging{}
			err := json.Unmarshal([]byte(msg.Body), &stage)
			if err != nil {
				log.Printf("PollStagingQueue failed - json failed to unmarshall: %v", err)
				continue
			}

			if idSet[msg.MessageId] != true {
				resp.Stages = append(resp.Stages, stage)
				resp.MessageIDs = append(resp.MessageIDs, msg.MessageId)
				resp.ReceiptHandles = append(resp.ReceiptHandles, msg.ReceiptHandle)
				idSet[msg.MessageId] = true
			}
		}

		return resp, nil
	}

}

// add error handling for duplicate queue creation
func PollStripeTxStatusQueue(svc interface{}, url string) (stripeTxStatusPollResponse, error) {
	resp := stripeTxStatusPollResponse{}
	idSet := make(map[string]bool) // check for duplicates
	handles := []string{}          // list of receipt handles
	messageIDs := []string{}       // list of message IDs
	statuses := []StripeTxStatus{}

	// set receive message options
	options := gosqs.RecMsgOptions{
		AttributeNames:          gosqs.RecMsgDefault.AttributeNames,
		MaxNumberOfMessages:     int64(3),
		MessageAttributeNames:   gosqs.RecMsgDefault.MessageAttributeNames,
		QueueURL:                url,
		ReceiveRequestAttemptId: gosqs.RecMsgDefault.ReceiveRequestAttemptId,
		VisibilityTimeout:       int64(90), // VisibilityTimeout must be greater than total retry period to prevent receiving same message twice
		WaitTimeSeconds:         int64(3),  // enable long polling (dev: 3, prod: 10)
	}

	// poll for messages with exponential backoff for errors & empty responses
	retries := 0
	maxRetries := 4
	backoff := 1000
	for {
		// receive messages from queue
		msgs, err := gosqs.ReceiveMessage(svc, options)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("PollStripeTxStatusQueue failed: %v -- max retries exceeded", err)
				return resp, err
			}
			log.Printf("PollStripeTxStatusQueue failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}
		// retry up to 2 times if no messages received w/o error response
		if len(msgs) == 0 {
			if retries > 2 {
				log.Printf("no messages received - max retries exceeded")
				return resp, nil
			}
			log.Printf("no messages received - retrying...")
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		} else {
			log.Printf("messages received: %v", len(msgs))
		}
		// get stage info from json in message body
		for _, msg := range msgs {
			status := StripeTxStatus{}
			err := json.Unmarshal([]byte(msg.Body), &status)
			if err != nil {
				log.Printf("PollStripeTxStatusQueue failed - json failed to unmarshall: %v (%v)", err, msg.MessageId)
				continue
			}
			if idSet[msg.MessageId] != true {
				messageIDs = append(messageIDs, msg.MessageId)
				statuses = append(statuses, status)
				handles = append(handles, msg.ReceiptHandle)
				idSet[msg.MessageId] = true
			}
		}

		resp.Statuses, resp.MessageIDs, resp.ReceiptHandles = statuses, messageIDs, handles
		return resp, nil
	}
}

func ChangeMsgVisibility(svc interface{}, url string, messageIDs, handles []string) error {
	if len(messageIDs) != len(handles) {
		err := fmt.Errorf("INVALID_ARGS")
		log.Printf("ChangeMsgVisibility failed: %v", err)
		return err
	}
	if len(messageIDs) == 0 {
		err := fmt.Errorf("INVALID_ARGS")
		log.Printf("ChangeMsgVisibility failed: %v", err)
		return err
	}
	// cancel timeout for messages that don't match
	batchInput := gosqs.BatchUpdateVisibilityTimeoutRequest{
		QueueURL:       url,
		MessageIDs:     messageIDs,
		ReceiptHandles: handles,
		TimeoutSeconds: 0,
	}

	retries := 0
	maxRetries := 4
	backoff := 1000

	retryHandlesMap := make(map[string]string)
	for i, handle := range handles {
		retryHandlesMap[messageIDs[i]] = handle
	}

	for {
		output, err := gosqs.ChangeMessageVisibilityBatch(svc, batchInput)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("ChangeStageMsgVisibility failed: %v -- max retries exceeded", err)
				return err
			}
			log.Printf("ChangeMessageVisibilityBatch error: %v -- retrying", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}

		// check for failed messages
		if len(output.Failed) > 0 {
			if retries > maxRetries {
				log.Printf("ChangeMsgVisibility failed: %v -- max retries exceeded", err)
				return err
			}
			retryIds := []string{}
			retryHandles := []string{}
			throttled := false
			for _, msg := range output.Failed {
				log.Printf("change messagage visiblity err: %v", msg.ErrorCode)
				if msg.ErrorCode == "RequestThrottled" {
					retryIds = append(retryIds, msg.MessageId)
					retryHandles = append(retryHandles, retryHandlesMap[msg.MessageId])
					throttled = true
				}
			}

			batchInput = gosqs.BatchUpdateVisibilityTimeoutRequest{
				QueueURL:       url,
				MessageIDs:     retryIds,
				ReceiptHandles: retryHandles,
				TimeoutSeconds: 0,
			}

			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++

			if throttled {
				log.Printf("Retrying throttled request...")
			} else {
				log.Printf("Retrying failed request...")
			}

			continue
		}
		return nil
	}

}
