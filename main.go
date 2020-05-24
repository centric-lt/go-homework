package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"encoding/json"
	"net/http"
	"net/http/httptrace"
	"net/url"
)

type measurementResponse struct {
	host string
	protocol string
	results measurementResults
}

type measurementResults struct {
	measurements []string
	averageLatency string
}

func handleRequests() {
	http.HandleFunc("/measure", measureRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func measureRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	keys := r.URL.Query()
	host := keys["host"][0]
	protocol := keys["protocol"][0]
	samples, _:= strconv.Atoi(keys["samples"][0])

	url := url.URL{Host: host, Scheme: protocol}
	fmt.Fprintln(w, url.String())

	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, "Host: "+host)
	fmt.Fprintln(w, "protocol: "+protocol)
	fmt.Fprintln(w, "samples: "+ strconv.Itoa(samples))

	results := measurementResults{}
	for i := 0; i < samples; i++ {
		results.measurements.
	}
	measureTTFB(url)
}

func measureTTFB(url url.URL) time.Duration {
	var t0, t1 time.Time

	req, _ := http.NewRequest("GET", url.String(), nil)
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() { t1 = time.Now() },
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	t0 = time.Now()
	http.DefaultTransport.RoundTrip(req)

	fmt.Println("t0: " + t0.String())
	fmt.Println("t1: " + t1.String())
	return t1.Sub(t0)
}

func main() {
	//handleRequests()
	url, _ := url.Parse("https://reddit.com")
	fmt.Println("before")
	m := measureTTFB(*url)
	fmt.Println(m.Milliseconds())
	fmt.Println("after")
}
