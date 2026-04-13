package demo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zhenyangze/go-admin-build/examples/demo/models"
	"github.com/zhenyangze/goadmin"
	"github.com/zhenyangze/goadmin/auth"
	"github.com/zhenyangze/goadmin/form"
	"github.com/zhenyangze/goadmin/grid"
	"github.com/zhenyangze/goadmin/show"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLoginAndGridRender(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(newGinHandler(app))
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
	if !strings.Contains(body, "Category") {
		t.Fatalf("expected custom row action label in article grid, body=%s", body)
	}
}

func TestDashboardPanelsRender(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(newGinHandler(app))
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	resp, err := client.Get(server.URL + "/admin")
	if err != nil {
		t.Fatal(err)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "Quick Access") ||
		!strings.Contains(body, "Recent Articles") ||
		!strings.Contains(body, "Module Guide") ||
		!strings.Contains(body, "Suggested Walkthrough") ||
		!strings.Contains(body, "Framework Coverage") ||
		!strings.Contains(body, "Demo Helpers") ||
		!strings.Contains(body, "Capability Matrix") {
		t.Fatalf("expected dashboard panels, body=%s", body)
	}
	if !strings.Contains(body, "Create Article") ||
		!strings.Contains(body, "Building Admin DSLs in Go") ||
		!strings.Contains(body, "What this demo proves") ||
		!strings.Contains(body, "Tickets") ||
		!strings.Contains(body, "Fresh reset") ||
		!strings.Contains(body, "Read-only pages") ||
		!strings.Contains(body, "Reports") {
		t.Fatalf("expected dashboard actions/items, body=%s", body)
	}
}

func TestDemoCenterAndExportRoute(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(newGinHandler(app))
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	resp, err := client.Get(server.URL + "/admin/demo-center")
	if err != nil {
		t.Fatal(err)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "Demo Center") || !strings.Contains(body, "Fresh run command") || !strings.Contains(body, "Export JSON") {
		t.Fatalf("expected demo center helper page, body=%s", body)
	}

	exportResp, err := client.Get(server.URL + "/admin/demo-export.json")
	if err != nil {
		t.Fatal(err)
	}
	exportBody := readBody(t, exportResp)
	if !strings.Contains(exportBody, "\"articles\"") || !strings.Contains(exportBody, "\"projects\"") || !strings.Contains(exportBody, "\"reports\"") {
		t.Fatalf("expected export payload, body=%s", exportBody)
	}
}

