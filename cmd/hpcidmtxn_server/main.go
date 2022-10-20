package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetLocalUIDFromUsername(username string) int {
	cmd := exec.Command("id", "-u", username)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	trimmedOut := strings.TrimSpace(string(out))
	uid, err := strconv.Atoi(trimmedOut)
	if err != nil {
		log.Fatal(err)
	}
	return uid
}

func main() {
	router := gin.Default()

	router.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")

		uid := GetLocalUIDFromUsername(name)

		c.String(http.StatusOK, "%s", fmt.Sprintf("%d", uid))
	})

	router.Run(":8080")
}
