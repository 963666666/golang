package rabbitmq_delay

import (
	"github.com/streadway/amqp"
	"sync"
)

func NewPools(addr string, maxconn int) *MQPool {
	return &MQPool{
		MaxConnect: maxconn,
		Addr:       addr,
		mutex:      sync.Mutex{},
		connects:   make([]*amqp.Connection, maxconn),
		index:      0,
	}
}

type MQPool struct {
	MaxConnect int
	Addr       string
	connects   []*amqp.Connection
	index      int // 链接游标， 目的是并发使用connects
	mutex      sync.Mutex
}

func (mp *MQPool) Init() {

}
func (mp *MQPool) Connect() (conn *amqp.Connection, err error) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.index += 1
	index := mp.index % mp.MaxConnect
	conn = mp.connects[index]
	if conn == nil || conn.IsClosed() {
		conn, err = amqp.Dial(mp.Addr)
		if err != nil {
			panic(err)
		}
		mp.connects[index] = conn
	}
	return
}

func (mp *MQPool) Channel() (*amqp.Channel, error) {
	conn, err := mp.Connect()
	if err != nil {
		return nil, err
	}
	return conn.Channel()
}