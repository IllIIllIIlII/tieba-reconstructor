package thread

import (
	"fmt"
	"sync"
	config "tieba-reconstructor/config"
	vector "tieba-reconstructor/utils/containers"
)

const (
	threadUrlTemplate  string = "https://tieba.baidu.com/p/%s?pn=%d"
	commentUrlTemplate string = "https://tieba.baidu.com/p/totalComment?tid=%s&pn=%d&see_lz=%d" // thread id, page, see lz

)

type image struct {
	threadId string
	postId   string
	url      string
}
type ThreadTask struct {
	ForumName     string
	Id            string // thread id
	SeeLz         bool   // see lz
	FloorLimit    uint
	EnableComment bool // if enabled, comment will be scraped as well.
	DownloadPath  string
	Status        config.Status

	currentPage  uint
	currentFloor uint

	lock *sync.Mutex

	FailedPage        *vector.Vector    // pages failed to scrape are stored here.
	FailedPageComment *vector.Vector    // the same applies to comments based on page.
	FailedImage       map[string]string // 下载地址：本地名称
}

func New(id string, flimit uint, seelz bool, comment bool) *ThreadTask {
	return &ThreadTask{
		FailedPage:        vector.New(),
		FailedPageComment: vector.New(),
		FailedImage:       map[string]string{},
		Id:                id,

		FloorLimit:    flimit,
		SeeLz:         seelz,
		EnableComment: comment,

		currentPage:  0,
		currentFloor: 0,

		lock: &sync.Mutex{},
	}
}

func (t *ThreadTask) hasNext() bool {
	return t.currentFloor != t.FloorLimit
}

func (t *ThreadTask) nextPageUrl() string {
	if !t.hasNext() {
		return ""
	} else {
		t.currentPage += 1
		//		t.currentFloor += 15
		return fmt.Sprintf(threadUrlTemplate, t.Id, t.currentPage)
	}
}
