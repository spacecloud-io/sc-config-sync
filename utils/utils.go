package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/spaceuptech/helpers"

	"github.com/spaceuptech/sc-config-sync/model"
)

// CreateSpecObject returns the string equivalent of the git op object
func CreateSpecObject(api, objType string, meta map[string]string, spec interface{}) (*model.SpecObject, error) {
	v := model.SpecObject{
		API:  api,
		Type: objType,
		Meta: meta,
		Spec: spec,
	}

	return &v, nil
}

// MakeHTTPRequest gets spec object
func MakeHTTPRequest(token, method, url string, params map[string]string, vPtr interface{}) error {
	url = fmt.Sprintf("%s%s", model.GatewayAddr, url)

	reqBody, _ := json.Marshal(map[string]string{})
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer CloseTheCloser(resp.Body)

	data, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		respBody := map[string]interface{}{}
		if err := json.Unmarshal(data, &respBody); err != nil {
			return err
		}
		return helpers.Logger.LogError("", fmt.Sprintf("error while getting service got http status code %s", resp.Status), fmt.Errorf("%v", respBody["error"]), nil)
	}

	if err := json.Unmarshal(data, vPtr); err != nil {
		return err
	}

	return nil
}

// CloseTheCloser closes the closer
func CloseTheCloser(c io.Closer) {
	_ = c.Close()
}
