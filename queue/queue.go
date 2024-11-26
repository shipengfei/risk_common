package queue

import (
	"context"
	"sync"

	"github.com/Shopify/sarama"
	"gitlab.miliantech.com/infrastructure/ez"
	"gitlab.miliantech.com/infrastructure/log"
	"gitlab.miliantech.com/infrastructure/rabbitmq"
	"gitlab.miliantech.com/risk/base/risk_common/utils"
)

const (
	QueueType_Rabbitmq = "rabbitmq"
	QueueType_Kafka    = "kafka"
)

type AsyncQueue struct {
	QueueName       string
	Type            string
	RabbitmqHandler func(ctx context.Context, msg rabbitmq.Delivery)
	KafkaHandler    func(ctx context.Context, msg *sarama.ConsumerMessage)
}

var QueueMap = make(map[string]*AsyncQueue)

var GWG = new(sync.WaitGroup)

func RegisterConsumer(name string, queue *AsyncQueue) {
	QueueMap[name] = queue
}

func StartConsumer() {

	for _, queue := range QueueMap {
		switch queue.Type {
		case QueueType_Rabbitmq: // 支持异步延迟队列
			go func(queue *AsyncQueue) {
				defer utils.SimpleRecover(context.Background())
				if consumer := ez.GetRabbitmqConsumer(queue.QueueName); consumer != nil {
					consumer.Consume(func(d rabbitmq.Delivery) {
						queue.RabbitmqHandler(context.Background(), d)
					})
				} else {
					log.Error(context.Background(), "consumer.notStart.R."+queue.QueueName)
				}
			}(queue)
		case QueueType_Kafka: // 支持异步队列
			go func(queue *AsyncQueue) {
				defer utils.SimpleRecover(context.Background())
				if consumer := ez.GetKafkaReader(queue.QueueName); consumer != nil && consumer.Reader != nil {
					consumer.StartReader(queue.KafkaHandler)
				} else {
					log.Error(context.Background(), "consumer.notStart.K."+queue.QueueName)
				}
			}(queue)
		}
	}

}

func StopConsumer() {
	GWG.Wait()
	for _, queue := range QueueMap {
		switch queue.Type {
		case QueueType_Rabbitmq:
			if consumer := ez.GetRabbitmqConsumer(queue.QueueName); consumer != nil {
				consumer.Close()
			}
		case QueueType_Kafka:
			if consumer := ez.GetKafkaReader(queue.QueueName); consumer != nil && consumer.Reader != nil {
				consumer.Reader.Close()
			}
		}
	}
}
