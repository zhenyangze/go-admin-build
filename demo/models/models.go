package models

import (
	"time"

	"gorm.io/gorm"
)

// Category exercises the tree module.
type Category struct {
	ID          uint `gorm:"primaryKey"`
	ParentID    uint
	Name        string `gorm:"size:120"`
	Description string `gorm:"size:255"`
	Sort        int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Article exercises grid, form, and show with a belongs-to relation.
type Article struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:160"`
	Summary     string `gorm:"size:500"`
	Status      string `gorm:"size:40"`
	Featured    bool
	Tags        string `gorm:"size:255"`
	Image       string `gorm:"size:255"`
	Gallery     string `gorm:"size:1000"`
	PublishedAt time.Time
	VisibleFrom time.Time
	VisibleTo   time.Time
	CategoryID  uint
	Category    Category
	FAQs        []ArticleFAQ
	Links       []ArticleLink
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ArticleFAQ is a relation-backed nested child model for article FAQs.
type ArticleFAQ struct {
	ID        uint   `gorm:"primaryKey"`
	ArticleID uint   `gorm:"index"`
	Question  string `gorm:"size:255"`
	Answer    string `gorm:"size:1000"`
	Sort      int
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ArticleLink is a second relation-backed nested child model for article links/resources.
type ArticleLink struct {
	ID        uint   `gorm:"primaryKey"`
	ArticleID uint   `gorm:"index"`
	Label     string `gorm:"size:255"`
	URL       string `gorm:"size:500"`
	Sort      int
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Project is a second business module used to prove the framework supports more than content pages.
type Project struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:160"`
	Owner       string `gorm:"size:120"`
	Status      string `gorm:"size:40"`
	Budget      int
	StartsAt    time.Time
	EndsAt      time.Time
	Description string `gorm:"size:1000"`
	Milestones  []ProjectMilestone
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProjectMilestone is a relation-backed nested child model for project checkpoints.
type ProjectMilestone struct {
	ID        uint   `gorm:"primaryKey"`
	ProjectID uint   `gorm:"index"`
	Title     string `gorm:"size:255"`
	Deadline  time.Time
	Sort      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AuditLog is a read-only operational module used in the demo.
type AuditLog struct {
	ID         uint   `gorm:"primaryKey"`
	Actor      string `gorm:"size:120"`
	Action     string `gorm:"size:120"`
	Resource   string `gorm:"size:120"`
	ResourceID string `gorm:"size:120"`
	Level      string `gorm:"size:40"`
	IP         string `gorm:"size:64"`
	Detail     string `gorm:"size:1000"`
	CreatedAt  time.Time
}

// Ticket is a generic CRUD business module that mainly exercises the reusable gormstore repository.
type Ticket struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:160"`
	ProjectID   uint
	Project     Project
	Assignee    string `gorm:"size:120"`
	Priority    string `gorm:"size:40"`
	Status      string `gorm:"size:40"`
	Description string `gorm:"size:1000"`
	DueDate     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ReportSnapshot is a read-only reporting module used to demonstrate summary/statistics pages.
type ReportSnapshot struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:160"`
	Metric      string `gorm:"size:120"`
	Dimension   string `gorm:"size:120"`
	Value       int
	Trend       string `gorm:"size:40"`
	SnapshotAt  time.Time
	Description string `gorm:"size:500"`
	CreatedAt   time.Time
}
