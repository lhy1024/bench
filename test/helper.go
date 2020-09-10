package test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/lhy1024/bench/bench"
	"github.com/pingcap/log"
	"github.com/unrolled/render"
	"go.uber.org/zap"
)

type handler struct {
	r         *render.Render
	reports   []bench.WorkloadReport
	resources []bench.ResourceRequestItem
}

func newHandler() *handler {
	return &handler{
		r: render.New(render.Options{
			IndentJSON: true,
		}),
	}
}

func (h *handler) handleResource(w http.ResponseWriter, r *http.Request) {
	h.r.JSON(w, http.StatusOK, h.resources)
}

func handleScaleOut(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) getResults(w http.ResponseWriter, r *http.Request) {
	if len(h.reports) == 0 {
		h.r.JSON(w, http.StatusOK, h.reports)
	} else {
		h.r.JSON(w, http.StatusOK, h.reports[:1])
	}
}

func (h *handler) postResults(w http.ResponseWriter, r *http.Request) {
	var report bench.WorkloadReport
	err := readJson(r.Body, &report)
	if err != nil {
		h.r.JSON(w, http.StatusBadGateway, err.Error())
	} else {
		h.reports = append(h.reports, report)
		h.r.JSON(w, http.StatusOK, "")
	}
}

const (
	mockServerAddr = "127.0.0.1:21212"
)

func mockServer() {
	r := mux.NewRouter().PathPrefix("/api/cluster").Subrouter()
	h := newHandler()
	resource := bench.ResourceRequestItem{}
	resource.ID = 1
	h.resources = append(h.resources, resource)

	r.HandleFunc("/resource/{cluster}", h.handleResource).Methods("GET", "POST")
	r.HandleFunc("/scale_out/{cluster}/{id}/{component}", handleScaleOut).Methods("POST")
	r.HandleFunc("/workload/{cluster}/result", h.postResults).Methods("POST")
	r.HandleFunc("/workload/{cluster}/result", h.getResults).Methods("GET")

	srv := &http.Server{
		Addr:         mockServerAddr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error("start mock server meets error", zap.Error(err))
	}
}

func readJson(r io.ReadCloser, data interface{}) error {
	var err error
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, data)
}
