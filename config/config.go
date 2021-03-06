package config

import (
	commonConfig "github.com/kulycloud/common/config"
)

type Config struct {
	Host             string   `configName:"host"`
	Port             uint32   `configName:"port"`
	ControlPlaneHost string   `configName:"controlPlaneHost"`
	ControlPlanePort uint32   `configName:"controlPlanePort"`
	CertFile         string   `configName:"certFile"`
	KeyFile          string   `configName:"keyFile"`
	HTTPPorts        []string `configName:"httpPorts" defaultValue:"443"`
}

var GlobalConfig = &Config{}

func ParseConfig() error {
	parser := commonConfig.NewParser()
	parser.AddProvider(commonConfig.NewCliParamProvider())
	parser.AddProvider(commonConfig.NewEnvironmentVariableProvider())

	return parser.Populate(GlobalConfig)
}
