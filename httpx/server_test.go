package httpx

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

type mockRouter struct {
	answer func(writer http.ResponseWriter, request *http.Request)
}

func (m mockRouter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	m.answer(writer, request)
}

func TestServer(t *testing.T) {
	tests := []struct {
		name            string
		addr            string
		wantCtxValue    string
		wantRespCode    int
		wantRespContent string
	}{
		{name: "Happy case", addr: ":49152", wantCtxValue: "test", wantRespCode: http.StatusOK, wantRespContent: "passed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			onStartUpCalled := false
			router := &mockRouter{answer: func(writer http.ResponseWriter, request *http.Request) {
				if request.Context().Value(tt.name) != tt.wantCtxValue {
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}
				writer.WriteHeader(tt.wantRespCode)
				_, err := writer.Write([]byte(tt.wantRespContent))
				if err != nil {
					t.Fatal(err)
				}
			}}
			srv := &Server{
				onStartUp: func() {
					onStartUpCalled = true
				},
				middlewareFunc: func(req *http.Request) *http.Request {
					return req.WithContext(context.WithValue(req.Context(), tt.name, tt.wantCtxValue)) //nolint:staticcheck
				},
				Router: router,
			}
			srv.Server = http.Server{
				Addr:    tt.addr,
				Handler: srv,
			}
			go func() {
				time.Sleep(1 * time.Second)
				_ = srv.Shutdown(context.Background())
			}()

			startServerWithWait(t, srv, 1)

			if !onStartUpCalled {
				t.Fatalf("TestServer onStartUp was not called")
			}
			resp := sendRequest(t, http.MethodGet, "http://localhost"+tt.addr+"/status", nil)
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			if resp.StatusCode != tt.wantRespCode {
				t.Fatalf("TestServer status code %d, want %d", resp.StatusCode, tt.wantRespCode)
			}
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			content := string(b)
			if content != tt.wantRespContent {
				t.Fatalf("TestServer content '%s', want '%s'", content, tt.wantRespContent)
			}
		})
	}
}

func startServerWithWait(t *testing.T, srv *Server, waitSecs int) {
	srv.Start()
	msWaited := 0
	for {
		if srv.status == StatusStarted {
			break
		}
		if time.Duration(msWaited) == time.Duration(waitSecs)*1000 {
			t.Fatalf("Server did not start after %d seconds", waitSecs)
		}
		time.Sleep(1 * time.Millisecond)
		msWaited = msWaited + 1
	}
}

func sendRequest(t *testing.T, method string, url string, body []byte) *http.Response {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}
