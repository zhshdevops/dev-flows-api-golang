package coderepo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func ReadBody(resp *http.Response, obj interface{}) error {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = json.Unmarshal(data, obj)
	if err != nil {
		return err
	}
	return nil
}
