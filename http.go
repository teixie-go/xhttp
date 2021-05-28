package xhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func resolveClient(client ...*http.Client) *http.Client {
	if len(client) > 0 {
		return client[0]
	}
	return http.DefaultClient
}

func Get(uri string, client ...*http.Client) ([]byte, error) {
	c := resolveClient(client...)
	response, err := c.Get(uri)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http get error : uri=%v , statusCode=%v", uri, response.StatusCode)
	}
	return ioutil.ReadAll(response.Body)
}

func Post(uri string, data string, client ...*http.Client) ([]byte, error) {
	c := resolveClient(client...)
	response, err := c.Post(uri, "", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http get error : uri=%v , statusCode=%v", uri, response.StatusCode)
	}
	return ioutil.ReadAll(response.Body)
}

func PostForm(uri string, data url.Values, client ...*http.Client) ([]byte, error) {
	c := resolveClient(client...)
	response, err := c.PostForm(uri, data)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http get error : uri=%v , statusCode=%v", uri, response.StatusCode)
	}
	return ioutil.ReadAll(response.Body)
}

func PostJSON(uri string, obj interface{}, client ...*http.Client) ([]byte, error) {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	c := resolveClient(client...)
	response, err := c.Post(uri, "application/json;charset=utf-8", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http get error : uri=%v , statusCode=%v", uri, response.StatusCode)
	}
	return ioutil.ReadAll(response.Body)
}
