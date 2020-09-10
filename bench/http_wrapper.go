package bench

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pingcap/errors"
)

var (
	dialClient = &http.Client{}
)

type bodyOption struct {
	contentType string
	body        io.Reader
}

// BodyOption sets the type and content of the body
type BodyOption func(*bodyOption)

// WithBody returns a BodyOption
func WithBody(contentType string, body io.Reader) BodyOption {
	return func(bo *bodyOption) {
		bo.contentType = contentType
		bo.body = body
	}
}

func doRequest(url, method string,
	opts ...BodyOption) (string, error) {
	var resp string
	if method == "" {
		method = http.MethodGet
	}
	b := &bodyOption{}
	for _, o := range opts {
		o(b)
	}

	req, err := http.NewRequest(method, url, b.body)
	if err != nil {
		return "", err
	}
	if b.contentType != "" {
		req.Header.Set("Content-Type", b.contentType)
	}
	// the resp would be returned by the outer function
	resp, err = dial(req)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func dial(req *http.Request) (string, error) {
	resp, err := dialClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var msg []byte
		msg, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", errors.Errorf("[%d] %s", resp.StatusCode, msg)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func postJSON(url string, input map[string]interface{}) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}

	var msg []byte
	var r *http.Response
	r, err = dialClient.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		msg, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		return errors.Errorf("[%d] %s", r.StatusCode, msg)
	}
	return nil
}
