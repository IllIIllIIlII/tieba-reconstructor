package downloader

import (
	"io"
	"net/http"
	"os"
)

type TaskStatus int

const (
	Pending TaskStatus = iota
	Running
	Paused
	Failed
	Finished
)

type OnFinishedFunc func(err error)
type Callback func() error

type byteWriter struct {
	b *[]byte
}

func (b *byteWriter) Write(p []byte) (int, error) {
	*b.b = append(*b.b, p...)
	return len(p), nil
}
func (b *byteWriter) Close() error {
	b.b = nil
	return nil
}

type Task struct {
	Id       string
	req      http.Request
	Status   TaskStatus
	store    func(b *[]byte) error
	onFinish OnFinishedFunc
	before   Callback
	after    Callback
	Holder   io.WriteCloser
}

func NewTaskSaveToDisk(id string, req http.Request, path string, finish OnFinishedFunc) *Task {
	r := &Task{
		Id:       id,
		req:      req,
		Status:   Pending,
		onFinish: finish,
	}
	var f *os.File
	r.before = func() error {
		var err error
		f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		w := io.WriteCloser(f)
		r.Holder = w
		return err
	}

	// r.store = func(b *[]byte) error {
	// 	_, err := r.Holder.Write(*b)

	// 	return err
	// }
	r.after = func() error {
		return r.Holder.Close()
	}
	return r
}

func NewTaskToBytes(id string, req http.Request, target *[]byte, finish OnFinishedFunc) *Task {
	r := &Task{
		Id:       id,
		req:      req,
		Status:   Pending,
		onFinish: finish,
	}
	r.Holder = &byteWriter{b: target}
	r.before = func() error { return nil }
	r.after = func() error { return nil }

	return r
}
func (t *Task) failOnError(err error) {
	t.Status = Failed
	t.onFinish(err)

}
