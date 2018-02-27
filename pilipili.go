package main

import (
	"./tool"
	"log"
	"os"
	"sync/atomic"
	"time"
)

func main() {

	urls := os.Args[1:]
	var down_count int32 = 0
	for _, url := range urls {
		down := func(url string) {
			pili := tool.PiliPili()
			pili.Init(url)
			if pili.GetError() == nil {
				pili.DownloadDanmaku()
				pili.DownloadFlv()
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
