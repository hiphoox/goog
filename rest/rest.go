// Utility package for Go that makes working with web services even easier.
package rest

import (
  "bytes"
  "compress/gzip"
  "encoding/base64"
  "encoding/json"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "mime/multipart"
  "net/http"
  "net/http/cookiejar"
  "net/url"
  "os"
  "path"
  "reflect"
  "strings"
)

var enableDebug = true

var (
  ioReadCloserType = reflect.TypeOf((*io.ReadCloser)(nil)).Elem()
  bytesBufferType  = reflect.TypeOf((**bytes.Buffer)(nil)).Elem()
  restResponseType = reflect.TypeOf((*Response)(nil)).Elem()
)

// Response can be used as a response value, useful when you need to work with
// headers, status codes and such.
type Response struct {
  Status        string
  StatusCode    int
  Proto         string
  ProtoMajor    int
  ProtoMinor    int
  ContentLength int64
  http.Header
  Body []byte
}

// File can be used to represent a file that you'll later upload within a
// multipart request.
type File struct {
  Name string
  io.Reader
}

// MultipartBody struct for multipart requests, you can't generate a
// MultipartBody directly, use rest.NewMultipartBody() instead.
type MultipartBody struct {
  contentType string
  buf         io.Reader
}

// Client is useful in case you need to communicate with an API and you'd like
// to use the same prefix for all of your requests or in scenarios where it
// would be handy to keep a session cookie.
type Client struct {
  Header    http.Header
  Prefix    string
  CookieJar *cookiejar.Jar
}

// DefaulClient is the default client used on top level functions like
// rest.Get(), rest.Post(), rest.Delete() and rest.Put().
var DefaultClient = new(Client)

func init() {
  // If the enviroment variable REST_DEBUG is present, we enable verbose
  // logging.
  if os.Getenv("REST_DEBUG") != "" {
    enableDebug = true
  }
}

// New creates a new client, in all following GET, POST, PUT and DELETE
// requests given paths will be prefixed with the given client's prefix value.
func New(prefix string) (*Client, error) {
  var err error

  if _, err = url.Parse(prefix); err != nil {
    return nil, fmt.Errorf(ErrInvalidPrefix.Error(), err.Error())
  }

  self := new(Client)
  self.Prefix = strings.TrimRight(prefix, "/") + "/"
  self.Header = http.Header{}

  if self.CookieJar, err = cookiejar.New(nil); err != nil {
    return nil, err
  }

  return self, nil
}

