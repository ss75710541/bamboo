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
	ErrBadBody = errors.New("Bad weight bytes")
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
		path: fmt.Sprintf("%s/%s", conf.Path, "weights"),
		acl:  defaultACL(),
	}
	err = s.ensurePathExists()
	return s, err
}

func (z *ZKStorage) All() (weights []Weight, err error) {
	err = z.ensurePathExists()
	if err != nil {
		return
	}

	keys, _, err := z.conn.Children(z.path)
	if err != nil {
		return
	}

	weights = make([]Weight, 0, len(keys))
	for _, childPath := range keys {
		body, _, err := z.conn.Get(z.path + "/" + childPath)
		if err != nil {
			return nil, err
		}

		path, err := unescapePath(childPath)
		if err != nil {
			return nil, err
		}

		weight, err := parseWeight(body, path)
		if err != nil {
			log.Printf("Failed to parse weight at %v: %v", path, err)
			continue
		}

		weights = append(weights, weight)
	}

	return
}

func (z *ZKStorage) Upsert(weight Weight) (err error) {
	body, err := encodeWeight(weight)

	if err != nil {
		return
	}

	err = z.ensurePathExists()
	if err != nil {
		return err
	}

	path := z.weightPath(weight.ID)

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

func (z *ZKStorage) Delete(id string) error {
	path := z.weightPath(id)
	log.Println("path", path)
	return z.conn.Delete(path, -1)
}

func (z *ZKStorage) weightPath(id string) string {
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

func parseWeight(body []byte, path string) (weight Weight, err error) {
	err = json.Unmarshal(body, &weight)
	if err != nil {
		return weight, ErrBadBody
	}
	weight.ID = path

	return weight, nil
}

func encodeWeight(weight Weight) ([]byte, error) {
	return json.Marshal(weight)
}
