package dash

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/experiment/handler"
	"github.com/ooni/probe-engine/netx/trace"
)

const (
	softwareName    = "dash-client-go-test"
	softwareVersion = "0.0.1"
)

func TestClientNegotiate(t *testing.T) {
	t.Run("json.Marshal failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.JSONMarshal = func(v interface{}) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		_, err := clnt.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.NewRequest failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPNewRequest = func(
			method string, url string, body io.Reader,
		) (*http.Request, error) {
			return nil, errors.New("Mocked error")
		}
		_, err := clnt.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.Client.Do failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("Mocked error")
		}
		_, err := clnt.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Non successful response", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
			}, nil
		}
		_, err := clnt.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("ioutil.ReadAll failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		clnt.deps.IOUtilReadAll = func(r io.Reader) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		_, err := clnt.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("json.Unmarshal failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		_, err := clnt.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Invalid JSON or not authorized", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("{}")),
			}, nil
		}
		_, err := clnt.negotiate(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Success", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(strings.NewReader(`{
					"Authorization": "0xdeadbeef",
					"Unchoked": 1
				}`)),
			}, nil
		}
		_, err := clnt.negotiate(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestClientDownload(t *testing.T) {
	t.Run("http.NewRequest failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPNewRequest = func(
			method string, url string, body io.Reader,
		) (*http.Request, error) {
			return nil, errors.New("Mocked error")
		}
		current := new(clientResults)
		err := clnt.download(context.Background(), "abc", current)
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.Client.Do failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("Mocked error")
		}
		current := new(clientResults)
		err := clnt.download(context.Background(), "abc", current)
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Non successful response", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
			}, nil
		}
		current := new(clientResults)
		err := clnt.download(context.Background(), "abc", current)
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("ioutil.ReadAll failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		clnt.deps.IOUtilReadAll = func(r io.Reader) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		current := new(clientResults)
		err := clnt.download(context.Background(), "abc", current)
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Success", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		current := new(clientResults)
		err := clnt.download(context.Background(), "abc", current)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestClientCollect(t *testing.T) {
	t.Run("json.Marshal failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.JSONMarshal = func(v interface{}) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		err := clnt.collect(context.Background(), "abc")
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.NewRequest failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPNewRequest = func(
			method string, url string, body io.Reader,
		) (*http.Request, error) {
			return nil, errors.New("Mocked error")
		}
		err := clnt.collect(context.Background(), "abc")
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("http.Client.Do failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("Mocked error")
		}
		err := clnt.collect(context.Background(), "abc")
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Non successful response", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
			}, nil
		}
		err := clnt.collect(context.Background(), "abc")
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("ioutil.ReadAll failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		clnt.deps.IOUtilReadAll = func(r io.Reader) ([]byte, error) {
			return nil, errors.New("Mocked error")
		}
		err := clnt.collect(context.Background(), "abc")
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("json.Unmarshal failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}
		err := clnt.collect(context.Background(), "abc")
		if err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("Success", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.HTTPClientDo = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("[]")),
			}, nil
		}
		err := clnt.collect(context.Background(), "abc")
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestClientLoop(t *testing.T) {
	t.Run("negotiate failure", func(t *testing.T) {
		ch := make(chan clientResults)
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.Negotiate = func(ctx context.Context) (negotiateResponse, error) {
			return negotiateResponse{}, errors.New("Mocked error")
		}
		clnt.loop(context.Background(), ch)
		if clnt.err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("download failure", func(t *testing.T) {
		ch := make(chan clientResults)
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.Negotiate = func(ctx context.Context) (negotiateResponse, error) {
			return negotiateResponse{}, nil
		}
		clnt.deps.Download = func(
			ctx context.Context, authorization string, current *clientResults,
		) error {
			return errors.New("Mocked error")
		}
		clnt.loop(context.Background(), ch)
		if clnt.err == nil {
			t.Fatal("Expected an error here")
		}
	})

	t.Run("collect failure", func(t *testing.T) {
		ch := make(chan clientResults)
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.Negotiate = func(ctx context.Context) (negotiateResponse, error) {
			return negotiateResponse{}, nil
		}
		clnt.deps.Download = func(
			ctx context.Context, authorization string, current *clientResults,
		) error {
			return nil
		}
		clnt.deps.Collect = func(ctx context.Context, authorization string) error {
			return errors.New("Mocked error")
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range ch {
				// drain channel
			}
		}()
		clnt.loop(context.Background(), ch)
		if clnt.err == nil {
			t.Fatal("Expected an error here")
		}
		wg.Wait() // make sure we really terminate
	})
}

func TestClientStartDownload(t *testing.T) {
	t.Run("mlabns failure", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.Locate = func(ctx context.Context) (string, error) {
			return "", errors.New("Mocked error")
		}
		ch, err := clnt.StartDownload(context.Background())
		if err == nil {
			t.Fatal("Expected an error here")
		}
		if ch != nil {
			t.Fatal("Expected nil channel here")
		}
	})

	t.Run("common case", func(t *testing.T) {
		clnt := newClient(http.DefaultClient, new(trace.Saver), log.Log,
			handler.NewPrinterCallbacks(log.Log), softwareName, softwareVersion)
		clnt.deps.Loop = func(ctx context.Context, ch chan<- clientResults) {
			close(ch)
		}
		ch, err := clnt.StartDownload(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		for range ch {
			// drain channel
		}
	})
}
