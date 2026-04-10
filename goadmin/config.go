package goadmin

import "github.com/zhenyangze/go-admin-build/goadmin/theme"

// Config defines the reusable admin app settings.
type Config struct {
	AppName       string
	Title         string
	Prefix        string
	SessionSecret string
	SessionCookie string
	Theme         theme.Theme
}

// WithDefaults normalizes configuration values.
func (c Config) WithDefaults() Config {
	if c.AppName == "" {
		c.AppName = "Go Admin"
	}
	if c.Title == "" {
		c.Title = c.AppName
	}
	if c.Prefix == "" {
		c.Prefix = "/admin"
	}
	if c.SessionCookie == "" {
		c.SessionCookie = "goadmin_session"
	}
	if c.Theme.Accent == "" {
		c.Theme = theme.Default()
	}
	return c
}
