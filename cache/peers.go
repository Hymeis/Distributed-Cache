package cache

import pb "distributed-cache/cache/pb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, exists bool)
}

type PeerGetter interface {
	// Get(group string, key string) ([]byte, error)
	Get(in *pb.Request, out *pb.Response) error
}
