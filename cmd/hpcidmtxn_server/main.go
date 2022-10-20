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

func GetLocalUIDFromUsername(username string) (int, error) {
	cmd := exec.Command("id", "-u", username)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}
	trimmedOut := strings.TrimSpace(string(out))
	uid, err := strconv.Atoi(trimmedOut)
	if err != nil {
		log.Fatal(err)
	}
	return uid, nil
}

func main() {
	router := gin.Default()

	router.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")

		uid, err := GetLocalUIDFromUsername(name)
		if err != nil {
			c.String(http.StatusNotFound, "%s", "not found")
		}

		c.String(http.StatusOK, "%s", fmt.Sprintf("%d", uid))
	})

	router.Run(":8080")
}
