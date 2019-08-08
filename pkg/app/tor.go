package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/wybiral/torgo"
	"github.com/wybiral/tube/pkg/onionkey"
)

type tor struct {
	OnionKey   onionkey.Key
	Controller *torgo.Controller
}

func newTor(ct *TorConfig) (*tor, error) {
	addr := fmt.Sprintf("%s:%d", ct.Controller.Host, ct.Controller.Port)
	ctrl, err := torgo.NewController(addr)
	if err != nil {
		return nil, errors.New("unable to connect to Tor controller")
	}
	if len(ct.Controller.Password) > 0 {
		err = ctrl.AuthenticatePassword(ct.Controller.Password)
	} else {
		err = ctrl.AuthenticateCookie()
		if err != nil {
			err = ctrl.AuthenticateNone()
		}
	}
	if err != nil {
		return nil, errors.New("unable to authenticate to Tor controller")
	}
	key, err := onionkey.ReadFile("onion.key")
	if os.IsNotExist(err) {
		key = nil
	} else if err != nil {
		return nil, err
	}
	t := &tor{
		Controller: ctrl,
		OnionKey:   key,
	}
	return t, nil
}
