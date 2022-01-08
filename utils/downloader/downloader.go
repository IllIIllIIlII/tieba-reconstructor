package downloader

import (
	"log"
	"net/http"
	vector "tieba-reconstructor/utils/containers"
	network "tieba-reconstructor/utils/net"

	"github.com/pkg/errors"
)

type DownloaderStatus int

const (
	Idle DownloaderStatus = iota
	Started
	Gone
)

type Err404 struct{}

func (e Err404) Error() string {
	return "404"
}

type Downloader struct {
	tasks   *vector.Vector
	in      chan *Task
	out     chan *Task
	free    chan interface{}
	pause   chan interface{}
	resume  chan interface{}
	stop    chan interface{}
	workers []*taskWorker
	DownloaderStatus
}

type taskWorker struct {
	task   *Task
	client http.Client
}

func (w *taskWorker) download(t *Task) {
	w.task = t
	req := t.req
	resp, err := network.Retry(w.client, &req, 3)
	if err != nil {
		t.failOnError(err)
	}

	// 200-300: save result on demand
	// 404: return 404 error
	// else: return error
	code := resp.StatusCode
	if code >= 200 && code < 300 {
		body := resp.Body
		defer body.Close()
		defer t.Holder.Close()
		total := resp.ContentLength
		count := 0
		for int64(count) < total {
			buffer := []byte{}
			_, err := body.Read(buffer)
			if err != nil {
				t.failOnError(err)
				return
			}
			n, err := t.Holder.Write(buffer)
			if err != nil {
				t.failOnError(err)
				return
			}
			count += n
		}
		t.onFinish(nil)
		return

		// body, err := ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	t.Status = Failed
		// 	t.onFinish(err)
		// 	return
		// }
		// resp.Body.Close()
		// t.store(&body)
		// t.onFinish(nil)
		// return
	} else if code == 404 {
		t.failOnError(Err404{})
		return

	} else {
		t.failOnError(errors.Errorf("not a 2xx code"))
		return
	}

}
func newWorker() *taskWorker {
	r := &taskWorker{
		client: http.Client{
			Transport: http.DefaultTransport,
		},
	}

	return r
}

func New() *Downloader {
	r := &Downloader{
		in:    make(chan *Task, 1),
		out:   make(chan *Task, 1),
		tasks: vector.New(),

		DownloaderStatus: Idle,
	}
	//r.workers = make([]*taskWorker, 4)
	return r
}
func (d *Downloader) AddTask(t *Task) {
	d.in <- t
}
func (d *Downloader) SetWorkers(num int) {
	d.free = make(chan interface{}, num)
	d.pause = make(chan interface{}, num)
	d.resume = make(chan interface{}, num)
	d.stop = make(chan interface{}, num)
}
func (d *Downloader) Start() {
	go func() { // downloader receiving thread
		queue := make([]*Task, 0)
		for {
			select {
			case t := <-d.in: // when a task is added, queue or assign to worker
				d.tasks.Append(t)
				select {
				case <-d.free:
					d.out <- t
				default:
					queue = append(queue, t)
				}

			case <-d.free: // when a worker is free, take a task from the queue or wait for task to assign
				n := len(queue)
				var t *Task
				if n > 0 {
					t = queue[0]
					queue = queue[1:]

				} else {
					t = <-d.in
				}
				d.out <- t
			}

		}
	}()
	workernum := 4
	d.SetWorkers(workernum)
	for i := 0; i < workernum; i++ {
		w := newWorker()
		d.workers = append(d.workers, w)
		go func() { // worker receiving thread
			var x interface{}
			d.free <- x
			for {
				select {
				case t := <-d.out:
					err := t.before()
					if err == nil {
						w.download(t)
						t.after()
					} else {
						t.Status = Failed
						log.Println("Task id: ", t.Id, " failed because ", err.Error())
					}

				case <-d.pause:
					<-d.resume
				case <-d.stop:
					return
				}
			}
		}()
	}
}
