package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/zhenyangze/go-admin-build/examples/demo"
)

func main() {
	dbPath := os.Getenv("DEMO_DB_PATH")
	if dbPath == "" {
		dbPath = "tmp/demo/admin.db"
	}
	if os.Getenv("DEMO_RESET") == "1" {
		_ = os.Remove(dbPath)
		_ = os.RemoveAll(filepath.Join(filepath.Dir(dbPath), "uploads"))
	}
	app, _, err := demo.Build(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/admin")
	})
	router.Any("/admin", gin.WrapH(app.Handler()))
	router.Any("/admin/*any", gin.WrapH(app.Handler()))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("demo ready at http://127.0.0.1:%s/admin (admin/admin)", port)
	log.Fatal(router.Run(":" + port))
}
