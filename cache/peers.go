package cache

import pb "distributed-cache/cache/pb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerClient, exists bool)
}

type PeerClient interface {
	// Get(group string, key string) ([]byte, error)
	Get(in *pb.GetRequest, out *pb.Response) error
}
