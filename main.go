package main

func main() {

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
