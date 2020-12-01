package servers

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/kulycloud/common/logging"
	"github.com/kulycloud/ingress/config"
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

var tlsCfg = &tls.Config{
	MinVersion:               tls.VersionTLS12,
	CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
	PreferServerCipherSuites: true,
	CipherSuites: []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	},
}

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
		Addr:         fmt.Sprintf(":%v", port),
		TLSConfig:    tlsCfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	go func() {
		err := server.ListenAndServeTLS(config.GlobalConfig.CertFile, config.GlobalConfig.KeyFile)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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
