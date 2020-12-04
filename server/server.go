package server

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/segmentio/ksuid"
	"github.com/spaceuptech/helpers"

	"github.com/spaceuptech/sc-config-sync/admin"
)

// Server holds server module information
type Server struct {
	Admin *admin.Module
}

// New initializes server module
func New(secret string) *Server {
	return &Server{Admin: admin.New(secret)}
}

// Start starts server
func (s *Server) Start() error {
	// Start the server
	helpers.Logger.LogInfo("", "Starting space cloud config sync server on port 12000", nil)
	if err := http.ListenAndServe(":12000", loggerMiddleWare(s.routes())); err != nil {
		log.Fatal("Error while starting server:", err)
	}

	return nil
}

func (s *Server) routes() *mux.Router {
	// Set up the routes
	r := mux.NewRouter()
	r.Methods(http.MethodPost).Path("/db/sync").Handler(HandleDatabaseSync(s.Admin))
	return r
}

func loggerMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := r.Header.Get(helpers.HeaderRequestID)
		if requestID == "" {
			// set a new request id header of request
			requestID = ksuid.New().String()
			r.Header.Set(helpers.HeaderRequestID, requestID)
		}

		var reqBody []byte
		if r.Header.Get("Content-Type") == "application/json" {
			reqBody, _ = ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
		}

		helpers.Logger.LogInfo(requestID, "Request", map[string]interface{}{"method": r.Method, "url": r.URL.Path, "queryVars": r.URL.Query(), "body": string(reqBody)})
		next.ServeHTTP(w, r.WithContext(helpers.CreateContext(r)))

	})
}
