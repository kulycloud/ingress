package main

import (
	"github.com/kulycloud/common/logging"
	"github.com/kulycloud/ingress/communication"
	"github.com/kulycloud/ingress/config"
	"github.com/kulycloud/ingress/servers"
)

var logger = logging.GetForComponent("init")

func main() {
	defer logging.Sync()

	err := config.ParseConfig()
	if err != nil {
		logger.Fatalw("Error parsing config", "error", err)
	}
	logger.Infow("Finished parsing config", "config", config.GlobalConfig)

	communication.RegisterToControlPlane()

	servers.Dispatch()
}
