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
	weights, err := w.Storage.All()
	if err != nil {
		responseError(rw, err.Error())
		return
	}

	byId := make(map[string]application.Weight, len(weights))
	for _, weight := range weights {
		byId[weight.ID] = weight
	}

	responseJSON(rw, byId)
}

func (w *WeightAPI) Put(rw http.ResponseWriter, r *http.Request) {
	weight, err := parseBody(r)
	if err != nil {
		responseError(rw, err.Error())
		return
	}

	err = w.Storage.Upsert(weight)
	if err != nil {
		responseError(rw, err.Error())
		return
	}

	responseJSON(rw, weight)
}

func (w *WeightAPI) Delete(params martini.Params, rw http.ResponseWriter, r *http.Request) {
	id := params["id"]
	err := w.Storage.Delete(id)
	if err != nil {
		responseError(rw, err.Error())
		return
	}

	responseJSON(rw, new(map[string]string))
}

func parseBody(r *http.Request) (application.Weight, error) {
	var weight application.Weight
	payload, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal(payload, &weight)
	if err != nil {
		return weight, err
	}

	return weight, nil
}
