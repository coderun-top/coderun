package queue

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type Nums struct {
	BuildCount int `json:"build_num"`
	AgentCount int `json:"agent_num"`
	ConCount   int `json:"concurrency_num"`
}

type entry struct {
	item     *Task
	done     chan bool
	retry    int
	error    error
	deadline time.Time
}

type worker struct {
	filter   Filter
	channel  chan *Task
	agent_id string
}

type fifo struct {
	sync.Mutex

	workers   map[*worker]struct{}
	running   map[string]*entry
	pending   *list.List
	extension time.Duration
}

// New returns a new fifo queue.
func New() Queue {
	return &fifo{
		workers:   map[*worker]struct{}{},
		running:   map[string]*entry{},
		pending:   list.New(),
		extension: time.Minute * 10,
	}
}

// Push pushes an item to the tail of this queue.
func (q *fifo) Push(c context.Context, task *Task) error {
	q.Lock()
	q.pending.PushBack(task)
	q.Unlock()
	go q.process()
	return nil
}

// Poll retrieves and removes the head of this queue.
func (q *fifo) Poll(c context.Context, f Filter, agent_id string) (*Task, error) {
	q.Lock()
	w := &worker{
		channel:  make(chan *Task, 1),
		filter:   f,
		agent_id: agent_id,
	}
	q.workers[w] = struct{}{}
	q.Unlock()
	go q.process()

	for {
		select {
		case <-c.Done():
			q.Lock()
			delete(q.workers, w)
			q.Unlock()
			return nil, nil
		case t := <-w.channel:
			return t, nil
		}
	}
}

// Done signals that the item is done executing.
func (q *fifo) Done(c context.Context, id string) error {
	return q.Error(c, id, nil)
}

// Error signals that the item is done executing with error.
func (q *fifo) Error(c context.Context, id string, err error) error {
	q.Lock()
	state, ok := q.running[id]
	if ok {
		state.error = err
		close(state.done)
		delete(q.running, id)
	}
	q.Unlock()
	return nil
}

// Evict removes a pending task from the queue.
func (q *fifo) Evict(c context.Context, id string) error {
	q.Lock()
	defer q.Unlock()

	var next *list.Element
	for e := q.pending.Front(); e != nil; e = next {
		next = e.Next()
		task, ok := e.Value.(*Task)
		if ok && task.ID == id {
			q.pending.Remove(e)
			return nil
		}
	}
	return ErrNotFound
}

// Wait waits until the item is done executing.
func (q *fifo) Wait(c context.Context, id string) error {
	q.Lock()
	state := q.running[id]
	q.Unlock()
	if state != nil {
		select {
		case <-c.Done():
		case <-state.done:
			return state.error
		}
	}
	return nil
}

// Extend extends the task execution deadline.
func (q *fifo) Extend(c context.Context, id string) error {
	q.Lock()
	defer q.Unlock()

	state, ok := q.running[id]
	if ok {
		state.deadline = time.Now().Add(q.extension)
		return nil
	}
	return ErrNotFound
}

// Info returns internal queue information.
func (q *fifo) GetWorkers(c context.Context) []*Task {
	q.Lock()
	data := []*Task{}

	for w := range q.workers {
		t := <-w.channel
		data = append(data, t)
	}

	q.Unlock()
	return data
}

// Info returns internal queue information.
func (q *fifo) Info(c context.Context) InfoT {
	q.Lock()
	stats := InfoT{}
	var agents []string
	stats.Stats.Workers = len(q.workers)
	stats.Stats.Pending = q.pending.Len()
	stats.Stats.Running = len(q.running)

	for e := q.pending.Front(); e != nil; e = e.Next() {
		stats.Pending = append(stats.Pending, e.Value.(*Task))
	}
	for _, entry := range q.running {
		stats.Running = append(stats.Running, entry.item)
	}

	for w := range q.workers {
		agents = append(agents, w.agent_id)
	}
	stats.Agents = agents

	q.Unlock()
	return stats
}

func (q *fifo) GetRuning(c context.Context, project string) int {
	q.Lock()
	running_count := 0
	for _, entry := range q.running {
		if entry.item.Labels["project"] == project && entry.item.Labels["public"] == "true" {
			running_count = running_count + 1
		}
	}

	q.Unlock()
	return running_count
}

// helper function that loops through the queue and attempts to
// match the item to a single subscriber.
func (q *fifo) process() {
	defer func() {
		// the risk of panic is low. This code can probably be removed
		// once the code has been used in real world installs without issue.
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("queue: unexpected panic: %v\n%s", err, buf)
		}
	}()

	q.Lock()
	defer q.Unlock()

	// TODO(bradrydzewski) move this to a helper function
	// push items to the front of the queue if the item expires.
	for id, state := range q.running {
		if time.Now().After(state.deadline) {
			q.pending.PushFront(state.item)
			delete(q.running, id)
			close(state.done)
		}
	}

	var next *list.Element
loop:
	for e := q.pending.Front(); e != nil; e = next {
		running_count := 0
		next = e.Next()
		item := e.Value.(*Task)
		// log.Printf("project: %s", item.Labels["project"])

        // 这里进行并发数控制
		if item.Labels["public"] == "true" {
			for _, entry := range q.running {
				if item.Labels["project"] == entry.item.Labels["project"] && entry.item.Labels["public"] == "true" {
					running_count = running_count + 1
				}
			}
			log.Printf("count: %s", running_count)

			numURL := os.Getenv("DOUSER_NUM_URL")
			nums, err := GetNum(numURL + "/" + item.Labels["project"])
			if err != nil {
				log.Printf("get num error")
			}
			if nums.ConCount == 0 {
				break loop
			}
			if running_count > nums.ConCount-1 {
				break loop
			}
		}
		//if item.Labels["agent"] != "" {
		//	if item.Labels["agent"] != item.Labels["label"] {
		//			break loop
		//	}
		//}
		for w := range q.workers {
			if w.filter(item) {
				delete(q.workers, w)
				q.pending.Remove(e)

				q.running[item.ID] = &entry{
					item:     item,
					done:     make(chan bool),
					deadline: time.Now().Add(q.extension),
				}

				w.channel <- item
				break loop
			}
		}
	}
}

func GetNum(url string) (*Nums, error) {
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", url, nil)

	if err != nil {
		panic(err)
	}
	response, _ := client.Do(reqest)
	if response != nil {
		defer response.Body.Close()
	}
	// 检查状态码
	status := response.StatusCode
	if status != 200 {
		return nil, errors.New("http请求失败")
	}

	bodyByte, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}
	data := new(Nums)
	err = json.Unmarshal(bodyByte, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
