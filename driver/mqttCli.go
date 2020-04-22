package driver

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

type Mqtt struct{}

func (m *Mqtt) Response(msg string) error {
	publishMqtt(sendTopic, hid+":"+msg, 0, true)
	return nil
}
func (m *Mqtt) Init(url, id, send, ver, username, password string, callback func(msg string)) error {
	hid = id
	path = url
	sendTopic = send
	version = ver
	authUserName = username
	authPassword = password
	handleMsg = callback
	initMqtt()
	return nil
}
func (m *Mqtt) RegisterEvent(callback func(msg string)) {
	handleMsg = callback
}

var Client mqtt.Client

func initMqtt() {
	opts := mqtt.NewClientOptions().AddBroker(path)
	opts.SetClientID(hid)
	opts.SetUsername(authUserName)
	opts.SetPassword(authPassword)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, message mqtt.Message) {
		fmt.Printf("TOPIC: %s\n", message.Topic())
		fmt.Printf("MSG: %s\n", message.Payload())
		handleMsg(string(message.Payload()))
	})
	opts.SetPingTimeout(1 * time.Second)
	opts.SetCleanSession(false)
	opts.SetConnectRetry(true)
	opts.SetReconnectingHandler(func(client mqtt.Client, options *mqtt.ClientOptions) {
		fmt.Println("重连中...")
	})
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		publishMqtt(sendTopic, hid+":在线，当前版本【"+version+"】", 0, true)
		subscribeMqtt(hid, 0)
	})
	Client = mqtt.NewClient(opts)
	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		panic(token.Error())
	}
}
func subscribeMqtt(topic string, qos byte) error {
	if token := Client.Subscribe(topic, qos, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return token.Error()
	}
	return nil
}
func publishMqtt(topic, playload string, qos byte, retained bool) error {
	if token := Client.Publish(topic, qos, retained, playload); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return token.Error()
	}
	return nil
}
func subscribeCancelMqtt(topic string) error {
	if token := Client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return token.Error()
	}
	return nil
}
func closeMqtt() {
	Client.Disconnect(250)
}
