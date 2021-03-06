package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"
	"time"
)

const testServer = "127.0.0.1:62621"
const reqForm = 16777216

var client *Client

func init() {
	// Creating a new test server.
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseMultipartForm(reqForm)

			getValues, _ := url.ParseQuery(r.URL.RawQuery)

			response := map[string]interface{}{
				"method": r.Method,
				"proto":  r.Proto,
				"host":   r.Host,
				"header": r.Header,
				"url":    r.URL.String(),
				"get":    getValues,
				"post":   r.Form,
			}
			if r.Body != nil {
				response["body"], _ = ioutil.ReadAll(r.Body)
			}

			w.Header().Set("Content-Type", "application/json")

			if r.MultipartForm != nil {
				files := map[string]interface{}{}
				for key, val := range r.MultipartForm.File {
					files[key] = val
				}
				response["files"] = files
			}
			data, err := json.Marshal(response)
			if err == nil {
				w.Write(data)
			}
		},
	)
	go http.ListenAndServe(testServer, nil)

	time.Sleep(time.Second * 1)
}

func TestInit(t *testing.T) {
	var err error
	client, err = New("http://" + testServer)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	var buf map[string]interface{}
	var err error

	err = client.Get(&buf, "/search", url.Values{"term": {"some string"}})

	if err != nil {
		t.Fatalf("Failed test: %s\n", err.Error())
	}

	if buf["method"].(string) != "GET" {
		t.Fatalf("Test failed.")
	}

	if buf["url"].(string) != "/search?term=some+string" {
		t.Fatalf("Test failed.")
	}

	if buf["get"].(map[string]interface{})["term"].([]interface{})[0].(string) != "some string" {
		t.Fatalf("Test failed.")
	}

	err = client.Get(&buf, "/search", nil)

	if err != nil {
		t.Fatalf("Failed test: %s\n", err.Error())
	}

	if buf["method"].(string) != "GET" {
		t.Fatalf("Test failed.")
	}
}

func TestPost(t *testing.T) {
	var buf map[string]interface{}
	var err error

	err = client.Post(&buf, "/search?foo=the+quick", url.Values{"bar": {"brown fox"}})

	if err != nil {
		t.Fatalf("Failed test: %s\n", err.Error())
	}

	if buf["method"].(string) != "POST" {
		t.Fatalf("Test failed.")
	}

	if buf["post"].(map[string]interface{})["bar"].([]interface{})[0].(string) != "brown fox" {
		t.Fatalf("Test failed.")
	}

	if buf["get"].(map[string]interface{})["foo"].([]interface{})[0].(string) != "the quick" {
		t.Fatalf("Test failed.")
	}

	err = client.Post(&buf, "/search?foo=the+quick", nil)

	if err != nil {
		t.Fatalf("Failed test: %s\n", err.Error())
	}

	if buf["method"].(string) != "POST" {
		t.Fatalf("Test failed.")
	}

	if buf["get"].(map[string]interface{})["foo"].([]interface{})[0].(string) != "the quick" {
		t.Fatalf("Test failed.")
	}
}

func TestPut(t *testing.T) {
	var buf map[string]interface{}
	var err error

	err = client.Put(&buf, "/search?foo=the+quick", url.Values{"bar": {"brown fox"}})

	if err != nil {
		t.Fatalf("Failed test: %s\n", err.Error())
	}

	if buf["method"].(string) != "PUT" {
		t.Fatalf("Test failed.")
	}

	if buf["post"].(map[string]interface{})["bar"].([]interface{})[0].(string) != "brown fox" {
		t.Fatalf("Test failed.")
	}

	if buf["get"].(map[string]interface{})["foo"].([]interface{})[0].(string) != "the quick" {
		t.Fatalf("Test failed.")
	}

	err = client.Put(&buf, "/search?foo=the+quick", nil)

	if err != nil {
		t.Fatalf("Failed test: %s\n", err.Error())
	}

	if buf["method"].(string) != "PUT" {
		t.Fatalf("Test failed.")
	}

	if buf["get"].(map[string]interface{})["foo"].([]interface{})[0].(string) != "the quick" {
		t.Fatalf("Test failed.")
	}
}

func TestDelete(t *testing.T) {
	var buf map[string]interface{}
	var err error

	err = client.Delete(&buf, "/search?foo=the+quick", url.Values{"bar": {"brown fox"}})

	if err != nil {
		t.Fatalf("Failed test: %s\n", err.Error())
	}

	if buf["method"].(string) != "DELETE" {
		t.Fatalf("Test failed.")
	}

	if buf["get"].(map[string]interface{})["foo"].([]interface{})[0].(string) != "the quick" {
		t.Fatalf("Test failed.")
	}

	err = client.Delete(&buf, "/search?foo=the+quick", nil)

	if err != nil {
		t.Fatalf("Failed test: %s\n", err.Error())
	}

	if buf["method"].(string) != "DELETE" {
		t.Fatalf("Test failed.")
	}

	if buf["get"].(map[string]interface{})["foo"].([]interface{})[0].(string) != "the quick" {
		t.Fatalf("Test failed.")
	}
}

