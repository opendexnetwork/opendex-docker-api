package arby

import (
	"github.com/opendexnetwork/opendex-docker-api/config"
)

type RpcClient struct {
}

func NewRpcClient(config config.RpcConfig) *RpcClient {
	return &RpcClient{}
}

func (t *RpcClient) Close() error {
	return nil
}
