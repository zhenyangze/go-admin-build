package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/zhenyangze/go-admin-build/demo"
)

func main() {
	app, _, err := demo.Build("tmp/demo/admin.db")
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
