package system

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestGitHubPublicProfileAndCache(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "" {
			t.Error("public profile must use unauthenticated REST GET")
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/users/demo":
			w.Write([]byte(`{"login":"demo","name":"Demo","public_repos":1,"email":"not-returned@example.test"}`))
		case strings.HasSuffix(r.URL.Path, "/repos"):
			w.Write([]byte(`[{"name":"public","full_name":"demo/public","private":false},{"name":"secret","full_name":"demo/secret","private":true}]`))
		case r.URL.Path == "/search/commits":
			if r.URL.Query().Get("q") != "author:demo is:public" {
				t.Error("search must constrain public visibility")
			}
			w.Write([]byte(`{"items":[{"sha":"abcd","repository":{"full_name":"demo/public","private":false},"commit":{"message":"Fix\nMore detail","author":{"date":"2026-01-01T00:00:00Z","email":"secret@example.test"}}},{"sha":"secret","repository":{"full_name":"demo/secret","private":true}}]}`))
		case strings.HasSuffix(r.URL.Path, "/events/public"):
			w.Write([]byte(`[{"id":"1","type":"PushEvent","public":true,"repo":{"name":"demo/public"}},{"id":"2","type":"PushEvent","public":false,"repo":{"name":"demo/secret"}}]`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	reader := &githubReader{base: server.URL, client: server.Client(), cache: map[string]githubCacheItem{}}
	result, err := reader.read(context.Background(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Repositories) != 1 || len(result.Commits) != 1 || len(result.Events) != 1 {
		t.Fatalf("public data filtering failed: %+v", result)
	}
	if result.Commits[0].Message != "Fix" {
		t.Fatal("commit message should use first line")
	}
	raw, _ := json.Marshal(result)
	if strings.Contains(string(raw), "secret") || strings.Contains(string(raw), "@example.test") {
		t.Fatal("private details leaked")
	}
	initial := calls.Load()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := reader.read(context.Background(), "demo"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != initial {
		t.Fatal("cached reads should not repeat upstream requests")
	}

}
func TestGitHubValidationAndPartialErrors(t *testing.T) {
	for _, name := range []string{"../admin", "https://github.com/demo", "a--b", "-abc", "abc-", strings.Repeat("a", 40)} {
		if ValidGitHubUsername(name) {
			t.Errorf("accepted invalid login %q", name)
		}
	}
	if !ValidGitHubUsername("a-B1") {
		t.Fatal("valid login rejected")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/demo" {
			w.Write([]byte(`{"login":"demo"}`))
			return
		}
		w.WriteHeader(429)
		w.Write([]byte(`{"message":"sensitive-upstream-detail"}`))
	}))
	defer server.Close()
	reader := &githubReader{base: server.URL, client: server.Client(), cache: map[string]githubCacheItem{}}
	result, err := reader.read(context.Background(), "demo")
	if err != nil || len(result.Warnings) != 3 {
		t.Fatalf("partial errors must preserve profile: %v %+v", err, result)
	}
	raw, _ := json.Marshal(result)
	if strings.Contains(string(raw), "sensitive-upstream-detail") {
		t.Fatal("raw provider error leaked")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = reader.read(ctx, "other"); err == nil {
		t.Fatal("canceled request must return promptly")
	}
}
