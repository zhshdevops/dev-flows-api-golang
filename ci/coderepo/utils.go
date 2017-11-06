package coderepo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/golang/glog"
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
	glog.Infoln("data info :",string(data))
	return nil
}
