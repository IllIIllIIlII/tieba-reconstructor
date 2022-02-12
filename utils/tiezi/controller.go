package thread

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"tieba-reconstructor/config"
	"tieba-reconstructor/models"
	vector "tieba-reconstructor/utils/containers"
	"tieba-reconstructor/utils/downloader"
	network "tieba-reconstructor/utils/net"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

const (
	localImgUrlTemplate  string = "./images/%s_%d.jpg"
	imgSrcAttr           string = "src"
	hrefAttr             string = "href"
	datafieldstr         string = "data-field"
	datapid              string = "data-pid"
	floorLocation        string = ".l_post j_l_post l_post_bright noborder "
	floorContentLocation string = "d_post_content j_d_post_content  clearfix"
	tmpFolderRoot        string = "./tmp/"
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
	logger           *logrus.Logger
}

// return a controller
func NewThreadController() *ThreadController {
	root, _ := os.Getwd()
	return &ThreadController{
		threads:          *vector.New(),
		downloader:       *downloader.New(),
		phtmls:           make(chan pageInfo, 4),
		downloadRootPath: root,
		logger: &logrus.Logger{
			Out:       os.Stderr,
			Formatter: new(logrus.TextFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.InfoLevel,
		},
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
	var thread *models.TiebaThread
	var floors *[]models.Floor
	thread.Id = t.Id
	thread.FloorNumber = t.FloorLimit
	thread.Floors = *floors

	//添加贴子下载任务

	var url string
	url = t.nextPageUrl()
	if url != "" {
		pagenum := t.currentPage
		c.logger.Infoln("Start downloading thread ", thread.Id, ", page ", pagenum)
		req, _ := network.GetRequestFromUrl(url)
		b := &[]byte{}
		thisThreadFolder := filepath.Join(tmpFolderRoot, t.Id)

		var handleResult func(error)
		handleResult = func(err error) {
			end := false
			if err != nil {
				// 404则封顶，否则中断
				_, ok := err.(downloader.Err404)
				if !ok {
					t.Status = config.Failed
					return
				} else {
					end = true
				}
			} else { //提取内容
				t.currentPage += 1
				doc, err := goquery.NewDocumentFromReader(bytes.NewReader(*b))
				if err != nil {
					t.FailedPage.Append(pagenum)
				} else {
					forumName := ""
					hasForumName := false
					doc.Find("meta").Each(func(i int, s *goquery.Selection) {
						if hasForumName {
							return
						}
						fname, ok := s.Attr("name")
						if ok {
							forumName = fname
						}
						hasForumName = hasForumName || ok
					})
					if !hasForumName {
						c.logger.Fatalf("No forum name found in meta tags of thread %s !", thread.Id)
					}
					t.ForumName = forumName
					//对所有限制内的楼层
					//TODO: 处理两套class
					doc.Find(floorLocation).Each(func(i int, h *goquery.Selection) {

						imageFolder := filepath.Join(thisThreadFolder, "images")
						str, ok := h.Attr(datafieldstr)
						var floor int64
						if !ok {
							c.logger.Fatal("thread %s page %s has a floor without data-field!", t.Id, t.currentPage)
						} else {
							var postnumreg *regexp.Regexp = regexp.MustCompile("\"post_no\":\\d*,")
							floorData := postnumreg.FindString(str)
							intRegex := regexp.MustCompile(`\d*`)
							floorStr := intRegex.FindString(floorData)
							floor, _ = strconv.ParseInt(floorStr, 10, 32)
							if floor == int64(t.FloorLimit) { //当前超过楼层限制就停止
								end = true
							}
						}
						// 按需保存作者信息,名字，头像
						author := h.Find(".icon_relative j_user_card")
						author.Each(func(i int, h *goquery.Selection) {
							img := h.Find("img")
							if img.Length() > 0 {
								c.logger.Fatal("thread %s page %s floor %s does not have author face!", t.Id, t.currentPage, floor)
							}
							username := img.AttrOr("username", "unknown")
							avatar, ok := img.Attr("src")
							if ok {
								// 下载头像
								//TODO: 增加存入服务器模式下的获取保存路径
								filename := username + ".jpg"
								h.SetAttr("src", path.Join(imageFolder, filename))
								req, _ := network.GetRequestFromUrl(avatar)
								c.downloader.AddTask(downloader.NewTaskSaveToDisk(username, *req, imageFolder, username+".jpg", func(err error) {
									if err != nil {
										_, ok := err.(downloader.Err404)
										if !ok {
											t.lock.Lock()
											t.FailedImage[avatar] = username + ".jpg"
											t.lock.Unlock()
										}
									}
								}))
							}
						})
						val, ok := author.Attr(datafieldstr)
						if !ok {
							c.logger.Fatalf("thread %s page %s floor %s has no author data!", t.Id, t.currentPage, floor)
						}
						authorData := models.User{}
						err := json.Unmarshal([]byte(val), &authorData)
						if err != nil {
							c.logger.Fatalf("thread %s page %s floor %s has author data error!\n%s", t.Id, t.currentPage, floor, val)
						}
						// 保存贴子内容
						content := h.Find(floorContentLocation)
						if content.Length() == 0 {
							c.logger.Fatalf("thread %s page %s floor %s has no content!", t.Id, t.currentPage, floor)
						}
						content.Each(func(i int, h *goquery.Selection) {
							// 图片
							h.Find("img").Each(func(i int, h *goquery.Selection) {
								imgsrc, ok := h.Attr("src")
								if ok {
									// 下载图片
									filename := filepath.Base(imgsrc) // 提取最后一个/后的内容作为文件名
									//重定向图片链接到本地链接
									h.SetAttr("src", path.Join(imageFolder, filename))
									imgDwlUrl := "http://tiebapic.baidu.com/forum/pic/item/" + filename
									req, _ := network.GetRequestFromUrl(imgsrc)
									c.downloader.AddTask(downloader.NewTaskSaveToDisk(imgDwlUrl, *req, imageFolder, filename, func(err error) {
										if err != nil {
											_, ok := err.(downloader.Err404)
											if !ok {
												t.lock.Lock()
												t.FailedImage[imgDwlUrl] = filename
												t.lock.Unlock()
											}
										}
									}))
								}
							})
						})
						floorContent := content.Text()
						floorInfo := models.Floor{
							FloorNumber: int(floor),
							Content:     floorContent,
							Author:      authorData,
						}
					})
				}
			} // 爬完本页后
			if !end {
				// 添加下一贴子
				url = t.nextPageUrl()
				if url != "" {
					pagenum := t.currentPage
					c.logger.Println("Start downloading thread ", thread.Id, ", page ", pagenum)
					req, _ = network.GetRequestFromUrl(url)
					b := &[]byte{}
					tsk := downloader.NewTaskToBytes(t.Id, *req, b, handleResult)
					c.downloader.AddTask(tsk)
				}
			} else {
				//无遗漏则成功，遗漏则失败
			}
		}
		//添加贴子页面i下载任务
		tsk := downloader.NewTaskToBytes(t.Id, *req, b, handleResult)
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
