/* packakge queueops wraps the gosqs methods with retry logic for common SQS operations used
by the application. */
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

// PaymentStatusFifoQueue contains the name of the Payment Status queue.
// Messages sent to this queue are used to confirm the successful completion of
// payments before processing the objects sent to the Staging queue.
const PaymentStatusFifoQueue = "stripe-tx-status.fifo"

// InventoryUpdateFifoQueue contains the name of the Inventory Update queue.
// Messages sent to this queue are used to update inventory counts when an order
// is created or cancelled.
const InventoryUpdateFifoQueue = "inventory-update.fifo"

const InventoryActionAdd = "ADD"

const InventoryActionSub = "SUB"

// FulfillmentFifoQueue contains the name of the fulfillment queue used for viewing and actioning open orders.
const FulfillmentFifoQueue = "fufillment.fifo"

// ErrMsgNotDeleted contains the error code value for the error returned by PollFulfillmentQueueForDelete
// when the message targeted for deletion is not found in the polled batch.
const ErrMsgNotDeleted = "ERR_MSG_NOT_DELETED"

// ErrEmptyQueue contains the error code for when a queue is empty/exhausted an no messages are received.
const ErrEmptyQueue = "ERR_EMPTY_QUEUE"

// Staging contains pkg store objects to be staged in the StagingQueue, which are processed
// on receipt of a StripeTxStatus message.
type Staging struct {
	Order       *store.Order       `json:"order"`
	Customer    *store.Customer    `json:"customer"`
	Transaction *store.Transaction `json:"transaction"`
}

// InventoryUpdate is used to update inventory counts when a new order is made.
type InventoryUpdate struct {
	UserEmail string            `json:"user_email"`
	OrderID   string            `json:"order_id"`
	Items     []*store.CartItem `json:"items"`
}

// return type for PollStagingQueue
type stagingPollResponse struct {
	Stages         []Staging `json:"stages"`
	MessageIDs     []string  `json:"message_ids"`
	ReceiptHandles []string  `json:"receipt_handles"`
}

// return type for PollStripeTxStatusQueue
type paymentStatusPollResponse struct {
	Statuses       []store.PaymentStatus `json:"statuses"`
	MessageIDs     []string              `json:"message_ids"`
	ReceiptHandles []string              `json:"receipt_handles"`
}