func TestFormValidationFeedback(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(newGinHandler(app))
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	formResp, err := client.Get(server.URL + "/admin/articles/new")
	if err != nil {
		t.Fatal(err)
	}
	csrf := extractCSRF(t, readBody(t, formResp))
	resp, err := postMultipartForm(t, client, server.URL+"/admin/articles", url.Values{
		"_csrf":       {csrf},
		"Title":       {""},
		"Summary":     {"missing title"},
		"Status":      {"published"},
		"VisibleFrom": {"2026-04-20"},
		"VisibleTo":   {"2026-04-10"},
		"CategoryID":  {"1"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 validation response, got %d body=%s", resp.StatusCode, body)
	}
	if !strings.Contains(body, "Please correct the highlighted fields.") ||
		!strings.Contains(body, "This field is required.") ||
		!strings.Contains(body, "Start date must be before or equal to end date.") {
		t.Fatalf("expected inline validation feedback, body=%s", body)
	}
}

func TestCreateArticleAndRenderTree(t *testing.T) {
	app, db := newTestFixture(t)
	server := httptest.NewServer(newGinHandler(app))
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
		"_csrf":          {csrf},
		"Title":          {"Fresh integration test article"},
		"Summary":        {"Created through the admin form in an httptest server."},
		"Status":         {"published"},
		"Featured":       {"1"},
		"Tags":           {"go,gin,sqlite"},
		"PublishedAt":    {"2026-04-10T12:30"},
		"VisibleFrom":    {"2026-04-01"},
		"VisibleTo":      {"2026-04-30"},
		"CategoryID":     {"1"},
		"FAQ.0.Question": {"What is this?"},
		"FAQ.0.Answer":   {"An integration test article."},
		"FAQ.1.Question": {"Why nested?"},
		"FAQ.1.Answer":   {"To verify repeater form support."},
		"Links.0.Label":  {"Home"},
		"Links.0.URL":    {"https://example.com/home"},
		"Links.1.Label":  {"Docs"},
		"Links.1.URL":    {"https://example.com/docs"},
	}
	resp, err := postMultipartForm(t, client, server.URL+"/admin/articles", values, []testUploadFile{
		{Field: "Image", Name: "cover.png", Content: []byte("fake-image-content")},
		{Field: "Gallery", Name: "gallery-a.png", Content: []byte("gallery-a")},
		{Field: "Gallery", Name: "gallery-b.png", Content: []byte("gallery-b")},
	})
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
	if !strings.Contains(gridBody, "12:30") {
		t.Fatalf("expected published_at in grid, body=%s", gridBody)
	}
	if !strings.Contains(gridBody, "gin") || !strings.Contains(gridBody, "2026-04-01 → 2026-04-30") {
		t.Fatalf("expected tags/visible range in grid, body=%s", gridBody)
	}
	if !strings.Contains(gridBody, "/admin/uploads/") {
		t.Fatalf("expected uploaded image link in grid, body=%s", gridBody)
	}
	if !strings.Contains(gridBody, "2 files") {
		t.Fatalf("expected gallery count in grid, body=%s", gridBody)
	}
	if !strings.Contains(gridBody, "2 items") {
		t.Fatalf("expected faq count in grid, body=%s", gridBody)
	}
	if !strings.Contains(gridBody, "2 links") {
		t.Fatalf("expected links count in grid, body=%s", gridBody)
	}
	if !strings.Contains(gridBody, "badge badge-ok") {
		t.Fatalf("expected tags rendered as badges, body=%s", gridBody)
	}

	showResp, err := client.Get(server.URL + "/admin/articles/3")
	if err != nil {
		t.Fatal(err)
	}
	showBody := readBody(t, showResp)
	if !strings.Contains(showBody, "true") ||
		!strings.Contains(showBody, "2026-04-10 12:30") ||
		!strings.Contains(showBody, "gin") ||
		!strings.Contains(showBody, "/admin/uploads/") ||
		!strings.Contains(showBody, "2026-04-01") ||
		!strings.Contains(showBody, "2026-04-30") ||
		!strings.Contains(showBody, "What is this?") ||
		!strings.Contains(showBody, "To verify repeater form support.") ||
		!strings.Contains(showBody, "https://example.com/home") {
		t.Fatalf("expected featured/published fields in show page, body=%s", showBody)
	}
	if strings.Contains(showBody, `[{&quot;Question&quot;`) {
		t.Fatalf("expected faq rendered as formatted list, body=%s", showBody)
	}
	uploadURL := extractUploadURL(t, showBody)
	fileResp, err := client.Get(server.URL + uploadURL)
	if err != nil {
		t.Fatal(err)
	}
	fileBody := readBody(t, fileResp)
	if fileBody != "fake-image-content" {
		t.Fatalf("expected uploaded file body, got %q", fileBody)
	}
	var articleRecord models.Article
	if err := db.Preload("FAQs").Preload("Links").Where("title = ?", "Fresh integration test article").First(&articleRecord).Error; err != nil {
		t.Fatal(err)
	}
	if len(strings.Split(articleRecord.Gallery, ",")) != 2 {
		t.Fatalf("expected 2 gallery files, got %q", articleRecord.Gallery)
	}
	if len(articleRecord.FAQs) != 2 {
		t.Fatalf("expected 2 faq child records, got %d", len(articleRecord.FAQs))
	}
	if len(articleRecord.Links) != 2 {
		t.Fatalf("expected 2 link child records, got %d", len(articleRecord.Links))
	}

	editResp, err := client.Get(server.URL + "/admin/articles/" + formatUint(articleRecord.ID) + "/edit")
	if err != nil {
		t.Fatal(err)
	}
	editCSRF := extractCSRF(t, readBody(t, editResp))
	_, err = postMultipartForm(t, client, server.URL+"/admin/articles/"+formatUint(articleRecord.ID), url.Values{
		"_csrf":          {editCSRF},
		"_method":        {"PUT"},
		"Title":          {"Fresh integration test article"},
		"Summary":        {"Updated nested relation content."},
		"Status":         {"published"},
		"Featured":       {"1"},
		"Tags":           {"go,gin,sqlite"},
		"PublishedAt":    {"2026-04-10T12:30"},
		"VisibleFrom":    {"2026-04-01"},
		"VisibleTo":      {"2026-04-30"},
		"CategoryID":     {"1"},
		"FAQ.0.Question": {"Updated question"},
		"FAQ.0.Answer":   {"Updated answer"},
		"Links.0.ID":     {formatUint(articleRecord.Links[0].ID)},
		"Links.0.Label":  {"Home Updated"},
		"Links.0.URL":    {"https://example.com/home-updated"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Preload("FAQs").Preload("Links").First(&articleRecord, articleRecord.ID).Error; err != nil {
		t.Fatal(err)
	}
	if len(articleRecord.FAQs) != 1 || articleRecord.FAQs[0].Question != "Updated question" {
		t.Fatalf("expected faq relation replacement, got %+v", articleRecord.FAQs)
	}
	if len(articleRecord.Links) != 2 || articleRecord.Links[0].Label != "Home Updated" || articleRecord.Links[1].Label != "Docs" {
		t.Fatalf("expected link relation merge behavior, got %+v", articleRecord.Links)
	}
	var faqTotal int64
	if err := db.Unscoped().Model(&models.ArticleFAQ{}).Where("article_id = ?", articleRecord.ID).Count(&faqTotal).Error; err != nil {
		t.Fatal(err)
	}
	if faqTotal < 3 {
		t.Fatalf("expected soft-deleted faq history after replacement, total=%d", faqTotal)
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

func TestUploadValidationAndCleanup(t *testing.T) {
	app, db := newTestFixture(t)
	server := httptest.NewServer(newGinHandler(app))
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	formResp, err := client.Get(server.URL + "/admin/articles/new")
	if err != nil {
		t.Fatal(err)
	}
	csrf := extractCSRF(t, readBody(t, formResp))
	badResp, err := postMultipartForm(t, client, server.URL+"/admin/articles", url.Values{
		"_csrf":      {csrf},
		"Title":      {"Bad upload"},
		"Summary":    {"should fail"},
		"Status":     {"draft"},
		"CategoryID": {"1"},
	}, []testUploadFile{{Field: "Image", Name: "bad.txt", Content: []byte("bad")}})
	if err != nil {
		t.Fatal(err)
	}
	badBody := readBody(t, badResp)
	if badResp.StatusCode != http.StatusBadRequest || !strings.Contains(badBody, "not allowed") {
		t.Fatalf("expected invalid upload rejection, status=%d body=%s", badResp.StatusCode, badBody)
	}

	formResp, err = client.Get(server.URL + "/admin/articles/new")
	if err != nil {
		t.Fatal(err)
	}
	csrf = extractCSRF(t, readBody(t, formResp))
	_, err = postMultipartForm(t, client, server.URL+"/admin/articles", url.Values{
		"_csrf":      {csrf},
		"Title":      {"Cleanup article"},
		"Summary":    {"upload cleanup"},
		"Status":     {"published"},
		"CategoryID": {"1"},
	}, []testUploadFile{
		{Field: "Image", Name: "old.png", Content: []byte("old-image")},
		{Field: "Gallery", Name: "old-a.png", Content: []byte("old-a")},
		{Field: "Gallery", Name: "old-b.png", Content: []byte("old-b")},
	})
	if err != nil {
		t.Fatal(err)
	}

	var article models.Article
	if err := db.Where("title = ?", "Cleanup article").First(&article).Error; err != nil {
		t.Fatal(err)
	}
	oldImage := article.Image
	oldGallery := append([]string(nil), strings.Split(article.Gallery, ",")...)

	editResp, err := client.Get(server.URL + "/admin/articles/" + formatUint(article.ID) + "/edit")
	if err != nil {
		t.Fatal(err)
	}
	editCSRF := extractCSRF(t, readBody(t, editResp))
	_, err = postMultipartForm(t, client, server.URL+"/admin/articles/"+formatUint(article.ID), url.Values{
		"_csrf":      {editCSRF},
		"_method":    {"PUT"},
		"Title":      {"Cleanup article"},
		"Summary":    {"upload cleanup updated"},
		"Status":     {"published"},
		"CategoryID": {"1"},
	}, []testUploadFile{
		{Field: "Image", Name: "new.png", Content: []byte("new-image")},
		{Field: "Gallery", Name: "new-a.png", Content: []byte("new-a")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.First(&article, article.ID).Error; err != nil {
		t.Fatal(err)
	}
	if article.Image == oldImage || article.Image == "" {
		t.Fatalf("expected image replacement, old=%q new=%q", oldImage, article.Image)
	}
	assertStatusCode(t, client, server.URL+oldImage, http.StatusNotFound)
	for _, path := range oldGallery {
		if strings.TrimSpace(path) != "" {
			assertStatusCode(t, client, server.URL+path, http.StatusNotFound)
		}
	}

	editResp, err = client.Get(server.URL + "/admin/articles/" + formatUint(article.ID) + "/edit")
	if err != nil {
		t.Fatal(err)
	}
	deleteCSRF := extractCSRF(t, readBody(t, editResp))
	delResp, err := client.PostForm(server.URL+"/admin/articles/"+formatUint(article.ID)+"/delete", url.Values{"_csrf": {deleteCSRF}})
	if err != nil {
		t.Fatal(err)
	}
	_ = readBody(t, delResp)
	assertStatusCode(t, client, server.URL+article.Image, http.StatusNotFound)
}

func TestManageAuthResources(t *testing.T) {
	app, db := newTestFixture(t)
	server := httptest.NewServer(newGinHandler(app))
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	createResource(t, client, server.URL, "/admin/permissions/new", "/admin/permissions", url.Values{
		"Name": {"Reports"},
		"Slug": {"reports.manage"},
		"Path": {"/admin/reports*"},
	})

	var permission auth.Permission
	if err := db.Where("slug = ?", "reports.manage").First(&permission).Error; err != nil {
		t.Fatal(err)
	}

	createResource(t, client, server.URL, "/admin/menus/new", "/admin/menus", url.Values{
		"ParentID":       {"0"},
		"Order":          {"50"},
		"Title":          {"Reports"},
		"Icon":           {"chart-bar"},
		"URI":            {"reports"},
		"PermissionSlug": {permission.Slug},
	})

	var menu auth.Menu
	if err := db.Where("title = ?", "Reports").First(&menu).Error; err != nil {
		t.Fatal(err)
	}

	createResource(t, client, server.URL, "/admin/roles/new", "/admin/roles", url.Values{
		"Name":          {"Reporter"},
		"Slug":          {"reporter"},
		"PermissionIDs": {formatUint(permission.ID)},
		"MenuIDs":       {formatUint(menu.ID)},
	})

	var role auth.Role
	if err := db.Where("slug = ?", "reporter").First(&role).Error; err != nil {
		t.Fatal(err)
	}

	createResource(t, client, server.URL, "/admin/users/new", "/admin/users", url.Values{
		"Username": {"reporter"},
		"Name":     {"Reporter User"},
		"Password": {"secret123"},
		"RoleIDs":  {formatUint(role.ID)},
	})

	var user auth.User
	if err := db.Where("username = ?", "reporter").First(&user).Error; err != nil {
		t.Fatal(err)
	}

	roleResp, err := client.Get(server.URL + "/admin/roles/" + formatUint(role.ID))
	if err != nil {
		t.Fatal(err)
	}
	roleBody := readBody(t, roleResp)
	if !strings.Contains(roleBody, "Reports") || !strings.Contains(roleBody, "Reporter") {
		t.Fatalf("expected role detail to include linked menu and role, body=%s", roleBody)
	}

	userResp, err := client.Get(server.URL + "/admin/users/" + formatUint(user.ID))
	if err != nil {
		t.Fatal(err)
	}
	userBody := readBody(t, userResp)
	if !strings.Contains(userBody, "Reporter User") || !strings.Contains(userBody, "Reporter") {
		t.Fatalf("expected user detail to include linked role, body=%s", userBody)
	}

	treeResp, err := client.Get(server.URL + "/admin/menus/tree")
	if err != nil {
		t.Fatal(err)
	}
	treeBody := readBody(t, treeResp)
	if !strings.Contains(treeBody, "Reports") {
		t.Fatalf("expected menu tree to include created menu, body=%s", treeBody)
	}
}

func TestGridActionFlagsAndCreateLabel(t *testing.T) {
	app, _ := newTestFixture(t)
	app.Register(goadmin.Resource{
		Name:       "flags",
		Path:       "flags",
		Title:      "Flags",
		Repository: stubRepository{items: []any{stubRecord{ID: 1, Name: "One"}}},
		BuildGrid: func(b *grid.Builder) {
			b.CreateLabel = "New Widget"
			b.DisableView = true
			b.DisableDelete = true
			b.Column("ID", "ID")
			b.Column("Name", "Name")
		},
		BuildForm: func(b *form.Builder) {
			b.Text("Name", "Name")
			b.CancelBackURL = "/admin"
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
		},
	})

	server := httptest.NewServer(newGinHandler(app))
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	resp, err := client.Get(server.URL + "/admin/flags")
	if err != nil {
		t.Fatal(err)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "New Widget") {
		t.Fatalf("expected custom create label, body=%s", body)
	}
	if strings.Contains(body, ">View<") || strings.Contains(body, ">Delete<") {
		t.Fatalf("expected disabled row actions to be hidden, body=%s", body)
	}
	if !strings.Contains(body, ">Edit<") {
		t.Fatalf("expected edit action to remain visible, body=%s", body)
	}
}

func TestDemoCustomGridActions(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(newGinHandler(app))
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	articleResp, err := client.Get(server.URL + "/admin/articles")
	if err != nil {
		t.Fatal(err)
	}
	articleBody := readBody(t, articleResp)
	if !strings.Contains(articleBody, "Published") {
		t.Fatalf("expected articles page action, body=%s", articleBody)
	}
	if !strings.Contains(articleBody, ">Category<") {
		t.Fatalf("expected articles row action, body=%s", articleBody)
	}

	categoryResp, err := client.Get(server.URL + "/admin/categories")
	if err != nil {
		t.Fatal(err)
	}
	categoryBody := readBody(t, categoryResp)
	if !strings.Contains(categoryBody, "Tree View") {
		t.Fatalf("expected categories tree page action, body=%s", categoryBody)
	}

	menuResp, err := client.Get(server.URL + "/admin/menus")
	if err != nil {
		t.Fatal(err)
	}
	menuBody := readBody(t, menuResp)
	if !strings.Contains(menuBody, "Tree View") {
		t.Fatalf("expected menus tree page action, body=%s", menuBody)
	}
}

func TestProjectsModuleRenderAndCreate(t *testing.T) {
	app, db := newTestFixture(t)
	server := httptest.NewServer(newGinHandler(app))
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	resp, err := client.Get(server.URL + "/admin/projects")
	if err != nil {
		t.Fatal(err)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "Framework Launch") || !strings.Contains(body, "2 items") {
		t.Fatalf("expected seeded projects grid, body=%s", body)
	}

	formResp, err := client.Get(server.URL + "/admin/projects/new")
	if err != nil {
		t.Fatal(err)
	}
	csrf := extractCSRF(t, readBody(t, formResp))
	postResp, err := postMultipartForm(t, client, server.URL+"/admin/projects", url.Values{
		"_csrf":                 {csrf},
		"Name":                  {"Beta Rollout"},
		"Owner":                 {"Carol"},
		"Status":                {"active"},
		"Budget":                {"150000"},
		"StartsAt":              {"2026-04-15"},
		"EndsAt":                {"2026-06-01"},
		"Description":           {"Roll out the beta environment and internal feedback loop."},
		"Milestones.0.Title":    {"Spec freeze"},
		"Milestones.0.Deadline": {"2026-04-20"},
		"Milestones.1.Title":    {"Internal beta"},
		"Milestones.1.Deadline": {"2026-05-01"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = readBody(t, postResp)

	var project models.Project
	if err := db.Preload("Milestones").Where("name = ?", "Beta Rollout").First(&project).Error; err != nil {
		t.Fatal(err)
	}
	if len(project.Milestones) != 2 {
		t.Fatalf("expected 2 milestones, got %d", len(project.Milestones))
	}

	showResp, err := client.Get(server.URL + "/admin/projects/" + formatUint(project.ID))
	if err != nil {
		t.Fatal(err)
	}
	showBody := readBody(t, showResp)
	if !strings.Contains(showBody, "Spec freeze") || !strings.Contains(showBody, "Internal beta") {
		t.Fatalf("expected project milestones in show page, body=%s", showBody)
	}
}

func TestAuditLogsReadOnlyModule(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(newGinHandler(app))
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	resp, err := client.Get(server.URL + "/admin/audits")
	if err != nil {
		t.Fatal(err)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "Audit Logs") || !strings.Contains(body, "Automated smoke validation completed successfully.") {
		t.Fatalf("expected seeded audit logs grid, body=%s", body)
	}
	if strings.Contains(body, ">Create<") || strings.Contains(body, ">Edit<") || strings.Contains(body, ">Delete<") {
		t.Fatalf("expected audit log module to be read-only, body=%s", body)
	}

	showResp, err := client.Get(server.URL + "/admin/audits/1")
	if err != nil {
		t.Fatal(err)
	}
	showBody := readBody(t, showResp)
	if !strings.Contains(showBody, "admin") || !strings.Contains(showBody, "article") {
		t.Fatalf("expected audit log show page, body=%s", showBody)
	}
}

func TestTicketsModuleRenderAndCreate(t *testing.T) {
	app, db := newTestFixture(t)
	server := httptest.NewServer(newGinHandler(app))
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	resp, err := client.Get(server.URL + "/admin/tickets")
	if err != nil {
		t.Fatal(err)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "Polish dashboard spacing") || !strings.Contains(body, "Framework Launch") {
		t.Fatalf("expected seeded tickets grid, body=%s", body)
	}

	formResp, err := client.Get(server.URL + "/admin/tickets/new")
	if err != nil {
		t.Fatal(err)
	}
	csrf := extractCSRF(t, readBody(t, formResp))
	postResp, err := client.PostForm(server.URL+"/admin/tickets", url.Values{
		"_csrf":       {csrf},
		"Title":       {"Close launch checklist"},
		"ProjectID":   {"1"},
		"Assignee":    {"Carol"},
		"Priority":    {"high"},
		"Status":      {"open"},
		"DueDate":     {"2026-04-22"},
		"Description": {"Finalize release checklist and rollout notes."},
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = readBody(t, postResp)

	var ticket models.Ticket
	if err := db.Preload("Project").Where("title = ?", "Close launch checklist").First(&ticket).Error; err != nil {
		t.Fatal(err)
	}
	if ticket.Project.Name != "Framework Launch" {
		t.Fatalf("expected belongs-to project preload, got %+v", ticket.Project)
	}

	showResp, err := client.Get(server.URL + "/admin/tickets/" + formatUint(ticket.ID))
	if err != nil {
		t.Fatal(err)
	}
	showBody := readBody(t, showResp)
	if !strings.Contains(showBody, "Carol") || !strings.Contains(showBody, "Framework Launch") {
		t.Fatalf("expected ticket detail page, body=%s", showBody)
	}
}

func TestReportsReadOnlyModule(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(newGinHandler(app))
	defer server.Close()

	client := newClient(t)
	login(t, client, server.URL)

	resp, err := client.Get(server.URL + "/admin/reports")
	if err != nil {
		t.Fatal(err)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "Weekly Signups") || !strings.Contains(body, "Support Backlog") {
		t.Fatalf("expected seeded reports grid, body=%s", body)
	}
	if strings.Contains(body, ">Create<") || strings.Contains(body, ">Edit<") || strings.Contains(body, ">Delete<") {
		t.Fatalf("expected reports module to be read-only, body=%s", body)
	}

	showResp, err := client.Get(server.URL + "/admin/reports/1")
	if err != nil {
		t.Fatal(err)
	}
	showBody := readBody(t, showResp)
	if !strings.Contains(showBody, "Weekly Signups") || !strings.Contains(showBody, "signups") {
		t.Fatalf("expected report detail page, body=%s", showBody)
	}
}

func newTestApp(t *testing.T) *goadmin.App {
	t.Helper()
	app, _ := newTestFixture(t)
	return app
}

func newTestFixture(t *testing.T) (*goadmin.App, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", url.QueryEscape(t.Name()))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	app, err := BuildWithDB(db)
	if err != nil {
		t.Fatal(err)
	}
	return app, db
}

func newClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar}
}

func newGinHandler(app *goadmin.App) http.Handler {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/admin")
	})
	router.Any("/admin", gin.WrapH(app.Handler()))
	router.Any("/admin/*any", gin.WrapH(app.Handler()))
	return router
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

func createResource(t *testing.T, client *http.Client, baseURL, formURL, submitURL string, values url.Values) {
	t.Helper()
	formResp, err := client.Get(baseURL + formURL)
	if err != nil {
		t.Fatal(err)
	}
	formBody := readBody(t, formResp)
	values.Set("_csrf", extractCSRF(t, formBody))
	resp, err := client.PostForm(baseURL+submitURL, values)
	if err != nil {
		t.Fatal(err)
	}
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected success for %s, got %d body=%s", submitURL, resp.StatusCode, body)
	}
}

type testUploadFile struct {
	Field   string
	Name    string
	Content []byte
}

func postMultipartForm(t *testing.T, client *http.Client, url string, values url.Values, files []testUploadFile) (*http.Response, error) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, list := range values {
		for _, value := range list {
			if err := writer.WriteField(key, value); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, upload := range files {
		part, err := writer.CreateFormFile(upload.Field, upload.Name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(upload.Content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return client.Do(req)
}

func extractUploadURL(t *testing.T, body string) string {
	t.Helper()
	re := regexp.MustCompile(`(/admin/uploads/[A-Za-z0-9/_\-.]+)`)
	matches := re.FindStringSubmatch(body)
	if len(matches) != 2 {
		t.Fatalf("upload url not found in body: %s", body)
	}
	return matches[1]
}

func assertStatusCode(t *testing.T, client *http.Client, url string, expected int) {
	t.Helper()
	resp, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d for %s, got %d body=%s", expected, url, resp.StatusCode, string(body))
	}
}

func formatUint(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

type stubRecord struct {
	ID   uint
	Name string
}

type stubRepository struct {
	items []any
}

func (s stubRepository) List(_ context.Context, query goadmin.ListQuery) (goadmin.ListResult, error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	perPage := query.PerPage
	if perPage < 1 {
		perPage = 10
	}
	return goadmin.ListResult{
		Items:   s.items,
		Total:   int64(len(s.items)),
		Page:    page,
		PerPage: perPage,
	}, nil
}

func (s stubRepository) Get(_ context.Context, id string) (any, error) {
	for _, item := range s.items {
		record := item.(stubRecord)
		if formatUint(record.ID) == id {
			return record, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (s stubRepository) Create(context.Context, goadmin.Values) error         { return nil }
func (s stubRepository) Update(context.Context, string, goadmin.Values) error { return nil }
func (s stubRepository) Delete(context.Context, string) error                 { return nil }
