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
				logger.Errorw("storage error", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			com, err := commonHttp.NewCommunicator(routeStart.GetEndpoints())
			if err != nil {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			stream, err := com.Stream(r.Context())
			if err != nil {
				logger.Errorw("did not get stream", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = sendRequest(stream, r, routeStart)
			if err != nil {
				logger.Errorw("error while streaming request", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = writeResponse(stream, w)
			if err != nil {
				logger.Errorw("error while streaming response", "error", err)
			}
		} else {
			logger.Error("storage not ready")
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}
