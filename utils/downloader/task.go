package downloader

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type TaskStatus int

const (
	Pending TaskStatus = iota //等待下载
	Running
	Paused
	Failed
	Finished
)

type OnFinishedFunc func(err error)
type Callback func() error

//包裹字节slice使其可被当作io.Writer使用
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

//存放任务的信息
type Task struct {
	Id     string     //任务的id,需独特
	Status TaskStatus //任务的状态,参见TaskStatus

	req      http.Request
	onFinish OnFinishedFunc
	before   func() (bool, error)
	after    Callback
	holder   io.WriteCloser
}

//创建一个将内容写入到文件的任务,如果文件不存在,则创建文件
//id:任务的id,req:url请求,path:保存路径,finish:任务完成后的回调函数
func NewTaskSaveToDisk(id string, req http.Request, path string, fileName string, finish OnFinishedFunc) *Task {
	r := &Task{
		Id:       id,
		req:      req,
		Status:   Pending,
		onFinish: finish,
	}
	var f *os.File
	fullPath := filepath.Join(path, fileName)
	r.before = func() (bool, error) {
		dwlpath := fullPath
		var err error
		if _, err = os.Stat(path); os.IsNotExist(err) {
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return false, err
			}
		}
		if _, err = os.Stat(filepath.Join(path, fileName)); !os.IsNotExist(err) {
			return true, nil
		}

		f, err = os.OpenFile(dwlpath+".part", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			defer f.Close()
		}
		w := io.WriteCloser(f)
		r.holder = w
		return false, err
	}

	r.after = func() error {
		dwlpath := fullPath
		err := os.Rename(dwlpath+".part", dwlpath)
		if err != nil {
			r.failOnError(err)
		}
		return r.holder.Close()
	}
	return r
}

//创建一个将内容写入到变量的任务
//id:任务的id,req:url请求,target:待写入变量,finish:任务完成后的回调函数
func NewTaskToBytes(id string, req http.Request, target *[]byte, finish OnFinishedFunc) *Task {
	r := &Task{
		Id:       id,
		req:      req,
		Status:   Pending,
		onFinish: finish,
	}
	r.holder = &byteWriter{b: target}
	r.before = func() (bool, error) { return false, nil }
	r.after = func() error { return nil }

	return r
}
func (t *Task) failOnError(err error) {
	t.Status = Failed
	t.onFinish(err)

}
