package pulsar

import (
	"context"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

// MessageID represents the identifier for a message in Pulsar.
// MessageID 代表Pulsar中消息的标识符。
// This is an alias to the underlying Pulsar client's MessageID for convenience.
// 这是底层Pulsar客户端MessageID的别名，以方便使用。
type MessageID = pulsar.MessageID

// Message represents a message received from Pulsar.
// Message 代表从Pulsar接收到的消息。
// This can be an alias or a wrapper around pulsar.Message.
// For simplicity, we'll assume methods to access payload, properties, etc.
type Message interface {
	ID() MessageID
	Payload() []byte
	Properties() map[string]string
	PublishTime() time.Time
	EventTime() time.Time
	Key() string
	Topic() string
	// RedeliveryCount() uint32 // Available in pulsar.Message
	// GetSchemaValue(v interface{}) error // Available in pulsar.Message
}

// ProducerMessage represents a message to be sent to Pulsar.
// ProducerMessage 代表要发送到Pulsar的消息。
type ProducerMessage struct {
	Payload      []byte
	Key          string
	Properties   map[string]string
	EventTime    time.Time
	SequenceID   *int64
	OrderingKey  string
	DeliverAt    time.Time
	DeliverAfter time.Duration
	// Schema interface{} // If using schema
}

// ProducerOptions holds options for creating a Pulsar producer.
// ProducerOptions 保存创建Pulsar生产者的选项。
type ProducerOptions struct {
	Name                    string
	Properties              map[string]string
	SendTimeout             time.Duration
	MaxPendingMessages      int
	HashingScheme           pulsar.HashingScheme
	CompressionType         pulsar.CompressionType
	BatchingMaxPublishDelay time.Duration
	BatchingMaxMessages     uint
	// ... other pulsar.ProducerOptions fields as needed
}

// ConsumerOptions holds options for creating a Pulsar consumer.
// ConsumerOptions 保存创建Pulsar消费者的选项。
type ConsumerOptions struct {
	Name                string
	SubscriptionType    pulsar.SubscriptionType
	SubscriptionMode    pulsar.SubscriptionMode // Default is Durable
	Properties          map[string]string
	ReceiverQueueSize   int
	NackRedeliveryDelay time.Duration
	AckTimeout          time.Duration
	// ... other pulsar.ConsumerOptions fields as needed
}

// Producer defines the interface for a Pulsar message producer.
// Producer 定义了Pulsar消息生产者的接口。
type Producer interface {
	// Send sends a message to Pulsar.
	// Send 发送一条消息到Pulsar。
	Send(ctx context.Context, msg *ProducerMessage) (MessageID, error)

	// SendAsync sends a message asynchronously. Callback will be invoked with MessageID and error.
	// SendAsync 异步发送消息。回调函数将被MessageID和错误调用。
	SendAsync(ctx context.Context, msg *ProducerMessage, callback func(MessageID, *ProducerMessage, error))

	// Topic returns the topic this producer is producing to.
	// Topic 返回此生产者正在生产的主题。
	Topic() string

	// Flush flushes all pending messages.
	// Flush 刷新所有待处理的消息。
	Flush() error

	// Close closes the producer and releases resources.
	// Close 关闭生产者并释放资源。
	Close()
}

// Consumer defines the interface for a Pulsar message consumer.
// Consumer 定义了Pulsar消息消费者的接口。
type Consumer interface {
	// Chan returns a channel to consume messages from.
	// Chan 返回一个用于消费消息的通道。
	Chan() <-chan pulsar.ConsumerMessage // Exposing underlying type for direct use

	// Receive blocks until a message is received or context is cancelled.
	// Receive 阻塞直到接收到消息或上下文被取消。
	Receive(ctx context.Context) (pulsar.Message, error) // Exposing underlying type

	// Ack acknowledges a message.
	// Ack 确认一条消息。
	Ack(msg pulsar.Message) error // Exposing underlying type
	AckID(msgID MessageID) error

	// Nack negatively acknowledges a message, causing it to be redelivered.
	// Nack 否定确认一条消息，使其被重新投递。
	Nack(msg pulsar.Message) error // Exposing underlying type
	NackID(msgID MessageID) error

	// Unsubscribe unsubscribes the consumer.
	// Unsubscribe 取消订阅此消费者。
	Unsubscribe() error

	// Close closes the consumer and releases resources.
	// Close 关闭消费者并释放资源。
	Close()
}

// Client defines the interface for a Pulsar client.
// Client 定义了Pulsar客户端的接口。
type Client interface {
	// CreateProducer creates a new producer for the given topic.
	// CreateProducer 为指定主题创建一个新的生产者。
	CreateProducer(topic string, opts *ProducerOptions) (Producer, error)

	// Subscribe creates a new consumer for the given topics and subscription.
	// Subscribe 为指定主题和订阅创建一个新的消费者。
	Subscribe(topics []string, subscriptionName string, opts *ConsumerOptions) (Consumer, error)

	// Close closes the client and all associated producers/consumers.
	// Close 关闭客户端及所有相关的生产者/消费者。
	Close()
}
