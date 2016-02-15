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
//TRAVIS_API_PREFIX = "api.travis-ci.org/repos/"
//TRAVIS_API_POSTFIX = "/builds"
)

var (
	ORIGINAL_WORKING_DIR string
	monitors []Monitor
)

type Monitor struct {
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
		"Index Show",
		"GET",
		"/",
		Index,
	},
	Route{
		"Index Post",
		"POST",
		"/",
		PostUpdate,
	},
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

func PostUpdate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	for _, v := range r.Form {
		vJoined := strings.Join(v, "")
		println("received request for", vJoined)
		for _, mv := range monitors {
			if (strings.ToLower(vJoined) == strings.ToLower(mv.Travis)) {
				go runMonitorAction(mv)
			}
		}
	}
	fmt.Fprintln(w, "done")
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

	//TOCO check via API!!!

	println("running actions for", m.Travis)
	os.Chdir(m.Dir)
	for _, action := range m.Actions {
		runAction(action)
	}
	os.Chdir(ORIGINAL_WORKING_DIR)
}

func runAction(action string) (output []byte, err error) {
	var sSlice = strings.Split(action, " ")
	var command = sSlice[0]
	var args = sSlice[1:len(sSlice)]
	out, err := exec.Command(command, args...).Output()
	if (err != nil) {
		fmt.Printf("error %s\n", err)
	}
	return out, err
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}