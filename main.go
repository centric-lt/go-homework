package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"time"
)

type measurementResponse struct {
	Host     string             `json:"host"`
	Protocol string             `json:"protocol"`
	Results  measurementResults `json:"results"`
}

type measurementResults struct {
	Measurements   []string `json:"measurements"`
	AverageLatency string   `json:"averageLatency"`
}

func handleRequests() {
	http.HandleFunc("/measure", measureRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func measureRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/j")

	keys := r.URL.Query()
	m := measurementResponse{}

	m.Host = keys["host"][0]
	m.Protocol = keys["protocol"][0]

	samples, _ := strconv.Atoi(keys["samples"][0])
	m.Results.Measurements = make([]string, samples)

	u := url.URL{Host: m.Host, Scheme: m.Protocol}

	c := make(chan time.Duration, samples)
	for i := 0; i < samples; i++ {
		measureTTFB(u, c)
	}

	var sum int64 = 0
	for i := 0; i < samples; i++ {
		t := <-c
		sum += t.Milliseconds()
		m.Results.Measurements[i] = getScaleTime(t)
	}

	strSum, _ := time.ParseDuration(fmt.Sprintf("%dms", sum/int64(len(m.Results.Measurements))))
	m.Results.AverageLatency = getScaleTime(strSum)

	j, _ := json.MarshalIndent(&m, "", "    ")
	fmt.Fprintln(w, string(j))
}

func measureTTFB(url url.URL, c chan time.Duration) {
	var t0, t1 time.Time

	req, _ := http.NewRequest("GET", url.String(), nil)
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() { t1 = time.Now() },
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	t0 = time.Now()
	http.DefaultTransport.RoundTrip(req)

	c <- t1.Sub(t0)
	//fmt.Println("t0: " + t0.String())
	//fmt.Println("t1: " + t1.String())
}

func getScaleTime(d time.Duration) string {
	if d.Seconds() < 5 {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else {
		return fmt.Sprintf("%fs", d.Seconds())
	}
}

func main() {
	handleRequests()
}
