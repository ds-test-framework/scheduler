package util

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
)

var (
	ErrSendFailed       = errors.New("sending failed")
	ErrResponseReadFail = errors.New("failed to read response")
	ErrBadResponse      = errors.New("bad response")
)

// RequestOption can be used to modify the request that is to be sent
type RequestOption func(*http.Request)

// JsonRequest sets the content type to application/json
func JsonRequest() RequestOption {
	return func(r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
	}
}

func SendMsg(method, toAddr, msg string, options ...RequestOption) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, "http://"+toAddr, bytes.NewBuffer([]byte(msg)))
	if err != nil {
		return "", ErrSendFailed
	}

	for _, o := range options {
		o(req)
	}

	// logger.Debug(fmt.Sprintf("Transport: Sending message: %s", msg))
	resp, err := client.Do(req)
	if err != nil {
		return "", ErrSendFailed
	}
	defer resp.Body.Close()
	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if statusOK {
		bodyB, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", ErrResponseReadFail
		}
		return string(bodyB), nil
	}
	return "", ErrBadResponse
}
