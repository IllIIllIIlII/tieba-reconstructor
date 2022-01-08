package thread

import (
	"log"
	"os"
	"tieba-reconstructor/config"
	"tieba-reconstructor/models"
	vector "tieba-reconstructor/utils/containers"
	"tieba-reconstructor/utils/downloader"
	network "tieba-reconstructor/utils/net"
)

const (
	localImgUrlTemplate string = "./images/%s_%d.jpg"
	imgSrcAttr          string = "src"
	hrefAttr            string = "href"
	datafield           string = "data-field"
	datapid             string = "data-pid"
)

type pageInfo struct {
	content  []byte // raw content
	threadId string
	pageNum  uint
}

// The controller that controls the thread downloading task
type ThreadController struct {
	threads          vector.Vector
	downloader       downloader.Downloader
	phtmls           chan pageInfo
	imgurls          chan string
	downloadRootPath string
	logger           *log.Logger
}

// return a controller
func NewThreadController() *ThreadController {
	root, _ := os.Getwd()
	return &ThreadController{
		threads:          *vector.New(),
		downloader:       *downloader.New(),
		phtmls:           make(chan pageInfo, 4),
		downloadRootPath: root,
		logger:           log.Default(),
	}
}

// starts the controller
func (c *ThreadController) Start() {
	c.downloader.Start()
}

// add a thread task to the controller but not start it
func (c *ThreadController) AddTask(t *ThreadTask) {
	c.threads.Append(t)
}

/*
	Start a given thread task based on its index in the vector
	i: the index of the thread task
	Pages of the thread is scraped in the order of their page number. It starts from the first page.
	The content of each post is stored under HMTL nodes with class="d_post_content j_d_post_content ".
	Contents of each post will be extracted, and every image will be downloaded
	while their src link inside the content will be replaced with the local path of the image.
	This is how the downloaded content will be handled under "export to static html" mode:
		The local path is the root of the download path a given thread,
		where each reconstructed page in the thread appears.
		Images will be stored in the images folder under the root of the download path.
		Failed pages and images will be stored in "failed_pages.json" and "failed_images.json"
		under the root of the download path.
	This is for "saving to the server" mode if MongoDB is installed and this mode is chosen:
	TODO: implement this
*/
func (c *ThreadController) StartTask(i int) {
	t := c.threads.GetValue(i).(*ThreadTask)
	t.Status = config.Running
	var thread models.TiebaThread
	var floors []models.Floor
	thread.Id = t.Id
	thread.FloorNumber = t.FloorLimit
	thread.Floors = floors

	//添加贴子下载任务

	var url string
	url = t.nextPageUrl()
	if url != "" {
		pagenum := t.currentPage
		c.logger.Println("Start downloading thread ", thread.Id, ", page ", pagenum)
		req, _ := network.GetRequestFromUrl(url)
		b := &[]byte{}

		//添加贴子页面i下载任务
		tsk := downloader.NewTaskToBytes(t.Id, *req, b, func(err error) {
			if err != nil {
				// 404则无视，否则加入失败列表
				_, ok := err.(downloader.Err404)
				if !ok {

				}
			}
		})
		c.downloader.AddTask(tsk)
	}

	// for ; t.HasNext(); url = t.NextPageUrl() {
	// 	pagenum := t.currentPage
	// 	req, _ := network.GetRequestFromUrl(url)
	// 	b := &[]byte{}
	// 	tsk := downloader.NewTaskToBytes(t.Id, *req, b, func(err error) {
	// 		threadid := t.Id
	// 		if err != nil {
	// 			c.phtmls <- pageInfo{
	// 				threadId: threadid,
	// 				content:  *b,
	// 				pageNum:  pagenum,
	// 			}
	// 		} else {
	// 			if err.Error() == "Not a 2xx code" {
	// 				//TODO: unretrievable
	// 			} else {
	// 				t.failedPageChan <- pagenum
	// 			}
	// 		}
	// 	})
	// 	c.downloader.AddTask(tsk)
	// }
	// go func() {
	// 	regex := regexp.MustCompile("post_content_[0-9]+")
	// 	id := t.Id
	// 	for {
	// 		select {
	// 		case n := <-t.failedPageChan:
	// 			t.FailedPage.Append(n)
	// 		case p := <-c.phtmls:
	// 			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(p.content))
	// 			if err != nil {
	// 				t.FailedPage.Append(p.pageNum)
	// 			} else {
	// 				doc.Find(".l_post l_post_bright j_l_post clearfix  ").Each(func(i int, h *goquery.Selection) {
	// 					s := h.Find(".d_post_content j_d_post_content ")
	// 					author := h.Find(".p_author_face ") // user page and avatar
	// 					authorpage, _ := author.Attr(hrefAttr)
	// 					datafield, _ := h.Attr(datafield)
	// 					m := make(map[string]string)
	// 					json.Unmarshal([]byte(datafield), m)

	// 					postid, _ := h.Attr("id")
	// 					pid := regex.FindString(postid)
	// 					s.Children().Each(func(i int, s *goquery.Selection) {
	// 						imgurl, _ := s.Attr(imgSrcAttr)
	// 						nurl := fmt.Sprintf(localImgUrlTemplate, postid, i)
	// 						s.SetAttr(imgSrcAttr, nurl)
	// 						dwlpath := path.Join(c.downloadRootPath, id, fmt.Sprintf(localImgUrlTemplate, pid, i))
	// 						req, _ := network.GetRequestFromUrl(imgurl)
	// 						imgtsk := downloader.NewTaskSaveToDisk(postid, *req, dwlpath, func(err error) {
	// 							if err != nil {
	// 								//TODO: fail
	// 							} else {
	// 								//TODO: success
	// 							}
	// 						})
	// 						c.downloader.AddTask(imgtsk)
	// 					})
	// 				})
	// 			}
	// 		}
	// 	}
	// }()
}
