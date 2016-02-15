package main

import (
	_ "crypto/sha512"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/gorilla/mux"
	"net/http"
	"log"
	"os/exec"
	"os"
	"strings"
	"encoding/json"
	"fmt"
	"errors"
)

const (
	PORT = ":8888"
	TRAVIS_API_PREFIX = "http://api.travis-ci.org/repos/"
	TRAVIS_API_POSTFIX = "/builds"
)

var (
	ORIGINAL_WORKING_DIR string
	monitors []*Monitor
)

type Monitor struct {
	Travis        string `json:"travis" yaml:"travis"`
	Dir           string `json:"dir" yaml:"dir"`
	Actions       []string `json:"action" yaml:"action"`
	CurrentCommit string
}

func (m *Monitor) SetCommit(commit string) {
	m.CurrentCommit = commit
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}
type Routes []Route

type Build struct {
	ID          int
	Repo_ID     int
	Number      string
	State       string
	Result      int
	Started_at  string
	Finished_at string
	Duration    int
	Commit      string
	Branch      string
	Message     string
	Event_type  string
}

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

	println("receiving....")

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

func checkAPIForSuccess(m *Monitor) (Build, error) {

	apiPath := TRAVIS_API_PREFIX + m.Travis + TRAVIS_API_POSTFIX
	response, err := http.Get(apiPath)

	checkSoft(err)

	decoder := json.NewDecoder(response.Body)
	var builds []Build
	err = decoder.Decode(&builds)
	checkSoft(err)

	if (len(builds) < 1) {
		println("no builds for", m.Travis, "found, you may wish to check the spelling")
		return Build{}, errors.New("no builds")
	} else {
		latestBuild := builds[0]
		if (latestBuild.Result == 0) {
			return latestBuild, nil;

		} else {
			return Build{}, errors.New("non 0 result for build")
		}
	}
}

func loadConfig() {
	monitors = []*Monitor{}
	dat, err := ioutil.ReadFile(".malinois.yml")
	checkHard(err)
	err = yaml.Unmarshal(dat, &monitors)
	checkHard(err)

	for _, m := range monitors {

		apiPath := TRAVIS_API_PREFIX + m.Travis + TRAVIS_API_POSTFIX

		response, err := http.Get(apiPath)

		checkHard(err)

		if (response.StatusCode == 200) {

			decoder := json.NewDecoder(response.Body)
			var builds []Build
			err := decoder.Decode(&builds)
			checkHard(err)

			if (len(builds) < 1) {
				println("no builds for", m.Travis, "found, you may wish to check the spelling")
			}

		} else {
			log.Fatal(m.Travis, " returned a non 200 status code")
		}

	}

	println("loaded config OK")
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

func runMonitorAction(m *Monitor) {

	latestBuild, err := checkAPIForSuccess(m)

	if (err != nil) {
		println("the latest build failed, not running actions for", m.Travis);
	} else {
		//TODO check if we have run actions for this commit
		if (latestBuild.Commit != m.CurrentCommit) {
			m.SetCommit(latestBuild.Commit)

			println("commit", m.CurrentCommit)

			println("running actions for", m.Travis)
			os.Chdir(m.Dir)
			for _, action := range m.Actions {
				runAction(action)
			}
			os.Chdir(ORIGINAL_WORKING_DIR)
		} else {
			println("the actions for this commit have already been run")
		}
	}

}

func runAction(action string) (output []byte, err error) {
	var sSlice = strings.Split(action, " ")
	var command = sSlice[0]
	var args = sSlice[1:len(sSlice)]
	out, err := exec.Command(command, args...).Output()
	if (err != nil) {
		log.Println("error!",err)
	}
	return out, err
}

func checkHard(e error) {
	//causes exit
	if e != nil {
		log.Fatal(e)
	}
}
func checkSoft(e error) {
	//does not cause exit
	if (e != nil) {
		println("ERROR!", e)
	}
}