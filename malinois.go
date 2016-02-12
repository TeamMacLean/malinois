package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
	"log"
	"os/exec"
	"os"
	"strings"
)

const (
	PORT = ":8888"
)

var (
	ORIGINAL_WORKING_DIR string
	monitors []Monitor
)

type Monitor struct {
	Github  string `json:"github" yaml:"github"`
	Travis  string `json:"travis" yaml:"travis"`
	Dir     string `json:"dir" yaml:"dir"`
	Actions []string `json:"action" yaml:"action"`
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}
type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},
	Route{
		"FF",
		"POST",
		"/",
		NAMEME,
	},
	//Route{
	//	"TodoShow",
	//	"GET",
	//	"/todos/{todoId}",
	//	TodoShow,
	//},
}

func main() {

	ORIGINAL_WORKING_DIR, _ = os.Getwd()

	_, err := checkForGit()
	if (err != nil) {
		log.Fatal("Could not find git on path")
	}

	loadConfig()
	startServer()
}

func NAMEME(w http.ResponseWriter, r *http.Request) {

	println(r.URL.Query())

	thisMonitor := monitors[0]
	go runMonitorAction(thisMonitor)

	fmt.Fprintln(w, "Welcome!")
}

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
		Methods(route.Method).
		Path(route.Pattern).
		Name(route.Name).
		Handler(route.HandlerFunc)
	}

	return router
}

func loadConfig() {
	monitors = []Monitor{}
	dat, err := ioutil.ReadFile(".malinois.yml")
	check(err)
	err = yaml.Unmarshal(dat, &monitors)
	check(err)
}

func checkForGit() (path string, err error) {
	return exec.LookPath("git")
}

func startServer() {

	router := NewRouter()

	println("starting server on port", PORT)
	log.Fatal(http.ListenAndServe(PORT, router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, monitors)
}

func runMonitorAction(m Monitor) {
	os.Chdir(m.Dir)
	for _, action := range m.Actions {
		runAction(action)
	}
	os.Chdir(ORIGINAL_WORKING_DIR)
}

func runAction(action string) (output []byte, err error) {

	println("running action");

	var sSlice = strings.Split(action, " ")

	var command = sSlice[0]

	var args = sSlice[1:len(sSlice)]

	out, err := exec.Command(command, args...).Output()

	if (err != nil) {
		fmt.Printf("error %s\n", err)
	}
	fmt.Printf("output %s\n", out)

	return out, err
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}