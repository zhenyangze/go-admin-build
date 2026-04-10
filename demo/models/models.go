package models

import "time"

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
	ID         uint   `gorm:"primaryKey"`
	Title      string `gorm:"size:160"`
	Summary    string `gorm:"size:500"`
	Status     string `gorm:"size:40"`
	CategoryID uint
	Category   Category
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
