package communication

import (
	"context"
	commonCommunication "github.com/kulycloud/common/communication"
	"github.com/kulycloud/common/logging"
	"github.com/kulycloud/ingress/config"
	"time"
)

var logger = logging.GetForComponent("communication")

func RegisterToControlPlane() {
	go registerLoop()

	logger.Info("Starting listener")
	listener := commonCommunication.NewListener(logging.GetForComponent("listener"))
	if err := listener.Setup(config.GlobalConfig.Port); err != nil {
		logger.Panicw("error initializing listener", "error", err)
	}

	handler := NewIngressHandler()
	handler.Register(listener)
	go func() {
		if err := listener.Serve(); err != nil {
			logger.Panicw("error serving listener", "error", err)
		}
	}()
}

func registerLoop() {
	for {
		_, err := register()
		if err == nil {
			break
		}

		logger.Info("Retrying in 5s...")
		time.Sleep(5 * time.Second)
	}
}

func register() (*commonCommunication.ControlPlaneCommunicator, error) {
	comm := commonCommunication.NewControlPlaneCommunicator()
	err := comm.Connect(config.GlobalConfig.ControlPlaneHost, config.GlobalConfig.ControlPlanePort)
	if err != nil {
		logger.Errorw("Could not connect to control-plane", "error", err)
		return nil, err
	}
	err = comm.RegisterThisService(context.Background(), "ingress", config.GlobalConfig.Host, config.GlobalConfig.Port)
	if err != nil {
		logger.Errorw("Could not register service", "error", err)
		return nil, err
	}
	logger.Info("Registered to control-plane")
	return comm, nil
}
