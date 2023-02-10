package common

type StoreType int

const (
	_ StoreType = iota
	// 本地
	StoreLocal
	// ceph集群
	StoreCeph
	// 阿里OSS
	StoreOSS
	// 混合存储（ceph及OSS）
	StoreMix
	// 所有类型都存储一份
	StoreAll
)