package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

const pwd string = "/home/minty/Documents/Projects/PI_2.0/"

var views = pwd + "src/virhus/views/"
var assets = pwd + "src/virhus/assets/"
var upgrader = websocket.Upgrader{}

var sensor = template.Must(template.ParseFiles(views + "sensorCharts/index.html"))
var tpl = template.Must(template.ParseFiles(views + "index.html"))
var test = template.Must(template.ParseFiles(views + "test.html"))
var homeTemplate = template.Must(template.ParseFiles(views + "echo.html"))
var chart = template.Must(template.ParseFiles(views + "dashboard/templates/wrapper.html"))
var dashboard = template.Must(template.ParseFiles(views + "dashboard/source/dashboard.html"))

//Main function
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3003"
	}
	load()
	loadData()
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir(assets))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/home", home)
	mux.HandleFunc("/echo", echo)
	mux.HandleFunc("/test", testing)
	mux.HandleFunc("/secret", secret)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/logout", logout)
	mux.HandleFunc("/app", app)
	mux.HandleFunc("/overview", overview)

	mux.HandleFunc("/getData", getData)
	mux.HandleFunc("/getVHs", getVirtualHumans)
	mux.HandleFunc("/getVH", getVirtualHuman)
	mux.HandleFunc("/createVH", createVH)
	mux.HandleFunc("/getRecordings", getRecordings)
	mux.HandleFunc("/downloadRecord", downloadRecord)
	http.ListenAndServe(":"+port, mux)
}

// Websocket handler/server
func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

//Route Response, endpoint responses
func indexHandler(w http.ResponseWriter, r *http.Request) {
	defer timer(time.Now())
	tpl.Execute(w, nil)
}

func secret(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Print secret message
	fmt.Fprintln(w, "The cake is a lie!")
}

func login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Authentication goes here
	// ...

	// Set user as authenticated
	session.Values["authenticated"] = true
	session.Save(r, w)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Save(r, w)
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func testing(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
		Items []string
	}{
		Title: "My page",
		Items: []string{},
	}
	test.Execute(w, data)
}

func app(w http.ResponseWriter, r *http.Request) {
	//dat, err := ioutil.ReadFile(views + "dashboard/app2.html")
	//check(err)
	//get username
	user := "root"
	var options []string = []string{}
	for _, val := range VirtualHumans[user] {
		options = append(options, val.Name)
	}
	vhs := struct {
		VHs []string
	}{
		VHs: options,
	}
	dat := template.Must(template.ParseFiles(views + "dashboard/app2.html"))
	var tpl bytes.Buffer
	if err := dat.Execute(&tpl, vhs); err != nil {
		log.Fatalln(err)
		return
	}
	data := struct {
		App template.HTML
	}{
		App: template.HTML(tpl.String()),
	}

	//fmt.Println(data)
	chart.Execute(w, data)
	//chart.Execute(w, nil)
}

func overview(w http.ResponseWriter, r *http.Request) {
	dashboard.Execute(w, nil)
}

//Additional functions
func timer(start time.Time) {
	log.Println("Execution request: ", time.Since(start))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func notFound(w http.ResponseWriter) {
	http.Error(w, "404 not found", http.StatusNotFound)
}

func internalError(w http.ResponseWriter) {
	http.Error(w, "INTERNAL ERROR", http.StatusInternalServerError)
}

func ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{1})
}
