package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type RequestPayload struct {
	Id    string
	Token string
	User  string
	Msg   string
}

type VirtualHuman struct {
	Id   string
	User string
	Name string
	Path string
}

//Records of virtual humans
const Records string = pwd + "src/records/"

var ID int64 = 0

var VirtualHumans map[string][]VirtualHuman = map[string][]VirtualHuman{}

func load() {
	files, err := ioutil.ReadDir(Records)
	if err != nil {
		log.Fatal("Failed to Open Records", err)
		return
	}

	for _, f := range files {
		if _, ok := VirtualHumans["root"]; ok {
			x := VirtualHuman{
				Id:   strconv.FormatInt(atomic.AddInt64(&ID, 1), 10),
				User: "root",
				Path: Records + f.Name(),
				Name: f.Name(),
			}
			VirtualHumans["root"] = append(VirtualHumans["root"], x)

		} else {
			VirtualHumans["root"] = []VirtualHuman{VirtualHuman{
				Id:   strconv.FormatInt(atomic.AddInt64(&ID, 1), 10),
				User: "root",
				Path: Records + f.Name(),
				Name: f.Name(),
			}}
		}
	}
	log.Println(VirtualHumans)
}

func createVH(w http.ResponseWriter, r *http.Request) {
	//Verify auth
	defer timer(time.Now())
	if r.Method == "GET" {
		notFound(w)
		return
	}
	//Check if exist in database
	var s RequestPayload
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		internalError(w)
		return
	}
	id := strconv.FormatInt(atomic.AddInt64(&ID, 1), 10)
	json.Unmarshal(body, &s)
	//filter point
	path := Records + s.Msg
	if _, ok := VirtualHumans[s.User]; !ok {
		VirtualHumans[s.User] = []VirtualHuman{VirtualHuman{
			Id:   id,
			Path: path,
			User: s.User,
			Name: s.Msg,
		}}
	} else {
		VirtualHumans[s.User] = append(VirtualHumans[s.User], VirtualHuman{
			Id:   id,
			Path: path,
			User: s.User,
			Name: s.Msg,
		})
		if err != nil {
			internalError(w)
			log.Fatal(err)
			return
		}
	}
	//files, err := ioutil.ReadDir("./src/records/"+s.Id)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0700)
		log.Println("Virtual Human " + id + " created!")
	}

	//log.Println(runtime.NumGoroutine())
	//time.Sleep(15 * time.Second)

}

func downloadRecord(w http.ResponseWriter, r *http.Request) {
	//Verify auth
	//get from Database
	defer timer(time.Now())
	var s RequestPayload
	body, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &s)
	if err != nil {
		internalError(w)
		return
	}

	http.ServeFile(w, r, Records+s.Msg)

}

func getVirtualHumans(w http.ResponseWriter, r *http.Request) {
	//Verify auth
	//get from Database
	defer timer(time.Now())
	var s RequestPayload
	body, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &s)
	if err != nil {
		internalError(w)
		return
	}

	output, _ := json.Marshal(VirtualHumans)
	w.Write([]byte(output))
}

func getVirtualHuman(w http.ResponseWriter, r *http.Request) {

	//Verify auth
	//get from Database
	defer timer(time.Now())
	var s RequestPayload
	body, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &s)
	if err != nil {
		internalError(w)
		return
	}

	if _, ok := VirtualHumans[s.Msg]; !ok {
		notFound(w)
		return
	}
	output, _ := json.Marshal(VirtualHumans[s.Msg])
	w.Write([]byte(output))

}

func getRecordings(w http.ResponseWriter, r *http.Request) {
	//Verify auth
	//get from Database
	defer timer(time.Now())
	var s RequestPayload
	body, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &s)
	if err != nil {
		internalError(w)
		return
	}

	var arr []string = []string{}
	log.Println(s)
	files, err := ioutil.ReadDir(Records + s.Msg)
	for _, file := range files {
		arr = append(arr, file.Name())
	}
	output, _ := json.Marshal(arr)
	w.Write([]byte(output))
}
