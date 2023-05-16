package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetADUIDFromUsername(username string) (*int, error) {
	cmd := exec.Command("id", "-u", username)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	trimmedOut := strings.TrimSpace(string(out))
	uid, err := strconv.Atoi(trimmedOut)
	if err != nil {
		log.Fatal(err)
	}
	return &uid, nil
}

func GetADGIDFromGroupname(groupname string) (*int, error) {
	cmd := exec.Command("getent", "group", groupname)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	trimmedOut := strings.TrimSpace(string(out))
	gid, err := strconv.Atoi(strings.Split(trimmedOut, ":")[2])
	if err != nil {
		log.Fatal(err)
	}
	return &gid, nil
}

func GetADUsernameFromUID(uid int) (*string, error) {
	cmd := exec.Command("id", "-nu", fmt.Sprint(uid))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(string(out))
	return &username, nil
}

func loadT1IdMap(path string) (map[string]int, error) {
	t1Usermap := make(map[string]int)
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(body), "\n") {
		splitLine := strings.Split(line, ",")
		id, err := strconv.Atoi(splitLine[1])
		if err != nil {
			log.Fatalf("Error parsing line '%s': %v", line, err)
		}
		t1Usermap[splitLine[0]] = id
	}
	return t1Usermap, nil
}

func main() {
	usermap, err := loadT1IdMap("/etc/hpcidmtxn/t1users.csv")
	if err != nil {
		log.Fatal(err)
	}
	groupmap, err := loadT1IdMap("/etc/hpcidmtxn/t1groups.csv")
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	router.GET("/t1/user/:name", func(c *gin.Context) {
		name := c.Param("name")

		uid, ok := usermap[name]
		if !ok {
			c.String(http.StatusNotFound, "%s", "not found")
		}

		c.String(http.StatusOK, "%s", fmt.Sprintf("%d", uid))
	})

	router.GET("/t1/group/:name", func(c *gin.Context) {
		name := c.Param("name")

		gid, ok := groupmap[name]
		if !ok {
			c.String(http.StatusNotFound, "%s", "not found")
		}

		c.String(http.StatusOK, "%s", fmt.Sprintf("%d", gid))
	})

	router.GET("/t2/user/:name", func(c *gin.Context) {
		name := c.Param("name")

		uid, err := GetADUIDFromUsername(name)
		if err != nil {
			c.String(http.StatusNotFound, "%s", "not found")
		}

		c.String(http.StatusOK, "%s", fmt.Sprintf("%d", uid))
	})

	router.GET("/t2/uid/:uid", func(c *gin.Context) {
		param_uid := c.Param("uid")
		uid, err := strconv.Atoi(param_uid)
		if err != nil {
			c.String(http.StatusBadRequest, "invalid integer: %s", fmt.Sprintf("%d", uid))
		}

		username, err := GetADUsernameFromUID(uid)
		if err != nil {
			c.String(http.StatusNotFound, "not found")
		}

		c.String(http.StatusOK, "%s", username)
	})

	router.GET("/t2/group/:name", func(c *gin.Context) {
		name := c.Param("name")

		uid, err := GetADGIDFromGroupname(name)
		if err != nil {
			c.String(http.StatusNotFound, "%s", "not found")
		}

		c.String(http.StatusOK, "%s", fmt.Sprintf("%d", uid))
	})

	router.Run(":8080")
}
