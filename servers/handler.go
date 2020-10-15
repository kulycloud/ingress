package servers

import (
	"net/http"
)

func registerHandler() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// parse request to grpc
		// get route step and send foreward
		// somehow write back response

		// for now:
		w.Write([]byte("Ping"))
	})
}
