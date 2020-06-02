package main

import (
	"golua_timer/rabbitmq_delay"
	"math/rand"
	"time"
)

var (
	addr = "root:123456@tcp(192.168.99.100:3306)/hello"
	rabbitmqAddr = "amqp://guest:guest@192.168.0.239:5672"
)

func main() {
	rabbitmq_delay.InitMysqlDB(addr)
	//rabbitmq_delay.InitRabbitmq(rabbitmqAddr)
	mqPool := rabbitmq_delay.NewPools(rabbitmqAddr, 10)
	rabbitmq_delay.InitRabbitmqChannel(rabbitmqAddr)

	rand.Seed(time.Now().UnixNano())
	go rabbitmq_delay.NewConsume(mqPool, "1")
	go rabbitmq_delay.NewConsume(mqPool, "2")

	for i := 0; i < 1000; i++ {
		n := rand.Int63n(10)
		rabbitmq_delay.PublishDelayMsg(mqPool, "helloworld", i, n*1000)
	}

	for {

	}
}
