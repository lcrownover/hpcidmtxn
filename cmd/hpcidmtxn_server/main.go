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

func GetLocalUsernameFromUID(uid int) (string, error) {
	cmd := exec.Command("id", "-nu", fmt.Sprint(uid))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	username := strings.TrimSpace(string(out))
	return username, nil
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

	router.GET("/uid/:uid", func(c *gin.Context) {
		param_uid := c.Param("uid")
		uid, err := strconv.Atoi(param_uid)
		if err != nil {
			c.String(http.StatusBadRequest, "invalid integer: %s", fmt.Sprintf("%d", uid))
		}

		username, err := GetLocalUsernameFromUID(uid)
		if err != nil {
			c.String(http.StatusNotFound, "not found")
		}

		c.String(http.StatusOK, "%s", username)
	})

	router.Run(":8080")
}
