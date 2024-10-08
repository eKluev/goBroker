package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type userHandler struct {
	queue    *map[string]Deque
	mutex    sync.Mutex
	waitList map[string][]chan struct{}
}

func (h *userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	queueName := r.URL.Path

	h.mutex.Lock()
	deque, exists := (*h.queue)[queueName]
	if !exists {
		deque = Deque{}
		(*h.queue)[queueName] = deque
	}
	h.mutex.Unlock()

	if method == "PUT" {
		message := r.URL.Query().Get("v")
		if len(message) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.mutex.Lock()
		deque.PushFront(message)
		(*h.queue)[queueName] = deque

		if waiters, ok := h.waitList[queueName]; ok && len(waiters) > 0 {
			firstWaiter := waiters[0]
			h.waitList[queueName] = waiters[1:]
			close(firstWaiter)
		}
		h.mutex.Unlock()
	}

	if method == "GET" {
		timeoutStr := r.URL.Query().Get("timeout")
		timeout := 0
		if len(timeoutStr) > 0 {
			t, err := strconv.Atoi(timeoutStr)
			if err == nil {
				timeout = t
			}
		}

		h.mutex.Lock()
		result, exist := deque.PopBack()
		(*h.queue)[queueName] = deque
		h.mutex.Unlock()

		if exist {
			_, err := w.Write([]byte(result))
			if err != nil {
				return
			}
		} else if timeout > 0 {
			ch := make(chan struct{})
			h.mutex.Lock()
			h.waitList[queueName] = append(h.waitList[queueName], ch)
			h.mutex.Unlock()

			select {
			case <-ch:
				h.mutex.Lock()
				d, _ := (*h.queue)[queueName]
				result, exist = d.PopBack()
				(*h.queue)[queueName] = d
				h.mutex.Unlock()

				if exist {
					_, _ = w.Write([]byte(result))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			case <-time.After(time.Duration(timeout) * time.Second):
				h.mutex.Lock()
				waiters := h.waitList[queueName]
				for i := 0; i < len(waiters); i++ {
					if waiters[i] == ch {
						h.waitList[queueName] = append(waiters[:i], waiters[i+1:]...)
						break
					}
				}
				h.mutex.Unlock()

				w.WriteHeader(http.StatusNotFound)
			}

		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func main() {
	port := flag.Int("port", 8080, "Port for HTTP server")
	flag.Parse()

	queue := make(map[string]Deque)
	waitList := make(map[string][]chan struct{})

	mux := http.NewServeMux()
	mux.Handle("/", &userHandler{queue: &queue, waitList: waitList})
	address := fmt.Sprintf(":%d", *port)
	err := http.ListenAndServe(address, mux)
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
