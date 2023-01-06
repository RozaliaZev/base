package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
)

const addrProxy string = "localhost:9000"

var(
	counter int = 0
	firstInstanceHost string = "http://localhost:3001"
	secondInstanceHost string = "http://localhost:8081"
)


func main() {
   http.HandleFunc("/", handleProxy)
    log.Fatalln(http.ListenAndServe(addrProxy,nil))
}


func handleProxy(w http.ResponseWriter, r *http.Request) {
    
    textBytes, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Fatalln(err)
    }
    text := string(textBytes)
    log.Println(r.Method + text)

    if counter == 0 {
        client := http.Client{}
        req, err  := http.NewRequest(r.Method, firstInstanceHost + r.URL.EscapedPath(), bytes.NewBuffer(textBytes))
        if err != nil {
            log.Fatalln(err)
        }
        _, err = client.Do(req)
	    if err != nil {
		    log.Fatalln(err)
	    }
        counter++
        return
    }

    if counter != 0 {
        client := http.Client{}
        req, err  := http.NewRequest(r.Method, secondInstanceHost + r.URL.EscapedPath(), bytes.NewBuffer(textBytes))
        if err != nil {
            log.Fatalln(err)
        }
        _, err = client.Do(req)
	    if err != nil {
		    log.Fatalln(err)
	    }
        counter--
        return
    }
}
