package main

import (
	"fmt"
	"net/http"
)

type userHandler struct {
	queue *map[string]Deque
}

func (h *userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	queueName := r.URL.Path
	fmt.Println("start: ", method, queueName, r.URL.Query(), h.queue, (*h.queue)[queueName])

	if method == "PUT" {
		message := r.URL.Query().Get("v")
		if len(message) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		deque, exists := (*h.queue)[queueName]
		if !exists {
			deque = Deque{}
		}
		deque.PushFront(message)
		(*h.queue)[queueName] = deque
	}

	if method == "GET" {
		timeout := r.URL.Query().Get("timeout")
		// TODO: timeouts while GET request
		deque, _ := (*h.queue)[queueName]
		result, exist := deque.PopBack()
		(*h.queue)[queueName] = deque

		if exist {
			_, err := w.Write([]byte(result))
			if err != nil {
				return
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}

	fmt.Println("end: ", method, queueName, r.URL.Query(), h.queue, (*h.queue)[queueName])
}

func main() {
	queue := make(map[string]Deque)
	mux := http.NewServeMux()
	mux.Handle("/", &userHandler{&queue})
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return
	}
}

// Deque however better to use https://github.com/gammazero/deque for O(N) while insert
type Deque struct {
	queue []string
}

func (d *Deque) PushFront(message string) {
	d.queue = append([]string{message}, d.queue...)
}

func (d *Deque) PopBack() (string, bool) {
	if len(d.queue) == 0 {
		return "", false
	}
	lastIndex := len(d.queue) - 1
	lastElement := d.queue[lastIndex]
	d.queue = d.queue[:lastIndex]
	return lastElement, true
}
