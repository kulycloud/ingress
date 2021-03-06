package servers

import (
	commonHttp "github.com/kulycloud/common/http"
	"github.com/kulycloud/ingress/communication"
	"net/http"
)

func registerHandler() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if communication.Storage.Ready() {
			routeStart, err := communication.Storage.GetRouteStart(r.Context(), r.Host)
			if err != nil {
				logger.Warnw("storage error", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			com, err := commonHttp.NewCommunicator(r.Context(), routeStart.Step.Endpoints)
			if err != nil {
				logger.Warnw("communication error", "error", err, "endpoints", routeStart.Step.Endpoints)
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			stream, err := com.Stream(r.Context())
			if err != nil {
				logger.Warnw("did not get stream", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = sendRequest(stream, r, routeStart)
			if err != nil {
				logger.Warnw("error while streaming request", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = writeResponse(stream, w)
			if err != nil {
				logger.Warnw("error while streaming response", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			logger.Error("storage not ready")
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
