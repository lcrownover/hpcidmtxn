package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func GetRemoteUIDFromUsername(username string, serverName string) int {
	url := fmt.Sprintf("http://%s/user/%s", serverName, username)
	res, err := http.Get(url)
	if err != nil {
		log.Fatal("error making request to: %s", url)
	}
	uid, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal("error reading response body")
	}
	uidString := strings.TrimSpace(string(uid))
	uidInt, err := strconv.Atoi(uidString)
	if err != nil {
		log.Fatal("error converting uid string to int")
	}
	return uidInt
}

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

func GetUsersInPirg(pirgName string, serverName string) []User {
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
		uid := shared.GetUIDFromUsername(username, serverName)
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
	serverName := flag.String("s", "", "server name for hpcidmtxn_server")
	pirgName := flag.String("p", "", "pirg name")
	flag.Parse()

	if *serverName == "" {
		fmt.Println("server name is required")
		flag.Usage()
		os.Exit(2)
	}
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
