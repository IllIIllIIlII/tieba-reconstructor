package test

import (
	"sync"
	"testing"
	"tieba-reconstructor/utils/downloader"
	network "tieba-reconstructor/utils/net"
)

func TestDownloader(T *testing.T) {
	d := downloader.New()
	var b *[]byte = &[]byte{}
	req, _ := network.GetRequestFromUrl("https://tieba.baidu.com/p/1766018024")
	wg := &sync.WaitGroup{}
	task := downloader.NewTaskToBytes("", *req, b, func(err error) {

		if err != nil {
			T.Fatal(err)
		} else {
			T.Log(*b)
		}
		defer wg.Done()
	})
	wg.Add(1)
	d.Start()
	d.AddTask(task)
	wg.Wait()
}
