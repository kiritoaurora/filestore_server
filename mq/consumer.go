package mq

import "log"

var done chan bool

// 开始监听队列，获取消息
func StartConsume(qName string, cName string, callback func(msg []byte) bool) {
	// 通过channel.Consume获得消息信道
	msgs, err := channel.Consume(qName, cName, true, false, false, false, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// 循环获取队列的消息
	done = make(chan bool)

	go func() {
		for msg := range msgs {
			// 调用callback处理新的消息
			processSuc := callback(msg.Body)
			if !processSuc {
				// TODO:将任务写到另一个队列，用于异常情况的重试
				log.Println("处理异常")
			}
		} 
	}()
	
	// 阻塞
	<-done

	// 关闭Rabbitmq的Channel
	channel.Close()
}