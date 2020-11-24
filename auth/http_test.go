package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voc/srtrelay/stream"
)

func serverMock() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ok"))
	})
	handler.HandleFunc("/unauthorized", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
	})

	srv := httptest.NewServer(handler)

	return srv
}

func Test_httpAuth_Authenticate(t *testing.T) {
	srv := serverMock()
	defer srv.Close()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"AuthOk", "/ok", true},
		{"AuthFail", "/unauthorized", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			auth := NewHTTPAuth(HTTPAuthConfig{
				URL: srv.URL + tt.url,
			})

			streamid := stream.StreamID{}

			if got := auth.Authenticate(streamid); got != tt.want {
				t.Errorf("httpAuth.Authenticate() = %v, want %v", got, tt.want)
			}
		})
	}
}
