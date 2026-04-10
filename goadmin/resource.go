package goadmin

import (
	"github.com/zhenyangze/go-admin-build/goadmin/form"
	"github.com/zhenyangze/go-admin-build/goadmin/grid"
	"github.com/zhenyangze/go-admin-build/goadmin/show"
	"github.com/zhenyangze/go-admin-build/goadmin/tree"
)

// Resource describes one admin module with Dcat-like page builders.
type Resource struct {
	Name        string
	Path        string
	Title       string
	Description string
	Icon        string
	Permission  string
	Repository  Repository
	BuildGrid   func(*grid.Builder)
	BuildForm   func(*form.Builder)
	BuildShow   func(*show.Builder)
	BuildTree   func(*tree.Builder)
}

// HasTree reports whether the resource exposes a tree page.
func (r Resource) HasTree() bool {
	return r.BuildTree != nil
}
