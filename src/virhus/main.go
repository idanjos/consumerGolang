package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

//var pwd string = "/home/minty/Documents/TCIC/consumerGolang/"
var pwd = os.Args[1]
var vps = os.Args[2]

const (
	// TODO fill this in directly or through environment variable
	// Build a DSN e.g. postgres://username:password@url.com:5432/dbName
	DB_DSN = "postgres://postgres:admin@" + vps + ":5432/postgres?sslmode=disable"
)

var views = pwd + "/views/"
var assets = pwd + "/assets/"
var upgrader = websocket.Upgrader{}

var sensor = template.Must(template.ParseFiles(views + "sensorCharts/index.html"))
var tpl = template.Must(template.ParseFiles(views + "index.html"))
var test = template.Must(template.ParseFiles(views + "test.html"))
var homeTemplate = template.Must(template.ParseFiles(views + "echo.html"))
var chart = template.Must(template.ParseFiles(views + "dashboard/templates/wrapper.html"))
var dashboard = template.Must(template.ParseFiles(views + "dashboard/source/dashboard.html"))
var iot = template.Must(template.ParseFiles(views + "dashboard/source/dashboard.html"))

// var device = template.Must(template.ParseFiles(views + "dashboard/source/wrapper.html"))

//Main function
func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "3003"
	}

	db, err := sql.Open("postgres", DB_DSN)
	if err != nil {
		log.Fatal("Failed to open a DB connection: ", err)
	}
	defer db.Close()
	go rabbbitConsumer()
	conn, err := amqp.Dial("amqp://guest:guest@" + vps + ":5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	r := http.NewServeMux()

	r.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assets))))
	r.Handle("/iot/assets/", http.StripPrefix("/iot/assets/", http.FileServer(http.Dir(assets))))

	r.HandleFunc("/", route)

	r.HandleFunc("/iot", route)
	r.HandleFunc("/iot/", route)
	http.ListenAndServe(":"+port, r)
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

func iotPage(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", DB_DSN)
	if err != nil {
		log.Fatal("Failed to open a DB connection: ", err)
	}
	defer db.Close()
	stats := struct {
		TotalMsgs int
		TotalBand int
		NDevices  int
	}{
		TotalMsgs: 0,
		TotalBand: 0,
		NDevices:  0,
	}
	sql := "select count(data_id) as totalMsgs,sum(size) as totalBand, COUNT ( DISTINCT device_id ) as nDevices from data_register dr;"
	err = db.QueryRow(sql).Scan(&stats.TotalMsgs, &stats.TotalBand, &stats.NDevices)
	if err != nil {
		log.Fatal("Failed to execute query: ", err)
	}
	log.Println(stats)
	dat := template.Must(template.ParseFiles(views + "dashboard/app.html"))
	var tpl bytes.Buffer
	if err := dat.Execute(&tpl, stats); err != nil {
		log.Fatalln(err)
		return
	}

	data := struct {
		App template.HTML
	}{
		App: template.HTML(tpl.String()),
	}

	// iot.Execute(w, nil)
	chart.Execute(w, data)
}

func devicePage(w http.ResponseWriter, r *http.Request) {
	vars := strings.Split(r.URL.Path, "/")
	device := vars[2]

	db, err := sql.Open("postgres", DB_DSN)
	if err != nil {
		log.Fatal("Failed to open a DB connection: ", err)
	}
	defer db.Close()
	var macAddr string
	sql2 := "select macAddr from device_table dr where device_id = $1;"
	err = db.QueryRow(sql2, device).Scan(&macAddr)
	if err != nil {
		log.Fatal("Failed to execute query: ", err)
	}

	var options []string = []string{}
	rows, err := db.Query("select raw from data_register dr  where macAddr  = $1", macAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			log.Fatal(err)
		}
		options = append(options, raw)
	}
	stats := struct {
		TotalMsgs int
		TotalBand int
		MSGs      []string
	}{
		TotalMsgs: 0,
		TotalBand: 0,
		MSGs:      options,
	}
	sql := "select count(data_id) as totalMsgs,sum(size) as totalBand from data_register dr where device_id = $1;"
	err = db.QueryRow(sql, device).Scan(&stats.TotalMsgs, &stats.TotalBand)
	if err != nil {
		log.Fatal("Failed to execute query: ", err)
	}

	dat := template.Must(template.ParseFiles(views + "dashboard/app3.html"))
	var tpl bytes.Buffer
	if err := dat.Execute(&tpl, stats); err != nil {
		log.Fatalln(err)
		return
	}

	data := struct {
		App template.HTML
	}{
		App: template.HTML(tpl.String()),
	}

	// iot.Execute(w, nil)
	chart.Execute(w, data)
}

var rNum = regexp.MustCompile(`/iot/[0-9]+`) // Has digit(s)
var rAbc = regexp.MustCompile(`/iot`)        // Contains "abc"
func route(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	switch {
	case rNum.MatchString(r.URL.Path):
		devicePage(w, r)
	case rAbc.MatchString(r.URL.Path):
		iotPage(w, r)
	default:
		indexHandler(w, r)
	}
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func rabbbitConsumer() {
	conn, err := amqp.Dial("amqp://guest:guest@" + vps + ":5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	db, err := sql.Open("postgres", DB_DSN)
	if err != nil {
		log.Fatal("Failed to open a DB connection: ", err)
	}
	defer db.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		sqlStatement := `
insert into public.data_register (device_id,macAddr ,size ,raw )
VALUES ($1, $2, $3, $4)`

		for d := range msgs {
			data := strings.Split(string(d.Body), "mac ")
			if len(data) < 2 {
				continue
			}
			_, err = db.Exec(sqlStatement, 2, data[1], 50, d.Body)
			if err != nil {
				log.Println(err)
				panic(err)
			}

			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
