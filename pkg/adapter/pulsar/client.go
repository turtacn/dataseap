package pulsar

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/turtacn/dataseap/pkg/common/errors"
	"github.com/turtacn/dataseap/pkg/config"
	"github.com/turtacn/dataseap/pkg/logger"
)

type pulsarClient struct {
	client pulsar.Client
	cfg    config.PulsarConfig
	// For managing created producers/consumers if needed for graceful shutdown
	producers []Producer
	consumers []Consumer
	mu        sync.Mutex
}

type pulsarProducer struct {
	producer pulsar.Producer
	topic    string
}

type pulsarConsumer struct {
	consumer pulsar.Consumer
}

// NewPulsarClient creates a new Pulsar client.
// NewPulsarClient 创建一个新的Pulsar客户端。
func NewPulsarClient(cfg config.PulsarConfig) (Client, error) {
	if cfg.ServiceURL == "" {
		return nil, errors.New(errors.ConfigError, "Pulsar ServiceURL is not configured")
	}

	opts := pulsar.ClientOptions{
		URL: cfg.ServiceURL,
	}
	if cfg.OperationTimeout > 0 {
		opts.OperationTimeout = time.Duration(cfg.OperationTimeout) * time.Second
	} else {
		opts.OperationTimeout = 30 * time.Second // Default
	}
	// TODO: Add other options like TLS, Auth from cfg if available
	// opts.Authentication = pulsar.NewAuthenticationTLS(...)
	// opts.TLSTrustCertsFilePath = cfg.TLSTrustCertsFilePath
	// opts.TLSAllowInsecureConnection = cfg.TLSAllowInsecureConnection

	client, err := pulsar.NewClient(opts)
	if err != nil {
		logger.L().Errorw("Failed to create Pulsar client", "url", cfg.ServiceURL, "error", err)
		return nil, errors.Wrap(err, errors.NetworkError, "failed to create Pulsar client")
	}

	logger.L().Infow("Pulsar client created successfully", "url", cfg.ServiceURL)
	return &pulsarClient{
		client: client,
		cfg:    cfg,
	}, nil
}

// CreateProducer creates a new producer.
func (pc *pulsarClient) CreateProducer(topic string, opts *ProducerOptions) (Producer, error) {
	l := logger.L().With("topic", topic)
	pulsarOpts := pulsar.ProducerOptions{
		Topic: topic,
	}
	if opts != nil {
		if opts.Name != "" {
			pulsarOpts.Name = opts.Name
		}
		if opts.SendTimeout > 0 {
			pulsarOpts.SendTimeout = opts.SendTimeout
		}
		if opts.MaxPendingMessages > 0 {
			pulsarOpts.MaxPendingMessages = opts.MaxPendingMessages
		}
		pulsarOpts.Properties = opts.Properties
		pulsarOpts.HashingScheme = opts.HashingScheme
		pulsarOpts.CompressionType = opts.CompressionType
		pulsarOpts.BatchingMaxPublishDelay = opts.BatchingMaxPublishDelay
		pulsarOpts.BatchingMaxMessages = opts.BatchingMaxMessages
	}

	producer, err := pc.client.CreateProducer(pulsarOpts)
	if err != nil {
		l.Errorw("Failed to create Pulsar producer", "error", err)
		return nil, errors.Wrapf(err, errors.InternalError, "failed to create Pulsar producer for topic %s", topic)
	}

	p := &pulsarProducer{producer: producer, topic: topic}
	pc.mu.Lock()
	pc.producers = append(pc.producers, p)
	pc.mu.Unlock()
	l.Info("Pulsar producer created successfully")
	return p, nil
}

// Subscribe creates a new consumer.
func (pc *pulsarClient) Subscribe(topics []string, subscriptionName string, opts *ConsumerOptions) (Consumer, error) {
	l := logger.L().With("topics", strings.Join(topics, ","), "subscription", subscriptionName)
	pulsarOpts := pulsar.ConsumerOptions{
		Topics:           topics,
		SubscriptionName: subscriptionName,
	}
	if opts != nil {
		if opts.Name != "" {
			pulsarOpts.Name = opts.Name
		}
		if opts.ReceiverQueueSize > 0 {
			pulsarOpts.ReceiverQueueSize = opts.ReceiverQueueSize
		}
		if opts.NackRedeliveryDelay > 0 {
			pulsarOpts.NackRedeliveryDelay = opts.NackRedeliveryDelay
		}
		if opts.AckTimeout > 0 {
			pulsarOpts.AckTimeout = opts.AckTimeout
		}
		pulsarOpts.Type = opts.SubscriptionType             // Failover, Shared, KeyShared, Exclusive
		pulsarOpts.SubscriptionMode = opts.SubscriptionMode // Default is Durable
		pulsarOpts.Properties = opts.Properties
	} else {
		pulsarOpts.Type = pulsar.Shared // Default to Shared subscription
	}

	consumer, err := pc.client.Subscribe(pulsarOpts)
	if err != nil {
		l.Errorw("Failed to subscribe to Pulsar topics", "error", err)
		return nil, errors.Wrapf(err, errors.InternalError, "failed to subscribe to Pulsar topics %v with subscription %s", topics, subscriptionName)
	}

	c := &pulsarConsumer{consumer: consumer}
	pc.mu.Lock()
	pc.consumers = append(pc.consumers, c)
	pc.mu.Unlock()
	l.Info("Pulsar consumer subscribed successfully")
	return c, nil
}

