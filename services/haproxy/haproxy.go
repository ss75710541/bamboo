package haproxy

import (
	"runtime"

	conf "github.com/QubitProducts/bamboo/configuration"
	"github.com/QubitProducts/bamboo/services/marathon"
	"github.com/QubitProducts/bamboo/services/service"
)

type templateData struct {
	Apps     marathon.AppList
	Services map[string]service.Service
	NBProc   int
}

func GetTemplateData(config *conf.Configuration, storage service.Storage) (*templateData, error) {

	apps, err := marathon.FetchApps(config.Marathon, config)

	if err != nil {
		return nil, err
	}

	services, err := storage.All()
	if err != nil {
		return nil, err
	}

	byName := make(map[string]service.Service)
	for _, service := range services {
		byName[service.Id] = service
	}

	cores := runtime.NumCPU()
	if cores > 64 {
		return &templateData{apps, byName, 64}, nil
	} else {
		return &templateData{apps, byName, cores}, nil
	}
}