type FulfillmentPollResponsee struct {
	Orders         []*store.OrderSummary `json:"orders"`
	MessageIDs     []string              `json:"message_ids"`
	ReceiptHandles []string              `json:"receipt_handles"`
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

// SendInventoryUpdateMessage sends an InventoryUpdate message to the InventoryUpdate FIFO queue.
func SendInventoryUpdateMessage(svc interface{}, url string, update InventoryUpdate) (string, error) {
	// re-encode to JSON
	json, err := json.Marshal(update)
	if err != nil {
		log.Printf("SendInventoryUpdateMessage failed: %v", err)
		return "", err
	}

	deDupeID := gosqs.GenerateDedupeID(string(json))

	options := gosqs.SendMsgOptions{
		DelaySeconds:            gosqs.SendMsgDefault.DelaySeconds,
		MessageAttributes:       nil,
		MessageBody:             string(json),
		MessageDeduplicationId:  deDupeID,
		MessageGroupId:          deDupeID,
		MessageSystemAttributes: nil,
		QueueURL:                url,
	}

	retries := 0
	maxRetries := 2
	backoff := 1000

	for {
		resp, err := gosqs.SendMessage(svc, options)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("SendInventoryUpdateMessage failed: %v -- max retries exceeded", err)
				return "", err
			}
			log.Printf("SendInventoryUpdateMessage failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}
		return resp.MessageId, nil
	}
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

	deDupeID := gosqs.GenerateDedupeID(string(json))

	options := gosqs.SendMsgOptions{
		DelaySeconds:            gosqs.SendMsgDefault.DelaySeconds,
		MessageAttributes:       nil,
		MessageBody:             string(json),
		MessageDeduplicationId:  deDupeID,
		MessageGroupId:          deDupeID,
		MessageSystemAttributes: nil,
		QueueURL:                url,
	}

	retries := 0
	maxRetries := 4
	backoff := 1000

	for {
		resp, err := gosqs.SendMessage(svc, options)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("SendStagingMessage failed: %v -- max retries exceeded", err)
				return "", err
			}
			log.Printf("SendStagingMessage failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}
		return resp.MessageId, nil
	}
}

// SendPaymentStatusMessage sends a PaymentStatus object to the PaymentStatus Fifo queue
// and returns the message ID.
func SendPaymentStatusMessage(svc interface{}, url string, status store.PaymentStatus) (string, error) {
	// re-encode to JSON
	json, err := json.Marshal(status)
	if err != nil {
		log.Printf("SendPaymentStatusMessage failed: %v", err)
		return "", err
	}

	deDupeID := gosqs.GenerateDedupeID(string(json))

	options := gosqs.SendMsgOptions{
		DelaySeconds:            gosqs.SendMsgDefault.DelaySeconds,
		MessageAttributes:       nil,
		MessageBody:             string(json),
		MessageDeduplicationId:  deDupeID,
		MessageGroupId:          deDupeID,
		MessageSystemAttributes: nil,
		QueueURL:                url,
	}

	retries := 0
	maxRetries := 4
	backoff := 1000

	for {
		resp, err := gosqs.SendMessage(svc, options)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("SendPaymentStatusMessage failed: %v -- max retries exceeded", err)
				return "", err
			}
			log.Printf("SendPaymentStatusMessage failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}
		return resp.MessageId, nil
	}

}

// SendFulfillmentMessage sends an order to the Fulfillment queue for actioning.
func SendFulfillmentMessage(svc interface{}, url string, order *store.OrderSummary) (string, error) {
	// re-encode to JSON
	json, err := json.Marshal(order)
	if err != nil {
		log.Printf("SendInventoryUpdateMessage failed: %v", err)
		return "", err
	}

	deDupeID := gosqs.GenerateDedupeID(string(json))

	options := gosqs.SendMsgOptions{
		DelaySeconds:            gosqs.SendMsgDefault.DelaySeconds,
		MessageAttributes:       nil,
		MessageBody:             string(json),
		MessageDeduplicationId:  deDupeID,
		MessageGroupId:          deDupeID,
		MessageSystemAttributes: nil,
		QueueURL:                url,
	}

	retries := 0
	maxRetries := 2
	backoff := 1000

	for {
		resp, err := gosqs.SendMessage(svc, options)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("SendFulfillmentMessage failed: %v -- max retries exceeded", err)
				return "", err
			}
			log.Printf("SendFulfillmentMessage failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}
		return resp.MessageId, nil
	}
}

// PollStagingQueue receives 10 messages from the order staging queue.
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

// PollPaymentStatusQueue receives 3 messages from the Stripe Transaction Status queue.
func PollPaymentStatusQueue(svc interface{}, url string) (paymentStatusPollResponse, error) {
	resp := paymentStatusPollResponse{}
	idSet := make(map[string]bool) // check for duplicates
	handles := []string{}          // list of receipt handles
	messageIDs := []string{}       // list of message IDs
	statuses := []store.PaymentStatus{}

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
			status := store.PaymentStatus{}
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

// PollFulfillment receives 5 messages from the fulfillment queue.
func PollFulfillmentQueue(svc interface{}, url string) (FulfillmentPollResponsee, error) {
	resp := FulfillmentPollResponsee{}
	idSet := make(map[string]bool) // check for duplicates

	// set receive message options
	options := gosqs.RecMsgOptions{
		AttributeNames:          gosqs.RecMsgDefault.AttributeNames,
		MaxNumberOfMessages:     int64(5),
		MessageAttributeNames:   gosqs.RecMsgDefault.MessageAttributeNames,
		QueueURL:                url,
		ReceiveRequestAttemptId: gosqs.RecMsgDefault.ReceiveRequestAttemptId,
		VisibilityTimeout:       int64(90), // adjust to delete msg after actioning order
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
				log.Printf("PollFulfillmentQueue failed: %v -- max retries exceeded", err)
				return resp, err
			}
			log.Printf("PollFulfillmentQueue failed: %v -- retrying...", err)
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
			order := store.OrderSummary{}
			err := json.Unmarshal([]byte(msg.Body), &order)
			if err != nil {
				log.Printf("PollFulfillmentQueue failed - json failed to unmarshall: %v", err)
				continue
			}

			if idSet[msg.MessageId] != true {
				resp.Orders = append(resp.Orders, &order)
				resp.MessageIDs = append(resp.MessageIDs, msg.MessageId)
				resp.ReceiptHandles = append(resp.ReceiptHandles, msg.ReceiptHandle)
				idSet[msg.MessageId] = true
			}
		}

		return resp, nil
	}

}

