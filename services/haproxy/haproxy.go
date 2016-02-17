package haproxy

import (
	"runtime"

	"github.com/QubitProducts/bamboo/Godeps/_workspace/src/github.com/samuel/go-zookeeper/zk"
	conf "github.com/QubitProducts/bamboo/configuration"
	"github.com/QubitProducts/bamboo/services/marathon"
	"github.com/QubitProducts/bamboo/services/service"
)

type templateData struct {
	Apps     marathon.AppList
	Services map[string]service.Service
	NBProc   int
}

func GetTemplateData(config *conf.Configuration, conn *zk.Conn) (*templateData, error) {

	apps, err := marathon.FetchApps(config.Marathon, config)

	if err != nil {
		return nil, err
	}

	services, err := service.All(conn, config.Bamboo.Zookeeper)

	if err != nil {
		return nil, err
	}

	cores := runtime.NumCPU()
	if cores > 64 {
		return &templateData{apps, services, 64}, nil
	}

	return &templateData{apps, services, cores}, nil
}
