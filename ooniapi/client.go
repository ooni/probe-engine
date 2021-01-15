package ooniapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// This package returns the following errors.
var (
	ErrFailed    = errors.New("ooniapi: API failed")
	ErrIsZero    = errors.New("ooniapi: reflection: nil pointer")
	ErrNoSupport = errors.New("ooniapi: reflection: cast not supported")
	ErrNotStruct = errors.New("ooniapi: reflection: not a struct")
	ErrNotString = errors.New("ooniapi: reflection: not a string")
)

type requestType interface {
	isRequest()
}

type responseType interface {
	isResponse()
}

// Client is a client for the OONI API.
type Client struct {
	Authorization string
	BaseURL       string
	HTTPClient    *http.Client
	UserAgent     string
}

func (c Client) urlpath(urlpath string, in requestType) (string, error) {
	valueinfo := reflect.ValueOf(in)
	if valueinfo.Kind() == reflect.Ptr {
		valueinfo = valueinfo.Elem()
		if valueinfo.IsZero() {
			return "", ErrIsZero
		}
	}
	typeinfo := valueinfo.Type()
	if typeinfo.Kind() != reflect.Struct {
		return "", ErrNotStruct
	}
	for idx := 0; idx < typeinfo.NumField(); idx++ {
		ft := typeinfo.Field(idx)
		if tag := ft.Tag.Get("urlpath"); tag != "" {
			fv := valueinfo.Field(idx)
			if fv.Kind() != reflect.String {
				return "", ErrNotString
			}
			tag = fmt.Sprintf("{%s}", tag)
			v := fv.Interface().(string)
			urlpath = strings.ReplaceAll(urlpath, tag, v)
		}
	}
	return urlpath, nil
}

func (c Client) query(in requestType) (string, error) {
	valueinfo := reflect.ValueOf(in)
	if valueinfo.Kind() == reflect.Ptr {
		valueinfo = valueinfo.Elem()
		if valueinfo.IsZero() {
			return "", ErrIsZero
		}
	}
	typeinfo := valueinfo.Type()
	if typeinfo.Kind() != reflect.Struct {
		return "", ErrNotStruct
	}
	values := url.Values{}
	for idx := 0; idx < typeinfo.NumField(); idx++ {
		ft := typeinfo.Field(idx)
		if tag := ft.Tag.Get("query"); tag != "" {
			fv := valueinfo.Field(idx)
			switch fv.Kind() {
			case reflect.String:
				values.Add(tag, fv.Interface().(string))
			case reflect.Int64:
				values.Add(tag, fmt.Sprintf("%d", fv.Interface().(int64)))
			case reflect.Bool:
				values.Add(tag, fmt.Sprintf("%v", fv.Interface().(bool)))
			default:
				return "", ErrNoSupport
			}
		}
	}
	return values.Encode(), nil
}

func (c Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.Authorization)
	req.Header.Set("User-Agent", c.UserAgent)
}

func (c Client) newRequestGET(ctx context.Context, urlpath string, in requestType) (*http.Request, error) {
	URL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	urlpath, err = c.urlpath(urlpath, in)
	if err != nil {
		return nil, err
	}
	URL.Path = urlpath
	query, err := c.query(in)
	if err != nil {
		return nil, err
	}
	URL.RawQuery = query
	req, err := http.NewRequestWithContext(ctx, "GET", URL.String(), nil)
	if err != nil {
		return nil, err
	}
	c.setCommonHeaders(req)
	return req, nil
}

func (c Client) newRequestPOST(ctx context.Context, urlpath string, in requestType) (*http.Request, error) {
	URL, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	urlpath, err = c.urlpath(urlpath, in)
	if err != nil {
		return nil, err
	}
	URL.Path = urlpath
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", URL.String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	c.setCommonHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c Client) unmarshal(resp *http.Response, out interface{}) error {
	if resp.StatusCode != 200 {
		return ErrFailed
	}
	reader := io.LimitReader(resp.Body, 4<<20)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

type apispec struct {
	Method  string
	URLPath string
	In      requestType
	Out     responseType
}

func (c Client) api(ctx context.Context, desc apispec) error {
	methodfunc := map[string]func(context.Context, string, requestType) (*http.Request, error){
		"GET":  c.newRequestGET,
		"POST": c.newRequestPOST,
	}
	fn := methodfunc[desc.Method]
	req, err := fn(ctx, desc.URLPath, desc.In)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return c.unmarshal(resp, desc.Out)
}
