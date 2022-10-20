package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lcrownover/hpcidmtxn/internal/shared"
)

func main() {
	router := gin.Default()

	router.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")

		uid := shared.GetUIDFromUsername(name)

		c.String(http.StatusOK, "%s", uid)
	})

	router.Run(":8080")
}
