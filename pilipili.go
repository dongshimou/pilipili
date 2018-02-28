package main

import (
	"./pilipili"
	"log"
	"os"
	"sync/atomic"
	"time"
)

func main() {

	//url:="https://www.bilibili.com/video/av16239259/"

	urls := os.Args[1:]
	var down_count int32 = 0
	for _, url := range urls {
		down := func(url string) {
			pili := pilipili.New()
			pili.Init(url)
			if pili.GetError() == nil {
				for pili != nil && pili.GetError() == nil {
					pili.DownloadDanmaku()
					pili.DownloadFlv()
					pili = pili.NextPage()
				}
			} else {
				log.Println(pili.GetError().Error())
			}
			atomic.AddInt32(&down_count, 1)
		}
		log.Println("start download task =============  ")
		go down(url)
	}

	for {
		if atomic.LoadInt32(&down_count) == int32(len(urls)) {
			break
		}
		time.Sleep(time.Second)
	}
}
