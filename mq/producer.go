package mq

import (
	"filestore_server/config"
	"log"

	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var channel *amqp.Channel

// 监听异常关闭
var notifyClose chan *amqp.Error

func init() {
	// 是否开启异步转移，开启时才初始化rabbitmq连接
	if !config.AsyncTransferEnable {
		return
	}
	if initChannel() {
		channel.NotifyClose(notifyClose)
	}
	// 断线重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Printf("onNotifyChannelClosed: %+v\n", msg)
				initChannel()
			}
		}
	}()
}

func initChannel() bool {
	// 判断channel是否已创建
	if channel != nil {
		return true
	}

	// 获取rabbitmq的连接
	conn, err :=amqp.Dial(config.RabbitURL)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 打开一个channel，用于消息的发布于接收
	channel, err = conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}

func Publish(exchange string, routingkey string, msg []byte) bool {
	// 判断channel是否正常
	if !initChannel() {
		return false
	}

	// 执行消息发布动作
	err := channel.Publish(exchange, routingkey, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body: msg,
	})
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}