// Close closes the Pulsar client and all associated producers/consumers.
func (pc *pulsarClient) Close() {
	l := logger.L()
	l.Info("Closing Pulsar client...")
	pc.mu.Lock() // Ensure exclusive access for closing
	defer pc.mu.Unlock()

	for _, p := range pc.producers {
		p.Close()
	}
	pc.producers = nil // Clear slice

	for _, c := range pc.consumers {
		c.Close()
	}
	pc.consumers = nil // Clear slice

	pc.client.Close()
	l.Info("Pulsar client closed.")
}

// --- pulsarProducer implementation ---

func (pp *pulsarProducer) Send(ctx context.Context, msg *ProducerMessage) (MessageID, error) {
	pulsarMsg := &pulsar.ProducerMessage{
		Payload:      msg.Payload,
		Key:          msg.Key,
		Properties:   msg.Properties,
		EventTime:    msg.EventTime,
		SequenceID:   msg.SequenceID,
		OrderingKey:  msg.OrderingKey,
		DeliverAt:    msg.DeliverAt,
		DeliverAfter: msg.DeliverAfter,
	}
	msgID, err := pp.producer.Send(ctx, pulsarMsg)
	if err != nil {
		return nil, errors.Wrap(err, errors.NetworkError, "failed to send Pulsar message")
	}
	return msgID, nil
}

func (pp *pulsarProducer) SendAsync(ctx context.Context, msg *ProducerMessage, callback func(MessageID, *ProducerMessage, error)) {
	pulsarMsg := &pulsar.ProducerMessage{
		Payload:      msg.Payload,
		Key:          msg.Key,
		Properties:   msg.Properties,
		EventTime:    msg.EventTime,
		SequenceID:   msg.SequenceID,
		OrderingKey:  msg.OrderingKey,
		DeliverAt:    msg.DeliverAt,
		DeliverAfter: msg.DeliverAfter,
	}
	// The callback signature for pulsar.Producer.SendAsync is func(MessageID, *ProducerMessage, error)
	// We need to match this. Our *ProducerMessage is a wrapper, so we pass the original wrapped one.
	pp.producer.SendAsync(ctx, pulsarMsg, func(id pulsar.MessageID, m *pulsar.ProducerMessage, e error) {
		// The callback here uses the library's ProducerMessage. If the user's callback expects *our* ProducerMessage,
		// we'd need to re-wrap or adjust the interface. For now, assume the user's callback also wants our ProducerMessage.
		// This means the callback in the interface should be func(MessageID, *pulsar.ProducerMessage, error)
		// Or, we pass the original `msg` to the callback.
		callback(id, msg, e)
	})
}

func (pp *pulsarProducer) Topic() string {
	return pp.producer.Topic()
}

func (pp *pulsarProducer) Flush() error {
	return pp.producer.Flush()
}

func (pp *pulsarProducer) Close() {
	pp.producer.Close()
}

// --- pulsarConsumer implementation ---

func (pc *pulsarConsumer) Chan() <-chan pulsar.ConsumerMessage {
	return pc.consumer.Chan()
}

func (pc *pulsarConsumer) Receive(ctx context.Context) (pulsar.Message, error) {
	msg, err := pc.consumer.Receive(ctx)
	if err != nil {
		// Don't wrap context.Canceled or context.DeadlineExceeded as internal errors
		if err == context.Canceled || err == context.DeadlineExceeded {
			return nil, err
		}
		return nil, errors.Wrap(err, errors.NetworkError, "failed to receive Pulsar message")
	}
	return msg, nil
}

func (pc *pulsarConsumer) Ack(msg pulsar.Message) error {
	return pc.consumer.Ack(msg)
}

func (pc *pulsarConsumer) AckID(msgID MessageID) error {
	return pc.consumer.AckID(msgID)
}

func (pc *pulsarConsumer) Nack(msg pulsar.Message) error {
	pc.consumer.Nack(msg) // Nack in Pulsar client doesn't return error
	return nil
}

func (pc *pulsarConsumer) NackID(msgID MessageID) error {
	pc.consumer.NackID(msgID) // Nack in Pulsar client doesn't return error
	return nil
}

func (pc *pulsarConsumer) Unsubscribe() error {
	return pc.consumer.Unsubscribe()
}

func (pc *pulsarConsumer) Close() {
	pc.consumer.Close()
}
