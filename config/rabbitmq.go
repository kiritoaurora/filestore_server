package config

const (
	// 是否开启文件异步转移（默认同步）
	AsyncTransferEnable = true
	
	// rabbitmq服务的入口url
	RabbitURL = "amqp://guest:guest@127.0.0.1:5672/"

	// 用于文件transfer的交换机
	TransExchangeName = "uploadserver.trans"

	// ceph转移队列名
	TransCephQueueName = "uploadserver.trans.ceph"

	// ceph转移失败后写入另一个队列的队列名
	TransCephErrQueueName = "uploadserver.trans.ceph.err"

	// routingkey
	TransCephRoutingKey = "ceph"
)