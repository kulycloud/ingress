package servers

import (
	"context"
	"errors"
	"fmt"
	"github.com/kulycloud/common/logging"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var logger = logging.GetForComponent("servers")

var portsToListenOn = []int32{
	8080,
}
var servers = []*http.Server{}

func Dispatch() {
	registerHandler()

	ctx := getContext()

	for _, port := range portsToListenOn {
		dispatchServer(port)
	}

	<-ctx.Done()
	shutdownServers()
}

func getContext() context.Context {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-stop
		cancel()
	}()

	return ctx
}

func dispatchServer(port int32) {
	server := &http.Server{
		Addr: fmt.Sprintf(":%v", port),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Panicw("error starting server", "error", err, "port", server.Addr)
		}
	}()
	logger.Infow("started server", "port", server.Addr)
	servers = append(servers, server)
}

func shutdownServers() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, server := range servers {
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("error stopping server", "error", err, "port", server.Addr)
		}
	}
}
