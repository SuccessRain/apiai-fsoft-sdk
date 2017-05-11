package apiai

import (
	"encoding/json"
	"net/http"
	"fmt"
	"bytes"
	"net/url"
	"time"
)

// query text
//
// https://docs.api.ai/docs/query
func (c *Client) QueryText(query QueryRequest) (result QueryResponse, err error) {
	var bytes []byte
	if bytes, err = c.httpPost("query", nil, nil, query); err == nil {
		if err = json.Unmarshal(bytes, &result); err == nil {
			return result, nil
		}
	}

	return QueryResponse{}, err
}

// query voice in .wav(16000Hz, signed PCM, 16 bit, mono) format
//
// NOTE: this api requires paid plan
//
// https://docs.api.ai/docs/query
func (c *Client) QueryVoice(query QueryRequest, filepath string) (result QueryResponse, err error) {
	var bytes []byte
	if bytes, err = c.httpPostMultipart(
		"query",
		nil,
		map[string]interface{}{
			"request": query,
		},
		map[string]string{
			"voiceData": filepath,
		},
	); err == nil {
		if err = json.Unmarshal(bytes, &result); err == nil {
			return result, nil
		}
	}

	return QueryResponse{}, err
}
//************

type Event struct {
	Name string            `json:"name"`
	Data map[string]string `json:"data"`
}

type Context struct {
	Name     string            `json:"name"`
	Lifespan int               `json:"lifespan"`
	Params   map[string]string `json:"parameters"`
}

type EntityDescription struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Count   int    `json:"count"`
	Preview string `json:"preview"`
}

type Platform struct {
	Source string            `json:"source"`
	Data   map[string]string `json:"data"`
}

type Query struct {
	Query           []string            `json:"query"`
	Event           Event               `json:"event"`
	Version         string              `json:"-"`
	SessionId       string              `json:"sessionId"`
	Language        string              `json:"lang"`
	Contexts        []Context           `json:"contexts"`
	ResetContexts   bool                `json:"resetContexts"`
	Entities        []EntityDescription `json:"entities"`
	Timezone        string              `json:"timezone"`
	Location        Location            `json:"location"`
	OriginalRequest Platform            `json:"originalRequest"`
}

type Status struct {
	Code         int    `json:"code"`
	ErrorType    string `json:"errorType"`
	ErrorId      string `json:"errorId"`
	ErrorDetails string `json:"errorDetails"`
}

type Fulfilment struct {
	Speech   string    `json:"speech"`
	Messages []Message `json:"messages"`
}

type CardButton struct {
	Text     string
	Postback string
}

type Message struct {
	Type     int          `json:"type"`
	Speech   string       `json:"speech"`
	ImageUrl string       `json:"imageUrl"`
	Title    string       `json:"title"`
	Subtitle string       `json:"subtitle"`
	Buttons  []CardButton `json:"buttons"`
	Replies  []string     `json:"replies"`
	Payload  interface{}  `json:"payload"`
}

type Result struct {
	Source           string            `json:"source"`
	ResolvedQuery    string            `json:"resolvedQuery"`
	Action           string            `json:"action"`
	ActionIncomplete bool              `json:"actionIncomplete"`
	Params           map[string]string `json:"parameters"`
	Contexts         []Context         `json:"contexts"`
	Fulfillment      Fulfilment        `json:"fulfillment"`
	Score            float64           `json:"score"`
	Metadata         Metadata          `json:"metadata"`
}

type QueryResponse2 struct {
	Id        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Language  string    `json:"lang"`
	Result    Result    `json:"result"`
	Status    Status    `json:"status"`
	SessionId string    `json:"sessionId"`
}

const baseUrl = "https://api.api.ai/v1/"
const defaultVersion = "20150910"
const defaultQueryLang = "en"
const defaultSpeechLang = "en-US"
//const token = "75ec7ca2d07144eabad83f84ec1ac806"
//const session  = "acaa16d5ceaa4c18897c7ac523846fdb"


func (c *Client) Query(q Query, token string, session string) (*QueryResponse2, error) {
	q.Version = defaultVersion
	q.SessionId = session
	q.Language = defaultQueryLang
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(q)

	if err != nil {
		return nil, fmt.Errorf("Error on request, %v", err)
	}

	req, err := http.NewRequest("POST", c.buildUrl("query", nil), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-type", "application/json, charset=utf-8")
	req.Header.Set("Authorization", "Bearer " + token)

	httpClient := http.DefaultClient
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response *QueryResponse2
	switch resp.StatusCode {
	case http.StatusOK:
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&response)
		if err != nil {
			return nil, err
		}
		return response, nil
	default:
		return nil, fmt.Errorf("Error: status code is %v", resp.StatusCode)
	}
}

func (c *Client) buildUrl(endpoint string, params map[string]string) string {
	u := baseUrl + endpoint + "?v=" + defaultVersion
	if params != nil {
		for i, v := range params {
			u += "&" + i + "=" + url.QueryEscape(v)
		}
	}
	return u
}