package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/lcrownover/hpcidmtxn/internal/shared"
)

type User struct {
	Name string
	Uid  int
	Pirg string
}

func (u *User) IsPopulated() bool {
	if u.Name != "" && u.Uid != 0 && u.Pirg != "" {
		return true
	}
	return false
}

func GetUsersInPirg(pirgName string) []User {
	var userList []User
	cmd := exec.Command("getent", "group", pirgName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	trimmedOut := strings.TrimSpace(string(out))
	splitCmd := strings.Split(trimmedOut, ":")
	csus := splitCmd[len(splitCmd)-1]
	csul := strings.Split(csus, ",")
	for _, username := range csul {
		uid := shared.GetUIDFromUsername(username)
		u := User{Name: username, Pirg: pirgName, Uid: uid}
		if !u.IsPopulated() {
			log.Fatalln("Unable to get populate user: ", u)
		}
		userList = append(userList, u)
	}
	return userList
}

func FindAndChown(u User) {
	// findPath := fmt.Sprintf("/gpfs/projects/%s", u.Pirg)
	// cmd := exec.Command("find", findPath, "-user", u.Name, "-exec", "chown", string(u.Uid), "{}", "\\;")
	fmt.Printf("finding user '%s', pirg: '%s', uid: '%d'\n", u.Name, u.Pirg, u.Uid)
}

func main() {
	pirgName := flag.String("p", "", "pirg name")
	flag.Parse()

	if *pirgName == "" {
		fmt.Println("pirg is required")
		flag.Usage()
		os.Exit(2)
	}

	userList := GetUsersInPirg(*pirgName)

	var wg sync.WaitGroup

	for _, user := range userList {
		wg.Add(1)
		go func(u User) {
			defer wg.Done()
			FindAndChown(u)
		}(user)
	}

	wg.Wait()
}
