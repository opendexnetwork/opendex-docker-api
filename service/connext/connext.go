package connext

import (
	"context"
	"encoding/json"
	"errors"
	docker "github.com/docker/docker/client"
	"github.com/opendexnetwork/opendex-docker-api/config"
	"github.com/opendexnetwork/opendex-docker-api/service/core"
	"github.com/opendexnetwork/opendex-docker-api/service/opendexd"
)

type Service struct {
	*core.SingleContainerService
	*RpcClient
}

func New(
	name string,
	services map[string]core.Service,
	containerName string,
	dockerClient *docker.Client,
	rpcConfig config.RpcConfig,
) *Service {
	base := core.NewSingleContainerService(name, services, containerName, dockerClient)

	return &Service{
		SingleContainerService: base,
		RpcClient:              NewRpcClient(rpcConfig, base),
	}
}

func (t *Service) GetStatus(ctx context.Context) string {
	status := t.SingleContainerService.GetStatus(ctx)
	if status == "Disabled" {
		return status
	}
	if status != "Container running" {
		return status
	}

	// container is running

	svc := t.GetService("opendexd")
	if svc != nil {
		opendexdSvc := svc.(*opendexd.Service)
		info, err := opendexdSvc.GetInfo(ctx)
		if err == nil {
			return info.Connext.Status
		}
	}

	if t.IsHealthy(ctx) {
		return "Ready"
	} else {
		return "Starting..."
	}
}

func (t *Service) GetEthProvider() (string, error) {
	value, err := t.Getenv("VECTOR_CONFIG")
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", errors.New("VECTOR_CONFIG not found")
	}
	var cfg struct {
		ChainProviders map[string]string `json:"chainProviders"`
	}
	if err := json.Unmarshal([]byte(value), &cfg); err != nil {
		return "", err
	}
	// FIXME don't use "4" in mainnet, better remove this magic number
	return cfg.ChainProviders["4"], nil
}

func (t *Service) Close() error {
	err := t.RpcClient.Close()
	if err != nil {
		return err
	}
	return nil
}
