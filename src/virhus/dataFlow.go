package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const sourceData string = pwd + "src/sourceData/"
const recordData string = pwd + "src/records/"

var neutralFiles []string = []string{}
var happyFiles []string = []string{}
var fearFiles []string = []string{}

type SendData struct {
	Timestamp string
	Data      []string
}

type ReqData struct {
	Id        string
	Timestamp string
	Emotion   string
	Size      int
	Offset    int
}

func writeData(arr []string, fileName string, id string) {

	var record *os.File

	record, err := os.OpenFile(recordData+id+"/"+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0700)

	defer record.Close()
	if err != nil {
		log.Fatal("Error on opeing file", err)
	}
	//log.Println(arr)
	for _, line := range arr {
		//fmt.Println(line)
		if _, err := record.WriteString(line + "\n"); err != nil {
			log.Fatalln(err)
		}
	}

}

func GetData(r ReqData) SendData {

	var output SendData = SendData{
		Timestamp: strconv.FormatInt(time.Now().Unix(), 10),
		Data:      []string{},
	}
	if r.Timestamp != "0" {
		output.Timestamp = r.Timestamp
	}

	var fileHandle *os.File
	var err error

	switch r.Emotion {
	case "neutral":
		//x := rand.Int31n(int32(len(neutralFiles)))
		fileHandle, err = os.Open(sourceData + r.Emotion + "/" + neutralFiles[0])
	case "fear":
		//x := rand.Int31n(int32(len(fearFiles)))
		fileHandle, err = os.Open(sourceData + r.Emotion + "/" + fearFiles[0])
	case "happy":
		//x := rand.Int31n(int32(len(happyFiles)))
		fileHandle, err = os.Open(sourceData + r.Emotion + "/" + happyFiles[0])
	}

	if err != nil {
		log.Fatalln(err)
	}

	//fileHandle, _ := os.Open(sourceData + r.Emotion)
	log.Println(r)
	defer fileHandle.Close()

	fileScanner := bufio.NewScanner(fileHandle)

	for fileScanner.Scan() {

		for i := 0; i < r.Size; i++ {
			fileScanner.Scan()
			output.Data = append(output.Data, fileScanner.Text())

		}
		log.Println(r, len(output.Data))
		fileHandle.Close()
		writeData(output.Data, r.Id+"_"+output.Timestamp+".dat", r.Id)
		return output

	}

	return output

}

func loadData() {
	files, err := ioutil.ReadDir(sourceData + "neutral")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		neutralFiles = append(neutralFiles, file.Name())
	}

	files, err = ioutil.ReadDir(sourceData + "fear")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fearFiles = append(fearFiles, file.Name())
	}

	files, err = ioutil.ReadDir(sourceData + "happy")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		happyFiles = append(happyFiles, file.Name())
	}
	fmt.Println(neutralFiles, happyFiles, fearFiles)
}

func getData(w http.ResponseWriter, r *http.Request) {
	//Verify auth
	//get from Database
	defer timer(time.Now())
	var s ReqData
	body, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &s)
	if err != nil {
		log.Fatalln(err)
		internalError(w)
		return
	}

	output, _ := json.Marshal(GetData(s))

	w.Write([]byte(output))
}
