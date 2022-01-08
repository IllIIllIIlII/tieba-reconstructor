package thread

import (
	"fmt"
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
	Id            string // thread id
	SeeLz         bool   // see lz
	FloorLimit    uint
	EnableComment bool // if enabled, comment will be scraped as well.
	DownloadPath  string
	Status        config.Status

	currentPage  uint
	currentFloor uint

	failedPageChan    chan uint
	failedIamgeChan   chan image
	done              chan interface{}
	FailedPage        *vector.Vector     // pages failed to scrape are stored here.
	FailedPageComment *vector.Vector     // the same applies to comments based on page.
	FailedImage       *map[uint][]string // images failed to download from the original page
}

func New(id string, flimit uint, seelz bool, comment bool) *ThreadTask {
	return &ThreadTask{
		FailedPage:        vector.New(),
		FailedPageComment: vector.New(),
		FailedImage:       &map[uint][]string{},
		Id:                id,

		FloorLimit:    flimit,
		SeeLz:         seelz,
		EnableComment: comment,
		done:          make(chan interface{}),
		currentPage:   0,
		currentFloor:  0,
	}
}

func (t *ThreadTask) hasNext() bool {
	return true
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
