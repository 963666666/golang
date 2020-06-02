package rabbitmq_delay

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql
	"github.com/jmoiron/sqlx"
)

var (
	MysqlDB         *sqlx.DB
	rabbitmqConnect *amqp.Connection
	DelayQueue      amqp.Queue
	RecvQueue       amqp.Queue
	DelayChannel    *amqp.Channel
)

func InitMysqlDB(addr string) error {
	var err error
	MysqlDB, err = sqlx.Connect("mysql", addr)
	if err != nil {
		fmt.Printf("sqlx.Connect err: %s\n", err.Error())
		return err
	}
	return nil
}

func InitRabbitmqChannel(addr string) error {
	var err error
	rabbitmqConnect, err = amqp.Dial(addr)
	if err != nil {
		log.Fatalf("amqp.Dial err: %s\n", err)
		return err
	}
	DelayChannel, err = rabbitmqConnect.Channel()
	if err != nil {
		log.Fatalf("rabbitmqConnect.Channel err: %s\n", err)
		return err
	}
	return nil
}

// 消费者
func NewConsume(pool *MQPool, consumerId string) {
	DelayChannel, err := pool.Channel()
	if err != nil {
		log.Fatalf("abbitmqConnect.Channel err: %s\n", err)
		return
	}
	// 声明一个主要使用的 exchange
	err = DelayChannel.ExchangeDeclare(
		"test.timer", // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		fmt.Printf("ch.ExchangeDeclare err: %s\n", err.Error())
		return
	}

	// 声明一个x-delayed-message 类型的exchange
	err = DelayChannel.ExchangeDeclare(
		"test.delayed.timer", // name
		"x-delayed-message",  // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		amqp.Table{
			"x-delayed-type": "direct",
		}, // arguments
	)
	if err != nil {
		fmt.Printf("ch.ExchangeDeclare x-delayed-message err: %s\n", err.Error())
		return
	}

	// 声明一个常规的队列
	q, err := DelayChannel.QueueDeclare(
		"test.delayed.timer.queue", // name
		true,                      // durable
		false,                      // delete when unused
		false,                      // exclusive
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		fmt.Printf("DelayChannel.QueueDeclare test.delayed.timer.queue err: %s\n", err.Error())
		return
	}

	err = DelayChannel.QueueBind(
		q.Name,                   // queue name, 这里指的是 test_logs
		"test.delayed.timer.key", // routing key
		"test.delayed.timer",     // exchange
		false,
		nil)
	if err != nil {
		fmt.Printf("DelayChannel.QueueBind test.delayed.timer.queue err: %s\n", err.Error())
		return
	}

	// 这里监听的是 test_logs
	msgs, err := DelayChannel.Consume(
		q.Name, // queue name, 这里指的是 test_logs
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		fmt.Printf("DelayChannel.Consume err: %s\n", err.Error())
		return
	}

	// 延时队列消费
	for {
		fmt.Println("start listenning to consume")

		select {
		case d, ok := <-msgs:
			if !ok {
				return
			}
			// 判断定时器是否被取消
			go HandleMsg(string(d.Body), consumerId)
			d.Ack(true)
		}
	}
}

// 生产者

func PublishDelayMsg(mqPool *MQPool, msg string, msgId int, delay int64) error {
	// 记录创建时的时间戳
	createTimestamp := time.Now().UnixNano()
	msg = fmt.Sprintf("%d-%d-%s-%d", msgId, createTimestamp, msg, delay)

	ch, err := mqPool.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	headers := make(amqp.Table)

	if delay != 0 {
		headers["x-delay"] = delay
	}

	err = ch.Publish(
		"test.delayed.timer",     // exchange 这里为空则不选择 exchange
		"test.delayed.timer.key", // routing key
		false,                    // mandatory
		false,                    // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			ContentType:  "application/json",
			Body:         []byte(msg),
			Headers:      headers,
		})
	if err != nil {
		return err
	}
	return nil
}

// 消费者消费消息
func HandleMsg(msg string, consumeId string) error {
	consumeTimestamp := time.Now().UnixNano()
	msgSlice := strings.Split(msg, "-")
	msgId := msgSlice[0]
	createTime := msgSlice[1]
	msgC := msgSlice[2]
	delayed := msgSlice[3]

	result, err := MysqlDB.Exec("insert into user (`msgId`, `createTime`, `msg`, `consumeTime`, `delayed`, `consumeId`) value (?, ?, ?, ?, ?, ?)", msgId, createTime, msgC, consumeTimestamp, delayed, consumeId)
	if err != nil {
		fmt.Printf("mysqlDb.exec err: %s\n", err.Error())
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("result.LastInsertId err: %s\n", err.Error())
		return err
	}
	fmt.Println("mysqlDb.exec result is", id)

	return nil
}
