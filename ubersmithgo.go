// API client to access Ubersmith API
//
// Ubersmith API documentation itself can be found at http://www.ubersmith.com/kbase/index.php?_m=downloads&_a=view&parentcategoryid=2
package ubersmithgo

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Request convenient way to define map[string]string
type Request map[string]interface{}

// API client structure
type API struct {
	Host    string
	User    string
	Token   string
	DebugOn bool
}

// New creating new API client instance
func New(host string, user string, token string, debug bool) *API {
	return &API{
		Host:    host,
		User:    user,
		Token:   token,
		DebugOn: debug,
	}
}

// Call the method for calling ubersmith API
// 	api := ubersmithgo.New("http://yourubersmithurl.com/api/2.0/", "username", "token")
// 	r := api.Call("support.ticket_submit", uber.Request{
//		"subject":   "subject",
//		"body":      "message here",
//		"name":      "Admin",
//	})
func (a *API) Call(method string, params Request) *Response {
	r := Response{}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{Transport: tr}

	data := make(url.Values)
	data.Set("method", method)

	pageURL := a.Host + "?" + data.Encode()
	body, err := json.Marshal(params)

	req, err := http.NewRequest("POST", pageURL, bytes.NewBuffer(body))

	a.debug("[API REQUEST]: " + pageURL)
	a.debug("[API REQUEST DATA]: " + string(body))

	if err != nil {
		r.Status = false
		r.ErrorCode = 500
		r.ErrorMesssage = err.Error()
		return &r
	}

	req.Header.Set("Content-type", "application/json")
	req.SetBasicAuth(a.User, a.Token)

	res, err := c.Do(req)
	if err != nil {
		r.Status = false
		r.ErrorCode = 500
		r.ErrorMesssage = err.Error()
		return &r
	}

	defer res.Body.Close()

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		r.Status = false
		r.ErrorCode = 500
		r.ErrorMesssage = err.Error()
		return &r
	}

	a.debug("[API RESPONSE]: " + string(result))

	json.Unmarshal(result, &r)
	return &r
}

func (a *API) debug(i interface{}) {
	if a.DebugOn {
		log.Printf("%v\n", i)
	}
}

// Response handling response from Ubersmith
type Response struct {
	Status        bool            `json:"status"`
	ErrorCode     int             `json:"error_code"`
	ErrorMesssage string          `json:"error_message"`
	RawData       json.RawMessage `json:"data"`
}

// Key retrieve the data from the JSON API response by key
// 	api := ubersmithgo.New("http://yourubersmithurl.com/api/2.0/", "username", "token")
// 	r := api.Call("client.get", uber.Request{
//		"client_id":   "1000",
//	})
// 	name := r.Key("full_name").(string)
// 	tag1 := r.Key("tags.1.tag").(string)
func (r *Response) Key(key string) interface{} {
	d := make(map[string]interface{})
	json.Unmarshal(r.RawData, &d)

	keys := strings.Split(key, ".")
	var res interface{}

	res = d[keys[0]]
	for i := 1; i < len(keys); i++ {
		switch res.(type) {
		case map[string]interface{}:
			res = res.(map[string]interface{})[keys[i]]
		case []interface{}:
			length := len(res.([]interface{}))
			k := toInt(keys[i])
			if length > k {
				res = res.([]interface{})[k]
			} else {
				res = nil
			}
		}
	}
	return res
}

// Load loading the JSON response into interface
//
//	type client struct {
//		FirstName string `json:"first"`
//		LastName string `json:"last"`
//		Email string `json:"email"`
//	}
//
// 	api := ubersmithgo.New("http://yourubersmithurl.com/api/2.0/", "username", "token")
// 	r := api.Call("client.get", uber.Request{
//		"client_id":   "1000",
//	})
//	cl := client{}
//	r.Load(&cl)
func (r *Response) Load(i interface{}) error {
	return json.Unmarshal(r.RawData, &i)
}

func toInt(num string) int {
	i, _ := strconv.Atoi(num)
	return i
}
