package demo

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/zhenyangze/go-admin-build/goadmin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLoginAndGridRender(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(app.Handler())
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	resp, err := client.Get(server.URL + "/admin/articles")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}
	if !strings.Contains(body, "Building Admin DSLs in Go") {
		t.Fatalf("expected seeded article in grid, body=%s", body)
	}
}

func TestCreateArticleAndRenderTree(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(app.Handler())
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	formResp, err := client.Get(server.URL + "/admin/articles/new")
	if err != nil {
		t.Fatal(err)
	}
	formBody := readBody(t, formResp)
	csrf := extractCSRF(t, formBody)

	values := url.Values{
		"_csrf":      {csrf},
		"Title":      {"Fresh integration test article"},
		"Summary":    {"Created through the admin form in an httptest server."},
		"Status":     {"published"},
		"CategoryID": {"1"},
	}
	resp, err := client.PostForm(server.URL+"/admin/articles", values)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected redirect-followed success, got %d", resp.StatusCode)
	}

	gridResp, err := client.Get(server.URL + "/admin/articles")
	if err != nil {
		t.Fatal(err)
	}
	gridBody := readBody(t, gridResp)
	if !strings.Contains(gridBody, "Fresh integration test article") {
		t.Fatalf("expected created article in grid, body=%s", gridBody)
	}

	treeResp, err := client.Get(server.URL + "/admin/categories/tree")
	if err != nil {
		t.Fatal(err)
	}
	treeBody := readBody(t, treeResp)
	if !strings.Contains(treeBody, "Getting Started") {
		t.Fatalf("expected seeded category in tree, body=%s", treeBody)
	}
}

func newTestApp(t *testing.T) *goadmin.App {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	app, err := BuildWithDB(db)
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func newClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar}
}

func login(t *testing.T, client *http.Client, base string) {
	t.Helper()
	resp, err := client.PostForm(base+"/admin/login", url.Values{
		"username": {"admin"},
		"password": {"admin"},
	})
	if err != nil {
		t.Fatal(err)
	}
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%s", resp.StatusCode, body)
	}
}

func extractCSRF(t *testing.T, body string) string {
	t.Helper()
	re := regexp.MustCompile(`name="_csrf" value="([^"]+)"`)
	matches := re.FindStringSubmatch(body)
	if len(matches) != 2 {
		t.Fatalf("csrf token not found in body: %s", body)
	}
	return matches[1]
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
