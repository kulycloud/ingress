package communication

import (
	commonCommunication "github.com/kulycloud/common/communication"
	"github.com/kulycloud/common/logging"
	"github.com/kulycloud/ingress/config"
)

var logger = logging.GetForComponent("communication")

var Storage *commonCommunication.StorageCommunicator

func RegisterToControlPlane() {
	communicator := commonCommunication.RegisterToControlPlane("ingress",
		config.GlobalConfig.Host, config.GlobalConfig.Port,
		config.GlobalConfig.ControlPlaneHost, config.GlobalConfig.ControlPlanePort)

	logger.Info("Starting listener")
	listener := commonCommunication.NewListener(logging.GetForComponent("listener"))
	if err := listener.Setup(config.GlobalConfig.Port); err != nil {
		logger.Panicw("error initializing listener", "error", err)
	}

	handler := NewIngressHandler()
	handler.Register(listener)
	go func() {
		if err := <-listener.Serve(); err != nil {
			logger.Panicw("error serving listener", "error", err)
		}
	}()

	Storage = listener.Storage
	<-communicator
}
