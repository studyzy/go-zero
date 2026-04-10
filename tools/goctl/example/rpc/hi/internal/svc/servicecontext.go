package svc

import "github.com/studyzy/go-zero/tools/goctl/example/rpc/hi/internal/config"

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
