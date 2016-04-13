package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/QubitProducts/bamboo/Godeps/_workspace/src/github.com/samuel/go-zookeeper/zk"
	conf "github.com/QubitProducts/bamboo/configuration"
)

var (
	//ErrBadBody invalid bytes
	ErrBadBody = errors.New("Bad application bytes")
)

type ZKStorage struct {
	conn *zk.Conn
	conf conf.Zookeeper
	path string
	acl  []zk.ACL
}

func NewZKStorage(conn *zk.Conn, conf conf.Zookeeper) (s *ZKStorage, err error) {
	s = &ZKStorage{
		conn: conn,
		conf: conf,
		path: fmt.Sprintf("%s/%s", conf.Path, "applications"),
		acl:  defaultACL(),
	}
	err = s.ensurePathExists()
	return s, err
}

func (z *ZKStorage) All() (applications []Application, err error) {
	err = z.ensurePathExists()
	if err != nil {
		return
	}

	keys, _, err := z.conn.Children(z.path)
	if err != nil {
		return
	}

	applications = make([]Application, 0, len(keys))
	for _, childPath := range keys {
		body, _, err := z.conn.Get(z.path + "/" + childPath)
		if err != nil {
			return nil, err
		}

		path, err := unescapePath(childPath)
		if err != nil {
			return nil, err
		}

		app, err := parseApplication(body, path)
		if err != nil {
			log.Printf("Failed to parse application at %v: %v", path, err)
			continue
		}

		applications = append(applications, app)
	}

	return
}

func (z *ZKStorage) Upsert(app Application) (err error) {
	body, err := encodeApplication(app)

	if err != nil {
		return
	}

	err = z.ensurePathExists()
	if err != nil {
		return err
	}

	path := z.applicationPath(app.ID)

	ok, _, err := z.conn.Exists(path)
	if err != nil {
		return
	}

	if ok {
		_, err = z.conn.Set(path, body, -1)
		if err != nil {
			log.Print("Failed to set path", err)
			return
		}

		// Trigger an event on the parent
		_, err = z.conn.Set(z.path, []byte{}, -1)
		if err != nil {
			log.Print("Failed to trigger event on parent", err)
			err = nil
		}

	} else {
		_, err = z.conn.Create(path, body, 0, z.acl)
		if err != nil {
			log.Print("Failed to set create", err)
			return
		}
	}
	return
}

func (z *ZKStorage) Delete(appId string) error {
	path := z.applicationPath(appId)
	log.Println("path", path)
	return z.conn.Delete(path, -1)
}

func (z *ZKStorage) applicationPath(id string) string {
	return z.path + "/" + escapePath(id)
}

func (z *ZKStorage) ensurePathExists() error {
	pathExists, _, _ := z.conn.Exists(z.path)
	if pathExists {
		return nil
	}

	// This is a fairly rare, and fairly critical, operation, so I'm going to be verbose
	log.Print("Creating base zk path", z.path)
	_, err := z.conn.Create(z.path, []byte{}, 0, z.acl)
	if err != nil {
		log.Print("Failed to create base zk path", err)
	}

	return err
}

func escapePath(path string) string {
	return url.QueryEscape(path)
}

func unescapePath(path string) (string, error) {
	return url.QueryUnescape(path)
}

func defaultACL() []zk.ACL {
	return []zk.ACL{zk.ACL{Perms: zk.PermAll, Scheme: "world", ID: "anyone"}}
}

func parseApplication(body []byte, path string) (app Application, err error) {
	err = json.Unmarshal(body, &app)
	if err != nil {
		return app, ErrBadBody
	}
	app.ID = path

	return app, nil
}

func encodeApplication(app Application) ([]byte, error) {
	return json.Marshal(app)
}
