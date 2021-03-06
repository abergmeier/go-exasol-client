/*
    AUTHOR

	Grant Street Group <developers@grantstreet.com>

	COPYRIGHT AND LICENSE

	This software is Copyright (c) 2019 by Grant Street Group.
	This is free software, licensed under:
	    MIT License
*/

package exasol

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"

	"github.com/gorilla/websocket"
)

type WSConn interface {
	Close() error
	EnableWriteCompression(bool)
	WriteJSON(interface{}) error
	ReadJSON(interface{}) error
}

var (
	defaultDialer = *websocket.DefaultDialer
)

func init() {
	defaultDialer.Proxy = nil // TODO use proxy env
	defaultDialer.EnableCompression = false
}

func WSConnect(conf ConnConf, log Logger) (WSConn, error) {
	uri := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	u := url.URL{
		Scheme: "ws",
		Host:   uri,
	}
	log.Debugf("Connecting to %s", u.String())
	// According to documentation:
	// > It is safe to call Dialer's methods concurrently.
	ws, resp, err := defaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Debugf("resp:%s", resp)
		return nil, err
	}
	return ws, nil
}

// Request and Response are pointers to structs representing the API JSON.
// The Response struct is updated in-place.

func (c *Conn) send(request, response interface{}) error {
	receiver, err := c.asyncSend(request)
	if err != nil {
		return err
	}
	return receiver(response)
}

func (c *Conn) asyncSend(request interface{}) (func(interface{}) error, error) {
	err := c.ws.WriteJSON(request)
	if err != nil {
		return nil, c.error("WebSocket API Error sending: %s", err)
	}

	return func(response interface{}) error {
		err = c.ws.ReadJSON(response)
		if err != nil {
			if regexp.MustCompile(`abnormal closure`).
				MatchString(err.Error()) {
				return fmt.Errorf("Server terminated statement")
			}
			return fmt.Errorf("WebSocket API Error recving: %s", err)
		}
		r := reflect.Indirect(reflect.ValueOf(response))
		status := r.FieldByName("Status").String()
		if status != "ok" {
			err := reflect.Indirect(r.FieldByName("Exception")).
				FieldByName("Text").String()
			return fmt.Errorf("Server Error: %s", err)
		}
		return nil
	}, nil
}