// PollFulfillment receives 5 messages from the fulfillment queue.
func PollFulfillmentQueueForDelete(svc interface{}, url, orderID string) error {
	// set receive message options
	options := gosqs.RecMsgOptions{
		AttributeNames:          gosqs.RecMsgDefault.AttributeNames,
		MaxNumberOfMessages:     int64(10),
		MessageAttributeNames:   gosqs.RecMsgDefault.MessageAttributeNames,
		QueueURL:                url,
		ReceiveRequestAttemptId: gosqs.RecMsgDefault.ReceiveRequestAttemptId,
		VisibilityTimeout:       int64(3), // adjust to delete msg after actioning order
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
				log.Printf("PollFulfillmentQueueForDelete failed: %v -- max retries exceeded", err)
				return err
			}
			log.Printf("PollFulfillmentQueueForDelete failed: %v -- retrying...", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}
		// retry up to 2 times if no messages received w/o error response
		if len(msgs) == 0 {
			if retries > 1 {
				log.Printf("no messages received - max retries exceeded")
				return fmt.Errorf(ErrEmptyQueue)
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
			order := store.OrderSummary{}
			err := json.Unmarshal([]byte(msg.Body), &order)
			if err != nil {
				log.Printf("PollFulfillmentQueueForDelete failed - json failed to unmarshall: %v", err)
				continue
			}

			// delete message if matching orderID
			if order.OrderID == orderID {
				err := DeleteMessages(svc, url, []string{msg.MessageId}, []string{msg.ReceiptHandle})
				if err != nil {
					log.Printf("PollFulfillmentQueueForDelete failed - json failed to unmarshall: %v", err)
					return err
				}
				return nil
			}
		}

		return fmt.Errorf(ErrMsgNotDeleted)
	}

}

func ChangeMsgVisibility(svc interface{}, url string, messageIDs, handles []string, timeout int) error {
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
		TimeoutSeconds: timeout,
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
				return fmt.Errorf("MAX_RETRIES_EXCEEDED")
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

func DeleteMessages(svc interface{}, url string, messageIDs, handles []string) error {
	if len(messageIDs) != len(handles) {
		err := fmt.Errorf("INVALID_ARGS")
		log.Printf("DeleteMessages failed: %v", err)
		return err
	}
	if len(messageIDs) == 0 {
		err := fmt.Errorf("INVALID_ARGS")
		log.Printf("DeleteMessages failed: %v", err)
		return err
	}
	// cancel timeout for messages that don't match
	batchInput := gosqs.DeleteMessageBatchRequest{
		QueueURL:       url,
		MessageIDs:     messageIDs,
		ReceiptHandles: handles,
	}

	retries := 0
	maxRetries := 4
	backoff := 1000

	retryHandlesMap := make(map[string]string)
	for i, handle := range handles {
		retryHandlesMap[messageIDs[i]] = handle
	}

	for {
		output, err := gosqs.DeleteMessageBatch(svc, batchInput)
		if err != nil {
			// retry with backoff if error
			if retries > maxRetries {
				log.Printf("DeleteMessages failed: %v -- max retries exceeded", err)
				return err
			}
			log.Printf("DeleteMessages error: %v -- retrying", err)
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff = backoff * 2
			retries++
			continue
		}

		// check for failed messages
		if len(output.Failed) > 0 {
			if retries > maxRetries {
				log.Printf("DeleteMessages failed: %v -- max retries exceeded", err)
				return fmt.Errorf("MAX_RETRIES_EXCEEDED")
			}
			retryIds := []string{}
			retryHandles := []string{}
			throttled := false
			for _, msg := range output.Failed {
				log.Printf("DeleteMessage err: %v", msg.ErrorCode)
				if msg.ErrorCode == "RequestThrottled" {
					retryIds = append(retryIds, msg.MessageID)
					retryHandles = append(retryHandles, retryHandlesMap[msg.MessageID])
					throttled = true
				}
			}

			batchInput = gosqs.DeleteMessageBatchRequest{
				QueueURL:       url,
				MessageIDs:     retryIds,
				ReceiptHandles: retryHandles,
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
