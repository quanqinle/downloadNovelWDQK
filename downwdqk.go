//
// downwdqk.go
//
// down novel "WuDongQianKun" from webpage http://www.gd560.com/
//
package main

import (
	"fmt"
)

const (
	BASEURL = "http://www.gd560.com/"
	COLUMN  = "ooxx"
	START   = "第一章"
	END     = "第二章"
	FILE    = "wdqk.txt"

	DEBUG = true
)

func debug(msg string) {
	if DEBUG {
		//log.Println("[debug msg]", msg)
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("[debug msg]%s:%3d(%s) %v\n", file, line, runtime.FuncForPC(pc).Name(), msg)
		}
	}
}

func GetAllPageUrl(url string) section_urls []string {

	return
}

func SaveTofile(section_urls []string) {
	for _, section_url := range section_urls {
		debug("section_url :" + section_url) //下载图片：http://ww2.sinaimg.cn/mw600/8upmvnj.jpg

		//解析图片的名称
		re3 := regexp.MustCompile(`.+?/`)
		img_name := re3.ReplaceAllString(section_url, "") //什么意思，为什么字符串是空的？
		debug("img_file:" + img_name)

		urlRetrieve(section_url, img_name)
	}

}

func main() {
	SaveTofile(GetAllPageUrl(url))
}
