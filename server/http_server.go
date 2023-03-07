package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func rootHandler(l *Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l.log("Got Request / at" + fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)))
		http.ServeFile(w, r, "./static/input.html")
	}
}

func apiInHandler(l *Logger, c chan InputData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l.log("got Request /api at " + fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond)))
		data := <-c
		out := processData(&data, l)
		io.WriteString(w, data.toString()+strconv.Itoa(out.mode))
	}
}

func imageHandler(l *Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		fileBytes, _ := ioutil.ReadFile("static/images/" + name + ".png")
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(fileBytes)
	}
}

func host(l *Logger, c chan InputData) {
	http.HandleFunc("/", rootHandler(l))
	http.HandleFunc("/api/in", apiInHandler(l, c))
	http.HandleFunc("/image", imageHandler(l))

	err := http.ListenAndServe("localhost:3333", nil)

	if errors.Is(err, http.ErrServerClosed) {
		l.log("server closed\n")
	} else if err != nil {
		l.log(fmt.Sprintf("error starting server: %s\n", err))
	}
}
