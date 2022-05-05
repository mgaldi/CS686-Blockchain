package main

import (
	"bytes"
	"crypto/dsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
)

type Validator struct {
	ProcessNr   int
	StakeAmount int64
	Signature   Signature
	Public      *dsa.PublicKey
}
type Signature struct {
	R *big.Int
	S *big.Int
}
type Book struct {
	Title  string
	Author string
	Pages  int
	Year   int
}

const PORTS_GAP = 4
const RANGE_API_START = 46000

var (
	validators []Validator
	vCheck     map[int]bool
)

func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		for _, v := range validators {
			go http.Post(fmt.Sprintf("http://localhost:%d/block/new", (v.ProcessNr*PORTS_GAP)+RANGE_API_START), "application/json", bytes.NewBuffer(body))
		}
	}
	http.Redirect(w, r, "/", http.StatusFound)

}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/history" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	resp, err := http.Get("http://localhost:8880/list")

	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	fmt.Fprintln(w, bodyString)

}
func validatorsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var received Validator
		err := json.NewDecoder(r.Body).Decode(&received)
		if err != nil {
			fmt.Println(err)
		}
		if _, ok := vCheck[received.ProcessNr]; !ok {
			validators = append(validators, received)
			vCheck[received.ProcessNr] = true
		}
	}
	if r.Method == "GET" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(&validators)
	}
}

func main() {
	fileServer := http.FileServer(http.Dir("./blockchainweb/build"))
	http.Handle("/", fileServer)
	http.HandleFunc("/new", formHandler)
	http.HandleFunc("/history", historyHandler)
	http.HandleFunc("/validators", validatorsHandler)
	vCheck = make(map[int]bool)
	fmt.Printf("Starting server\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
