package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
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
	w.Header().Set("Content-Type", "application/json")

	//Get query parameters
	keys := r.URL.Query()

	m := measurementResponse{}

	//Check if host was specified
	if keys["host"] != nil {
		m.Host = keys["host"][0]
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "host not specified")
		return
	}

	//If protocol not specified, use default "http"
	if keys["protocol"] != nil {
		m.Protocol = keys["protocol"][0]
	} else {
		m.Protocol = "http"
	}

	var samples int
	// If samples not specified, use default 1
	if keys["samples"] != nil {
		samples, _ = strconv.Atoi(keys["samples"][0])
		if samples < 1 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "given samples value not allowed")
			return
		}
	} else {
		samples = 1
	}

	m.Results.Measurements = make([]string, samples)

	// Create url object
	u := url.URL{Host: m.Host, Scheme: m.Protocol}

	// Create channel for goroutine
	c := make(chan time.Duration, samples)
	for i := 0; i < samples; i++ {
		go measureTTFB(u, c)
	}

	// Get measurements
	var sum int64 = 0
	for i := 0; i < samples; i++ {
		t := <-c
		sum += t.Milliseconds()
		m.Results.Measurements[i] = getScaleTime(t)
	}

	// Convert measurements to string
	strSum, _ := time.ParseDuration(fmt.Sprintf("%dms", sum/int64(len(m.Results.Measurements))))
	m.Results.AverageLatency = getScaleTime(strSum)

	j, _ := json.MarshalIndent(&m, "", "    ")
	fmt.Fprintln(w, string(j))
}

func measureTTFB(url url.URL, c chan time.Duration) {
	var t0, t1 time.Time

	req, _ := http.NewRequest("GET", url.String(), nil)
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() { t1 = time.Now() }, // When the first response arrives, get the time
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	t0 = time.Now()
	http.DefaultTransport.RoundTrip(req)

	c <- t1.Sub(t0)
}

func getScaleTime(d time.Duration) string {
	if d.Seconds() < 5 {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else {
		return fmt.Sprintf("%fs", d.Seconds())
	}
}

func main() {
	lambda.Start(handleRequests)
	//handleRequests()
}