func TestPostMultipart(t *testing.T) {
	fileRest, err := os.Open("rest.go")

	if err != nil {
		panic(err)
	}

	defer fileRest.Close()

	files := map[string][]File{
		"file": []File{
			File{
				path.Base(fileRest.Name()),
				fileRest,
			},
		},
	}

	body, err := NewMultipartBody(url.Values{"foo": {"bar"}}, files)

	var buf map[string]interface{}

	err = client.PostMultipart(&buf, "/post", body)

	if buf["method"].(string) != "POST" {
		t.Fatalf("Test failed.")
	}

	if buf["files"].(map[string]interface{})["file"].([]interface{})[0].(map[string]interface{})["Filename"].(string) != "rest.go" {
		t.Fatalf("Test failed.")
	}

	body, err = NewMultipartBody(nil, files)

	err = client.PostMultipart(&buf, "/post", body)

	if buf["method"].(string) != "POST" {
		t.Fatalf("Test failed.")
	}

	if buf["files"].(map[string]interface{})["file"].([]interface{})[0].(map[string]interface{})["Filename"].(string) != "rest.go" {
		t.Fatalf("Test failed.")
	}

	body, err = NewMultipartBody(url.Values{"foo": {"bar"}}, nil)

	err = client.PostMultipart(&buf, "/post", body)

	if buf["method"].(string) != "POST" {
		t.Fatalf("Test failed.")
	}

	if buf["post"].(map[string]interface{})["foo"].([]interface{})[0].(string) != "bar" {
		t.Fatalf("Test failed.")
	}
}

func TestPutMultipart(t *testing.T) {
	fileRest, err := os.Open("rest.go")

	if err != nil {
		panic(err)
	}

	defer fileRest.Close()

	files := map[string][]File{
		"file": []File{
			File{
				path.Base(fileRest.Name()),
				fileRest,
			},
		},
	}

	body, err := NewMultipartBody(url.Values{"foo": {"bar"}}, files)

	var buf map[string]interface{}

	err = client.PutMultipart(&buf, "/put", body)

	if buf["method"].(string) != "PUT" {
		t.Fatalf("Test failed.")
	}

	if buf["files"].(map[string]interface{})["file"].([]interface{})[0].(map[string]interface{})["Filename"].(string) != "rest.go" {
		t.Fatalf("Test failed.")
	}

	body, err = NewMultipartBody(nil, files)

	err = client.PutMultipart(&buf, "/put", body)

	if buf["method"].(string) != "PUT" {
		t.Fatalf("Test failed.")
	}

	if buf["files"].(map[string]interface{})["file"].([]interface{})[0].(map[string]interface{})["Filename"].(string) != "rest.go" {
		t.Fatalf("Test failed.")
	}

	body, err = NewMultipartBody(url.Values{"foo": {"bar"}}, nil)

	err = client.PutMultipart(&buf, "/put", body)

	if buf["method"].(string) != "PUT" {
		t.Fatalf("Test failed.")
	}

	if buf["post"].(map[string]interface{})["foo"].([]interface{})[0].(string) != "bar" {
		t.Fatalf("Test failed.")
	}
}


func TestDefaultClient(t *testing.T) {
	var err error
	var buf []byte
	err = Get(&buf, "https://github.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(buf) == 0 {
		t.Fatalf("Expecting something in buf.")
	}
}

func TestGetJSONMap(t *testing.T) {
	var err error

	var res map[string]interface{}
	err = Get(&res, "http://ip.jsontest.com", nil)

	if err != nil {
		t.Fatal(err)
	}

	if _, ok := res["ip"]; ok == false {
		t.Fatalf(`Expecting key "ip".`)
	}

	t.Logf("Your IP address is: %s", res["ip"])
}

func TestGetStruct(t *testing.T) {
	var err error

	type ip_t struct {
		IP string `json:"ip"`
	}

	var res ip_t

	if err = Get(&res, "http://ip.jsontest.com", nil); err != nil {
		t.Fatal(err)
	}

	if res.IP == "" {
		t.Fatalf("Expecting IP value.")
	}

	t.Logf("Your IP address is: %s", res.IP)
}

func TestBasicAuth(t *testing.T) {
	var buf []byte

	type basicAuth_t struct {
		Header struct {
			Authorization []string `json:"authorization"`
		} `json:"header"`
	}

	var res basicAuth_t

	client, err := New("http://" + testServer)

	if err != nil {
		t.Fatal(err)
	}

	client.SetBasicAuth("foo", "bar")

	if client.Header.Get("Authorization") != "Basic Zm9vOmJhcg==" {
		t.Fatalf("Failed to encode foo:bar.")
	}

	res = basicAuth_t{}
	if err = client.Get(&buf, "/auth", nil); err != nil {
		t.Fatal(err)
	}

	json.Unmarshal(buf, &res)

	if res.Header.Authorization[0] != "Basic Zm9vOmJhcg==" {
		t.Fatalf("Failed to send foo:bar.")
	}

	res = basicAuth_t{}
	if err = client.Get(&res, "/auth", nil); err != nil {
		t.Fatal(err)
	}
}
