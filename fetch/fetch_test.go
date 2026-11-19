package fetch

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchUrlSimple(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	resp, err := FetchUrlSimple(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "hello" {
		t.Fatalf("body = %q", body)
	}
}

func TestFetchUrlPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"a":1}` {
			t.Fatalf("body = %q", body)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	resp, err := FetchUrl(srv.URL, PostJsonConfig([]byte(`{"a":1}`)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
}

func TestFetchUrlNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	resp, err := FetchUrlSimple(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
}
