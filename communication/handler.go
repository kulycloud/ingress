package communication

import (
	commonCommunication "github.com/kulycloud/common/communication"
	protoIngress "github.com/kulycloud/protocol/ingress"
)

var _ protoIngress.IngressServer = &IngressHandler{}

type IngressHandler struct{}

func NewIngressHandler() *IngressHandler {
	return &IngressHandler{}
}

func (handler *IngressHandler) Register(listener *commonCommunication.Listener) {
	protoIngress.RegisterIngressServer(listener.Server, handler)
}
