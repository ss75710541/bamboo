package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/QubitProducts/bamboo/Godeps/_workspace/src/github.com/go-martini/martini"
	"github.com/QubitProducts/bamboo/configuration"
	"github.com/QubitProducts/bamboo/services/application"
)

var (
	//ErrBadApp invalid application
	ErrBadApp = errors.New("Bad application")
)

type WeightAPI struct {
	Config  *configuration.Configuration
	Storage application.Storage
}

func (w *WeightAPI) All(rw http.ResponseWriter, r *http.Request) {
	applications, err := w.Storage.All()
	if err != nil {
		responseError(rw, err.Error())
		return
	}

	byId := make(map[string]application.Application, len(applications))
	for _, app := range applications {
		byId[app.ID] = app
	}

	responseJSON(rw, byId)
}

func (w *WeightAPI) Create(rw http.ResponseWriter, r *http.Request) {
	app, err := parseApplication(r)
	if err != nil {
		responseError(rw, err.Error())
		return
	}

	err = w.Storage.Upsert(app)
	if err != nil {
		responseError(rw, err.Error())
		return
	}

	responseJSON(rw, app)
}

func (w *WeightAPI) Put(params martini.Params, rw http.ResponseWriter, r *http.Request) {
	app, err := parseApplication(r)
	if err != nil {
		responseError(rw, err.Error())
		return
	}

	err = w.Storage.Upsert(app)
	if err != nil {
		responseError(rw, err.Error())
		return
	}

	responseJSON(rw, app)
}

func (w *WeightAPI) Delete(params martini.Params, rw http.ResponseWriter, r *http.Request) {
	appID := params["app_id"]
	err := w.Storage.Delete(appID)
	if err != nil {
		responseError(rw, err.Error())
		return
	}

	responseJSON(rw, new(map[string]string))
}

func parseApplication(r *http.Request) (application.Application, error) {
	var appModel application.Application
	payload, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal(payload, &appModel)
	if err != nil {
		return appModel, err
	}

	return appModel, nil
}
