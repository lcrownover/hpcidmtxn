package shared

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func GetUIDFromUsername(username string) int {
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
