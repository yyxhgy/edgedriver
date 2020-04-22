package driver

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

type Rabbitmq struct{}

func (m *Rabbitmq) Response(msg string) error {
	err := resp(hid + ":" + msg)
	return err
}
func (m *Rabbitmq) Init(url, id, send, ver, username, password string, callback func(msg string)) error {
	hid = id
	path = url
	sendTopic = send
	version = ver
	authUserName = username
	authPassword = password
	handleMsg = callback
	err := connectInitial()
	createQueue(hid)
	worker(sendTopic)
	return err
}
func (m *Rabbitmq) RegisterEvent(callback func(msg string)) {
	handleMsg = callback
}

var Connection *amqp.Connection

func connectInitial() error {
	var err error
	if Connection != nil && !Connection.IsClosed() {
		return err
	}
	Connection, err = amqp.Dial(path)
	if err != nil {
		return err
	}
	reconnectRegister()
	return err
}
func reconnectRegister() {
	go func() {
		er := make(chan *amqp.Error)
		Connection.NotifyClose(er)
		<-er
		connectInitial()
		worker(hid)
	}()
}
func createQueue(queueName string) error {
	var err error
	connectInitial()
	ch, err := Connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	_, _ = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	return err
}
func resp(msg string) error {
	var err error
	connectInitial()
	ch, err := Connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		sendTopic, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(hid + ":\n" + msg),
		})
	if err != nil {
		return err
	}
	return err
}

func worker(queueName string) error {
	var err error
	connectInitial()
	ch, err := Connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return err
	}
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Println(string(d.Body))
			handleMsg(string(d.Body))
			time.Sleep(1000)
			d.Ack(false)
		}
	}()
	<-forever

	return err
}
