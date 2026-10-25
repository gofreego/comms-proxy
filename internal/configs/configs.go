package configs

import (
	"context"
	"fmt"

	repo "github.com/gofreego/comms-proxy/internal/repository"
	"github.com/gofreego/comms-proxy/internal/service"

	"github.com/gofreego/goutils/api"
	"github.com/gofreego/goutils/api/debug"
	"github.com/gofreego/goutils/configutils"
	"github.com/gofreego/goutils/logger"
)

type Configuration struct {
	LogConfig    bool               `yaml:"LogConfig"`
	Logger       logger.Config      `yaml:"Logger"`
	ConfigReader configutils.Config `yaml:"ConfigReader"`
	AppNames     []string           `yaml:"AppNames"`
	Server       Server             `yaml:"Server" `
	Repository   repo.Config        `yaml:"Repository"`
	Service      service.Config     `yaml:"Service"`
	Debug        debug.Config       `yaml:"Debug"`
}

type Server struct {
	GRPCPort int             `yaml:"GRPCPort"`
	HTTPPort int             `yaml:"HTTPPort"`
	CORS     *api.CORSConfig `yaml:"CORS"`
}

func (c *Server) GetCORSConfig() *api.CORSConfig {
	return c.CORS
}

func LoadConfig(ctx context.Context, path string, env string) *Configuration {
	filePath := fmt.Sprintf("%s/%s.yaml", path, env)
	var conf Configuration
	err := configutils.ReadConfig(ctx, filePath, &conf)
	if err != nil {
		logger.Panic(ctx, "failed to read configs : %v", err)
	}
	if conf.LogConfig {
		configutils.LogConfig(ctx, conf)
	}
	return &conf
}
