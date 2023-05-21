package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// migrationdata:
// {'pirg': 'hpcrcf', 't1gid': '50202', 't2gid': '50202', 'users': [{'username': 'dmajchrz', 't1uid': '2939', 't2uid': '261151'}, {'username': 'lrc', 't1uid': '2780', 't2uid': '79413'}]}
type UserData struct {
	Username string `json:"username"`
	T1UID    string `json:"t1uid"`
	T2UID    string `json:"t2uid"`
}
type MigrationData struct {
	Pirg  string     `json:"pirg"`
	T1GID string     `json:"t1gid"`
	T2GID string     `json:"t2gid"`
	Users []UserData `json:"users"`
}

func GetADUIDFromUsername(username string) (int, error) {
	cmd := exec.Command("getent", "passwd", username)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}
	trimmedOut := strings.TrimSpace(string(out))
	uid, err := strconv.Atoi(strings.Split(trimmedOut, ":")[2])
	if err != nil {
		log.Fatal(err)
	}
	return uid, nil
}

func GetADGIDFromGroupname(groupname string) (int, error) {
	cmd := exec.Command("getent", "group", groupname)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}
	trimmedOut := strings.TrimSpace(string(out))
	gid, err := strconv.Atoi(strings.Split(trimmedOut, ":")[2])
	if err != nil {
		log.Fatal(err)
	}
	return gid, nil
}

func GetADUsernameFromUID(uid int) (string, error) {
	cmd := exec.Command("id", "-nu", fmt.Sprint(uid))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	username := strings.TrimSpace(string(out))
	return username, nil
}

func GetGroupMemberships() (string, error) {
	data, err := os.ReadFile("/etc/hpcidmtxn/t1group-memberships.txt")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func loadT1IdMap(path string) (map[string]int, error) {
	t1Usermap := make(map[string]int)
	body, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading file '%s': %v", path, err)
	}
	for _, line := range strings.Split(string(body), "\n") {
		if len(line) == 0 {
			continue
		}
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

	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	router.GET("/t1/user/:name", func(c *gin.Context) {
		name := c.Param("name")

		usermap, err := loadT1IdMap("/etc/hpcidmtxn/t1users.csv")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "failed to load usermap",
			})
		}

		uid, ok := usermap[name]
		if !ok {
			c.String(http.StatusNotFound, "%s", "not found")
		}

		c.String(http.StatusOK, "%s", fmt.Sprintf("%d", uid))
	})

	router.GET("/t1/group/:name", func(c *gin.Context) {
		name := c.Param("name")

		groupmap, err := loadT1IdMap("/etc/hpcidmtxn/t1groups.csv")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "failed to load groupmap",
			})
		}

		gid, ok := groupmap[name]
		if !ok {
			c.String(http.StatusNotFound, "%s", "not found")
		}

		c.String(http.StatusOK, "%s", fmt.Sprintf("%d", gid))
	})

	router.GET("/t2/user/:name", func(c *gin.Context) {
		name := c.Param("name")

		uid, err := GetADUIDFromUsername(name)
		if uid == 0 || err != nil {
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

	router.GET("/t2/groups", func(c *gin.Context) {
		gm, err := GetGroupMemberships()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "failed to load group memberships",
			})
		}

		c.String(http.StatusOK, "%s", fmt.Sprintf("%s", gm))
	})

	router.GET("/t2/group/:name", func(c *gin.Context) {
		name := c.Param("name")

		gid, err := GetADGIDFromGroupname(name)
		if gid == 0 || err != nil {
			c.String(http.StatusNotFound, "%s", "not found")
		}

		c.String(http.StatusOK, "%s", fmt.Sprintf("%d", gid))
	})

	router.POST("/t1/groupmemberships", func(c *gin.Context) {
		// upload file
		file, err := c.FormFile("file")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "no file uploaded",
			})
		}

		if err := c.SaveUploadedFile(file, "/etc/hpcidmtxn/t1group-memberships.txt"); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("'%s' failed to upload", file.Filename),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("'%s' uploaded!", file.Filename),
		})
	})

	router.POST("/t1/groups", func(c *gin.Context) {
		// upload file
		file, err := c.FormFile("file")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "no file uploaded",
			})
		}

		if err := c.SaveUploadedFile(file, "/etc/hpcidmtxn/t1groups.csv"); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("'%s' failed to upload", file.Filename),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("'%s' uploaded!", file.Filename),
		})
	})

	router.POST("/t1/users", func(c *gin.Context) {
		// upload file
		file, err := c.FormFile("file")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "no file uploaded",
			})
		}

		if err := c.SaveUploadedFile(file, "/etc/hpcidmtxn/t1users.csv"); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("'%s' failed to upload", file.Filename),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("'%s' uploaded!", file.Filename),
		})
	})

	router.GET("/migrationdata/:pirg", func(c *gin.Context) {
		pirgName := c.Param("pirg")
        filePath := fmt.Sprintf("/etc/hpcidmtxn/data/%s.json", pirgName)
        if _, err := os.Stat(filePath); err != nil {
            c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
                "message": fmt.Sprintf("pirg '%s' not found in migration data", pirgName),
            })
        }
		var outputData MigrationData
		jsonFile, err := os.Open(fmt.Sprintf("/etc/hpcidmtxn/data/%s.json", pirgName))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("failed to open file"),
			})
		}
		defer jsonFile.Close()
		jsonParser := json.NewDecoder(jsonFile)
		if err = jsonParser.Decode(&outputData); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("failed to read file"),
			})
		}
		c.JSON(http.StatusOK, outputData)
	})

	router.POST("/migrationdata", func(c *gin.Context) {
		var input MigrationData
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		pirgName := input.Pirg
		file, err := json.MarshalIndent(input, "", " ")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("failed to marshal json"),
			})
		}
		err = ioutil.WriteFile(fmt.Sprintf("/etc/hpcidmtxn/data/%s.json", pirgName), file, 0644)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("failed to write file"),
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "data written",
		})
	})

	router.Run(":8080")
}
