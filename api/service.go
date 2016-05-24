package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/QubitProducts/bamboo/Godeps/_workspace/src/github.com/go-martini/martini"
	conf "github.com/QubitProducts/bamboo/configuration"
	"github.com/QubitProducts/bamboo/services/service"
)

const (
	HttpDefaultTimeout = time.Second * 15
)

var (
	MarathonUser      string
	MarathonPassword  string
	MarathonEndpoints []string
	BambooEndpoint    string
)

type ServiceAPI struct {
	Config  *conf.Configuration
	Storage service.Storage
}

// marathon call back urls
type MarathonCallback struct {
	CallbackUrls []string
}

func LoadConfig(config conf.Configuration) {
	MarathonUser = config.Marathon.User
	MarathonPassword = config.Marathon.Password
	MarathonEndpoints = config.Marathon.Endpoints()
	BambooEndpoint = config.Bamboo.Endpoint + "/api/marathon/event_callback"
}

func (d *ServiceAPI) All(w http.ResponseWriter, r *http.Request) {
	services, err := d.Storage.All()

	if err != nil {
		responseError(w, err.Error())
		return
	}

	byId := make(map[string]service.Service, len(services))
	for _, s := range services {
		byId[s.Id] = s
	}

	responseJSON(w, byId)
}

func (d *ServiceAPI) Create(w http.ResponseWriter, r *http.Request) {
	service, err := extractService(r)

	if err != nil {
		responseError(w, err.Error())
		return
	}

	err = d.Storage.Upsert(service)
	if err != nil {
		responseError(w, err.Error())
		return
	}

	responseJSON(w, service)
}

func (d *ServiceAPI) Put(params martini.Params, w http.ResponseWriter, r *http.Request) {
	service, err := extractService(r)
	if err != nil {
		responseError(w, err.Error())
		return
	}

	err = d.Storage.Upsert(service)
	if err != nil {
		responseError(w, err.Error())
		return
	}

	responseJSON(w, service)
}

func (d *ServiceAPI) Delete(params martini.Params, w http.ResponseWriter, r *http.Request) {
	serviceId := params["_1"]
	err := d.Storage.Delete(serviceId)
	if err != nil {
		responseError(w, err.Error())
		return
	}

	responseJSON(w, new(map[string]string))
}

func extractService(r *http.Request) (service.Service, error) {
	var serviceModel service.Service
	payload, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal(payload, &serviceModel)
	if err != nil {
		return serviceModel, errors.New("Unable to decode JSON request")
	}
	if !strings.HasPrefix(serviceModel.Id, "/") {
		serviceModel.Id = "/" + serviceModel.Id
	}

	return serviceModel, nil
}

func responseError(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusBadRequest)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	for _, marathonEnp := range MarathonEndpoints {
		if registered := checkMarathonCallback(marathonEnp); !registered {
			if err := registerMarathonEvent(marathonEnp); err != nil {
				http.Error(w, "healthcheck failed", http.StatusInternalServerError)
				return
			}
		}
	}
	io.WriteString(w, "healthcheck success")
}

func responseJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	bites, _ := json.Marshal(data)
	w.Write(bites)
}

// register singla marathon event
func registerMarathonEvent(marathonEnp string) error {
	client := &http.Client{}
	url := marathonEnp + "/v2/eventSubscriptions?callbackUrl=" + BambooEndpoint
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	if len(MarathonUser) > 0 && len(MarathonPassword) > 0 {
		req.SetBasicAuth(MarathonUser, MarathonPassword)
	}
	resp, err := client.Do(req)
	if err != nil {
		errorMsg := "An error occurred while accessing Marathon callback system: %s\n"
		log.Printf(errorMsg, err)
		return err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
		return err
	}
	body := string(bodyBytes)
	if strings.HasPrefix(body, "{\"message") {
		warningMsg := "Access to the callback system of Marathon seems to be failed, response: %s\n"
		log.Printf(warningMsg, body)
		return fmt.Errorf(warningMsg, body)
	}

	return nil
}

// check marathon eventSubscriptions whether container bamboo endpoint
func checkMarathonCallback(marathonEnp string) bool {
	client := &http.Client{
		Timeout: HttpDefaultTimeout,
	}

	url := marathonEnp + "/v2/eventSubscriptions"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Check marathon eventSubscriptions got error: ", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Check marathon eventSubscriptions got error: ", err)
		return false
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Check marathon eventSubscriptions read resp body got error: ", err)
		return false
	}

	var marathonCallbak MarathonCallback
	if err := json.Unmarshal(body, &marathonCallbak); err != nil {
		log.Println("Unmarshal marathon callback got error: ", err)
		return false
	}

	for _, callbackUrl := range marathonCallbak.CallbackUrls {
		if callbackUrl == BambooEndpoint {
			return true
		}
	}

	return false
}
