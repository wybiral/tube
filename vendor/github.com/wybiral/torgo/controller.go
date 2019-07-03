package torgo

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/textproto"
	"strconv"
	"strings"
)

// A Controller instance is a control port connection that provides methods for
// communicating with Tor.
type Controller struct {
	// Array of available authentication methods.
	AuthMethods []string
	// Cookie file path (empty if not available).
	CookieFile string
	// Text is a textproto.Conn to the control port.
	Text *textproto.Conn
}

// NewController returns a new Controller instance connecting to the control
// port at addr.
func NewController(addr string) (*Controller, error) {
	text, err := textproto.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	c := &Controller{Text: text}
	err = c.getProtocolInfo()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Make a textproto request and expect the command to have a valid 250 status.
func (c *Controller) makeRequest(request string) (int, string, error) {
	id, err := c.Text.Cmd(request)
	if err != nil {
		return 0, "", err
	}
	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)
	return c.Text.ReadResponse(250)
}

// Perform PROTOCOLINFO command and parse the response.
func (c *Controller) getProtocolInfo() error {
	_, msg, err := c.makeRequest("PROTOCOLINFO 1")
	if err != nil {
		return err
	}
	lines := strings.Split(msg, "\n")
	authPrefix := "AUTH METHODS="
	cookiePrefix := "COOKIEFILE="
	for _, line := range lines {
		// Check for AUTH METHODS line
		if strings.HasPrefix(line, authPrefix) {
			// This line should be in a format like:
			/// AUTH METHODS=method1,method2,methodN COOKIEFILE=cookiefilepath
			line = line[len(authPrefix):]
			parts := strings.SplitN(line, " ", 2)
			c.AuthMethods = strings.Split(parts[0], ",")
			// Check for COOKIEFILE key/value
			if len(parts) == 2 && strings.HasPrefix(parts[1], cookiePrefix) {
				raw := parts[1][len(cookiePrefix):]
				c.CookieFile, err = strconv.Unquote(raw)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Perform GETINFO command based on key.
func (c *Controller) getInfo(key string) (string, error) {
	_, msg, err := c.makeRequest("GETINFO " + key)
	if err != nil {
		return "", err
	}
	lines := strings.Split(msg, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if parts[0] == key {
			return parts[1], nil
		}
	}
	return "", fmt.Errorf(key + " not found")
}

// Perform GETINFO command and convert response to int.
func (c *Controller) getInfoInt(key string) (int, error) {
	s, err := c.getInfo(key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

// GetAddress returns the current external IP address.
func (c *Controller) GetAddress() (string, error) {
	return c.getInfo("address")
}

// GetBytesRead returns total bytes downloaded.
func (c *Controller) GetBytesRead() (int, error) {
	return c.getInfoInt("traffic/read")
}

// GetBytesWritten returns total bytes uploaded.
func (c *Controller) GetBytesWritten() (int, error) {
	return c.getInfoInt("traffic/written")
}

// GetConfigFile return path to Tor config file.
func (c *Controller) GetConfigFile() (string, error) {
	return c.getInfo("config-file")
}

// GetTorPid returns PID for current Tor process.
func (c *Controller) GetTorPid() (int, error) {
	return c.getInfoInt("process/pid")
}

// GetVersion returns version of Tor server.
func (c *Controller) GetVersion() (string, error) {
	return c.getInfo("version")
}

// AuthenticateNone authenticate to controller without password or cookie.
func (c *Controller) AuthenticateNone() error {
	_, _, err := c.makeRequest("AUTHENTICATE")
	if err != nil {
		return err
	}
	return nil
}

// AuthenticatePassword authenticate to controller with password.
func (c *Controller) AuthenticatePassword(password string) error {
	quoted := strconv.Quote(password)
	_, _, err := c.makeRequest("AUTHENTICATE " + quoted)
	if err != nil {
		return err
	}
	return nil
}

// AuthenticateCookie authenticate to controller with cookie from current
// CookieFile path.
func (c *Controller) AuthenticateCookie() error {
	rawCookie, err := ioutil.ReadFile(c.CookieFile)
	if err != nil {
		return err
	}
	cookie := hex.EncodeToString(rawCookie)
	_, _, err = c.makeRequest("AUTHENTICATE " + cookie)
	if err != nil {
		return err
	}
	return nil
}

// AddOnion adds Onion hidden service. If no private key is supplied one will
// be generated and the PrivateKeyType and PrivateKey properties will be set
// with the newly generated one.
// The hidden service will use port mapping contained in Ports map supplied.
// ServiceID will be assigned based on the private key and will be the address
// of this hidden service (without the ".onion" ending).
func (c *Controller) AddOnion(onion *Onion) error {
	if onion == nil {
		return errors.New("torgo: onion cannot be nil")
	}
	if len(onion.Ports) == 0 {
		return errors.New("torgo: onion requires at least one port mapping")
	}
	req := "ADD_ONION "
	// If no key is supplied set PrivateKeyType to NEW
	if len(onion.PrivateKey) == 0 {
		if onion.PrivateKeyType == "" {
			onion.PrivateKey = "BEST"
		} else {
			onion.PrivateKey = onion.PrivateKeyType
		}
		onion.PrivateKeyType = "NEW"
	}
	req += fmt.Sprintf("%s:%s ", onion.PrivateKeyType, onion.PrivateKey)
	for remotePort, localAddr := range onion.Ports {
		req += fmt.Sprintf("Port=%d,%s ", remotePort, localAddr)
	}
	_, msg, err := c.makeRequest(req)
	if err != nil {
		return err
	}
	lines := strings.Split(msg, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if parts[0] == "ServiceID" {
			onion.ServiceID = parts[1]
		} else if parts[0] == "PrivateKey" {
			key := strings.SplitN(parts[1], ":", 2)
			onion.PrivateKeyType = key[0]
			onion.PrivateKey = key[1]
		}
	}
	return nil
}

// DeleteOnion deletes an onion by its serviceID (stop hidden service created
// by this controller).
func (c *Controller) DeleteOnion(serviceID string) error {
	_, _, err := c.makeRequest("DEL_ONION " + serviceID)
	if err != nil {
		return err
	}
	return nil
}

// Signal sends a signal to the server. Tor documentations defines the
// following signals :
//   * RELOAD
//   * SHUTDOWN
//   * DUMP
//   * DEBUG
//   * HALT
//   * CLEARDNSCACHE
//   * NEWNYM
//   * HEARTBEAT
//   * DORMANT
//   * ACTIVE
func (c *Controller) Signal(signal string) error {
	_, _, err := c.makeRequest("SIGNAL " + signal)
	if err != nil {
		return err
	}
	return nil
}
