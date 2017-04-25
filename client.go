package apiai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

const (
	Version = "20150910" // https://docs.api.ai/docs/versioning

	BaseUrl = "https://api.api.ai/v1"
)

type Client struct {
	AccessToken string `json:"access_token"`
	Verbose     bool   `json:"verbose"`
}

// get a new api client with given access token
func NewClient(accessToken string) *Client {
	return &Client{
		AccessToken: accessToken,
		Verbose:     false,
	}
}

// generate api url
func apiUrl(api string) string {
	return fmt.Sprintf("%s/%s?v=%s", BaseUrl, api, Version)
}

// get http header for authorization
func (c *Client) authHeader() string {
	return fmt.Sprintf("Bearer %s", c.AccessToken)
}

// http get
func (c *Client) httpGet(api string, headers, params map[string]string) (result []byte, err error) {
	url := apiUrl(api)
	if c.Verbose {
		log.Printf("[GET] requesting url: %s, headers: %+v, params: %+v\n", url, headers, params)
	}

	var req *http.Request
	if req, err = http.NewRequest("GET", url, nil); err == nil {
		req.Header.Set("Authorization", c.authHeader())
		for k, v := range headers { // additional http headers
			req.Header.Set(k, v)
		}

		// get params
		query := req.URL.Query()
		for k, v := range params {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()

		var resp *http.Response
		client := &http.Client{}
		if resp, err = client.Do(req); err == nil {
			defer resp.Body.Close()

			if result, err = ioutil.ReadAll(resp.Body); err == nil {
				if c.Verbose {
					log.Printf("response body: %s\n", string(result))
				}

				return result, err
			}
		}
	}

	return []byte{}, err
}

// http post, put, or delete (json)
func (c *Client) httpPostPutDelete(method, api string, headers, params map[string]string, object interface{}) (result []byte, err error) {
	url := apiUrl(api)
	if c.Verbose {
		log.Printf("[%s] requesting url: %s, headers: %+v, object: %+v\n", method, url, headers, object)
	}

	var data []byte
	if data, err = json.Marshal(object); err == nil {
		var req *http.Request
		if req, err = http.NewRequest(strings.ToUpper(method), url, bytes.NewBuffer(data)); err == nil {
			req.Header.Set("Authorization", c.authHeader())
			req.Header.Set("Content-Type", "application/json;charset=utf-8")
			for k, v := range headers { // additional http headers
				req.Header.Set(k, v)
			}

			// get params
			query := req.URL.Query()
			for k, v := range params {
				query.Add(k, v)
			}
			req.URL.RawQuery = query.Encode()

			var resp *http.Response
			client := &http.Client{}
			if resp, err = client.Do(req); err == nil {
				defer resp.Body.Close()

				if result, err = ioutil.ReadAll(resp.Body); err == nil {
					if c.Verbose {
						log.Printf("response body: %s\n", string(result))
					}

					return result, nil
				}
			}
		}
	}

	return []byte{}, err
}

// http post (json)
func (c *Client) httpPost(api string, headers, params map[string]string, object interface{}) (result []byte, err error) {
	return c.httpPostPutDelete("POST", api, headers, params, object)
}

// http put (json)
func (c *Client) httpPut(api string, headers map[string]string, object interface{}) (result []byte, err error) {
	return c.httpPostPutDelete("PUT", api, headers, nil, object)
}

// http delete
func (c *Client) httpDelete(api string, headers, params map[string]string, object interface{}) (result []byte, err error) {
	return c.httpPostPutDelete("DELETE", api, headers, params, object)
}

// http post (multipart)
func (c *Client) httpPostMultipart(api string, headers map[string]string, params map[string]interface{}, filepaths map[string]string) (result []byte, err error) {
	url := apiUrl(api)
	if c.Verbose {
		log.Printf("requesting url: %s, headers: %+v, params: %+v, filepaths: %+v\n", url, headers, params, filepaths)
	}

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	var fw io.Writer

	// write strings
	for k, v := range params {
		if fw, err = writer.CreateFormField(k); err != nil {
			return []byte{}, err
		}
		var data []byte
		if data, err = json.Marshal(v); err != nil {
			return []byte{}, err
		}
		if _, err = fw.Write(data); err != nil {
			return []byte{}, err
		}
	}

	// write file
	for k, v := range filepaths {
		if fw, err = writer.CreateFormFile(k, v); err != nil {
			return []byte{}, err
		}

		var file *os.File
		if file, err = os.Open(v); err != nil {
			return []byte{}, err
		}
		defer file.Close()
		if _, err = io.Copy(fw, file); err != nil {
			return []byte{}, err
		}
	}

	// close writer
	writer.Close()

	var req *http.Request
	if req, err = http.NewRequest("POST", url, &buffer); err == nil {
		req.Header.Set("Authorization", c.authHeader())
		req.Header.Set("Content-Type", writer.FormDataContentType())
		for k, v := range headers { // additional http headers
			req.Header.Set(k, v)
		}

		var resp *http.Response
		client := &http.Client{}
		if resp, err = client.Do(req); err == nil {
			defer resp.Body.Close()

			var bytes []byte
			if bytes, err = ioutil.ReadAll(resp.Body); err == nil {
				if c.Verbose {
					log.Printf("response: %s\n", string(bytes))
				}

				return bytes, nil
			}
		}
	}

	return []byte{}, err
}