// Taken from net/http
func basicAuth(username, password string) string {
  auth := username + ":" + password
  return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Sets the request's basic authorization header to be used in all requests.
func (self *Client) SetBasicAuth(username string, password string) {
  self.Header.Set("Authorization", "Basic "+basicAuth(username, password))
}

// Sets the request's basic authorization header to be used in all requests.
func (client *Client) GetHeader(header_name string) string {
  return client.Header.Get(header_name)
}

func (self *Client) newMultipartRequest(dst interface{}, method string, addr *url.URL, body *MultipartBody) error {
  var res *http.Response
  var req *http.Request

  var err error

  if body == nil {
    return ErrCouldNotCreateMultipart
  }

  if req, err = http.NewRequest(method, addr.String(), body.buf); err != nil {
    return err
  }

  req.Header.Set("Content-Type", body.contentType)

  if res, err = self.do(req); err != nil {
    return err
  }

  if err = self.handleResponse(dst, res); err != nil {
    return err
  }

  return nil
}

func (self *Client) newRequest(dst interface{}, method string, addr *url.URL, body *strings.Reader) error {
  var res *http.Response
  var req *http.Request

  var err error

  if body == nil {
    if req, err = http.NewRequest(method, addr.String(), nil); err != nil {
      return err
    }
  } else {
    if req, err = http.NewRequest(method, addr.String(), body); err != nil {
      return err
    }
  }

  switch method {
  case "POST", "PUT":
    if req.Header.Get("Content-Type") == "" {
      req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
    }
  }

  if res, err = self.do(req); err != nil {
    return err
  }

  if err = self.handleResponse(dst, res); err != nil {
    return err
  }

  return nil
}

// Put performs a HTTP PUT request and, when complete, attempts to convert the
// response body into the datatype given by dst (a pointer to a struct, map or
// []byte array).
func (self *Client) Put(dst interface{}, path string, data url.Values) error {
  var addr *url.URL
  var err error
  var body *strings.Reader

  if addr, err = url.Parse(self.Prefix + strings.TrimLeft(path, "/")); err != nil {
    return err
  }

  if data != nil {
    body = strings.NewReader(data.Encode())
  }

  return self.newRequest(dst, "PUT", addr, body)
}

// Delete performs a HTTP DELETE request and, when complete, attempts to
// convert the response body into the datatype given by dst (a pointer to a
// struct, map or []byte array).
func (self *Client) Delete(dst interface{}, path string, data url.Values) error {
  var addr *url.URL
  var err error
  var body *strings.Reader

  if addr, err = url.Parse(self.Prefix + strings.TrimLeft(path, "/")); err != nil {
    return err
  }

  if data != nil {
    body = strings.NewReader(data.Encode())
  }

  return self.newRequest(dst, "DELETE", addr, body)
}

// PutMultipart performs a HTTP PUT multipart request and, when complete,
// attempts to convert the response body into the datatype given by dst (a
// pointer to a struct, map or []byte array).
func (self *Client) PutMultipart(dst interface{}, uri string, data *MultipartBody) error {
  var addr *url.URL
  var err error

  if addr, err = url.Parse(self.Prefix + strings.TrimLeft(uri, "/")); err != nil {
    return err
  }

  return self.newMultipartRequest(dst, "PUT", addr, data)
}

// PostMultipart performs a HTTP POST multipart request and, when complete,
// attempts to convert the response body into the datatype given by dst (a
// pointer to a struct, map or []byte array).
func (self *Client) PostMultipart(dst interface{}, uri string, data *MultipartBody) error {
  var addr *url.URL
  var err error

  if addr, err = url.Parse(self.Prefix + strings.TrimLeft(uri, "/")); err != nil {
    return err
  }

  return self.newMultipartRequest(dst, "POST", addr, data)
}

// PostRaw performs a HTTP POST request with a custom body and, when complete,
// attempts to convert the response body into the datatype given by dst (a
// pointer to a struct, map or []byte array).
func (self *Client) PostRaw(dst interface{}, path string, body []byte) error {
  var addr *url.URL
  var err error
  var bodyReader *strings.Reader

  if addr, err = url.Parse(self.Prefix + strings.TrimLeft(path, "/")); err != nil {
    return err
  }

  if body != nil {
    bodyReader = strings.NewReader(string(body))
  }

  return self.newRequest(dst, "POST", addr, bodyReader)
}

// Post performs a HTTP POST request and, when complete, attempts to convert
// the response body into the datatype given by dst (a pointer to a struct, map
// or []byte array).
func (self *Client) Post(dst interface{}, path string, data url.Values) error {
  var addr *url.URL
  var err error
  var body *strings.Reader

  if addr, err = url.Parse(self.Prefix + strings.TrimLeft(path, "/")); err != nil {
    return err
  }

  if data != nil {
    body = strings.NewReader(data.Encode())
  }

  return self.newRequest(dst, "POST", addr, body)
}

// Get performs a HTTP GET request and, when complete, attempts to convert the
// response body into the datatype given by dst (a pointer to a struct, map or
// []byte array).
func (self *Client) Get(dst interface{}, path string, data url.Values) error {
  var addr *url.URL
  var err error

  if addr, err = url.Parse(self.Prefix + strings.TrimLeft(path, "/")); err != nil {
    return err
  }

  if data != nil {
    if addr.RawQuery == "" {
      addr.RawQuery = data.Encode()
    } else {
      addr.RawQuery = addr.RawQuery + "&" + data.Encode()
    }
  }

  return self.newRequest(dst, "GET", addr, nil)
}

// We don't need any GET vars
// We also don't care about the response body
func (client *Client) GetHeaders(path string) error {
  // var buf []byte
  err := client.Get(nil, path, nil)
      if err.Error() == "EOF" {
      err = nil
    }
  return err
}

// NewMultipartBody creates a *MultipartBody based on the given parameters.
// This is useful for PostMultipart() and PutMultipart().
func NewMultipartBody(params url.Values, filemap map[string][]File) (*MultipartBody, error) {

  dst := bytes.NewBuffer(nil)

  body := multipart.NewWriter(dst)

  if filemap != nil {
    for key, files := range filemap {

      for _, file := range files {

        writer, err := body.CreateFormFile(key, path.Base(file.Name))

        if err != nil {
          return nil, err
        }

        if _, err = io.Copy(writer, file.Reader); err != nil {
          return nil, err
        }
      }
    }
  }

  if params != nil {
    for key := range params {
      for _, value := range params[key] {
        body.WriteField(key, value)
      }
    }
  }

  body.Close()

  return &MultipartBody{body.FormDataContentType(), dst}, nil
}

// Returns the body of the request as a io.ReadCloser
func (self *Client) body(res *http.Response) (io.ReadCloser, error) {
  var body io.ReadCloser
  var err error

  if res.Header.Get("Content-Encoding") == "gzip" {
    if body, err = gzip.NewReader(res.Body); err != nil {
      return nil, err
    }
  } else {
    body = res.Body
  }

  return body, nil
}

func fromBytes(dst reflect.Value, buf []byte) error {
  var err error

  switch dst.Kind() {
  case reflect.String:
    // string
    dst.Set(reflect.ValueOf(string(buf)))
    return nil
  case reflect.Slice:
    switch dst.Type().Elem().Kind() {
    // []byte
    case reflect.Uint8:
      dst.Set(reflect.ValueOf(buf))
      return nil
    // []interface{}
    case reflect.Interface:
      t := []interface{}{}
      err = json.Unmarshal(buf, &t)

      if err == nil {
        dst.Set(reflect.ValueOf(t))
        return nil
      }
    }
  case reflect.Map:
    switch dst.Type().Elem().Kind() {
    case reflect.Interface:
      // map[string] interface{}
      m := map[string]interface{}{}

      err = json.Unmarshal(buf, &m)

      if err == nil {
        dst.Set(reflect.ValueOf(m))
        return nil
      }
    }
  }

  if err != nil {
    return err
  }

  return fmt.Errorf(ErrCouldNotConvert.Error(), reflect.TypeOf(buf), dst.Type())
}

func (self *Client) handleResponse(dst interface{}, res *http.Response) error {

  body, err := self.body(res)

  if err != nil {
    return err
  }

  if dst == nil {
    return nil
  }
  rv := reflect.ValueOf(dst)

  if rv.Kind() != reflect.Ptr || rv.IsNil() {
    return ErrDestinationNotAPointer
  }

  t := res.Header.Get("Content-Type")

  switch rv.Elem().Type() {
  case restResponseType:
    var err error

    r := Response{}

    r.Body, err = ioutil.ReadAll(body)

    if enableDebug == true {
      log.Printf("Body:\n%s\n", string(r.Body))
    }

    if err != nil {
      return err
    }

    r.Header = res.Header
    r.Status = res.Status
    r.StatusCode = res.StatusCode
    r.Proto = res.Proto
    r.ProtoMajor = res.ProtoMajor
    r.ProtoMinor = res.ProtoMinor
    r.ContentLength = res.ContentLength

    rv.Elem().Set(reflect.ValueOf(r))
  case ioReadCloserType:
    rv.Elem().Set(reflect.ValueOf(body))
  case bytesBufferType:
    buf, err := ioutil.ReadAll(body)

    if enableDebug == true {
      log.Printf("Body:\n%s\n", string(buf))
    }

    if err != nil {
      return err
    }

    dst := bytes.NewBuffer(buf)

    rv.Elem().Set(reflect.ValueOf(dst))
  default:
    buf, err := ioutil.ReadAll(body)

    if enableDebug == true {
      log.Printf("Body:\n%s\n", string(buf))
    }

    if err != nil {
      return err
    }

    if strings.HasPrefix(t, "application/json") == true {
      if rv.Elem().Kind() == reflect.Struct || rv.Elem().Kind() == reflect.Map {
        err = json.Unmarshal(buf, dst)
        return err
      }
    }

    err = fromBytes(rv.Elem(), buf)

    if err != nil {
      return err
    }
  }

  return nil
}

func (self *Client) do(req *http.Request) (*http.Response, error) {
  client := new(http.Client)

  // Adding cookie jar
  if self.CookieJar != nil {
    client.Jar = self.CookieJar
  }

  // Copying headers
  for k := range self.Header {
    req.Header.Set(k, self.Header.Get(k))
  }

  if req.Body == nil {
    req.Header.Del("Content-Type")
    req.Header.Del("Content-Length")
  }

  res, err := client.Do(req)

  if enableDebug == true {

    log.Printf("Fetching %v\n", req.URL.String())

    log.Printf("> %s %s", req.Method, req.Proto)
    for k := range req.Header {
      for kk := range req.Header[k] {
        log.Printf("> %s: %s", k, req.Header[k][kk])
      }
    }

    log.Printf("< %s %s", res.Proto, res.Status)
    for k := range res.Header {
      for kk := range res.Header[k] {
        log.Printf("< %s: %s", k, res.Header[k][kk])
      }
    }

    log.Printf("\n")
  }

  return res, err
}

// Get performs a HTTP GET request using the default client and, when complete,
// attempts to convert the response body into the datatype given by dst (a
// pointer to a struct, map or []byte array).
func Get(dest interface{}, uri string, data url.Values) error {
  return DefaultClient.Get(dest, uri, data)
}

// Post performs a HTTP POST request using the default client and, when
// complete, attempts to convert the response body into the datatype given by
// dst (a pointer to a struct, map or []byte array).
func Post(dest interface{}, uri string, data url.Values) error {
  return DefaultClient.Post(dest, uri, data)
}

// Put performs a HTTP PUT request using the default client and, when complete,
// attempts to convert the response body into the datatype given by dst (a
// pointer to a struct, map or []byte array).
func Put(dest interface{}, uri string, data url.Values) error {
  return DefaultClient.Put(dest, uri, data)
}

// Delete performs a HTTP DELETE request using the default client and, when
// complete, attempts to convert the response body into the datatype given by
// dst (a pointer to a struct, map or []byte array).
func Delete(dest interface{}, uri string, data url.Values) error {
  return DefaultClient.Delete(dest, uri, data)
}

// PostMultipart performs a HTTP POST multipart request using the default
// client and, when complete, attempts to convert the response body into the
// datatype given by dst (a pointer to a struct, map or []byte array).
func PostMultipart(dest interface{}, uri string, data *MultipartBody) error {
  return DefaultClient.PostMultipart(dest, uri, data)
}

// PutMultipart performs a HTTP PUT multipart request using the default client
// and, when complete, attempts to convert the response body into the datatype
// given by dst (a pointer to a struct, map or []byte array).
func PutMultipart(dest interface{}, uri string, data *MultipartBody) error {
  return DefaultClient.PutMultipart(dest, uri, data)
}
