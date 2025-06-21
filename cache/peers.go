package cache

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, exists bool)
}

type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
