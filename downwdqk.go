// 作者: 权芹乐
// 文件: downwdqk.go
// 日期：2013-4-18
// 描述: 从网页http://www.gd560.com/中下载小说《武动乾坤》，并保存到本地txt文件。
// go版本: go1.0.3
package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	BASEURL   = "http://www.gd560.com/" //武动乾坤章节目录
	START     = "第一章"                   //TODO：后期需求：下载制定章节的内容
	END       = "第二章"
	FILEBOOK  = "wdqk-nowtime.txt"
	FILEINDEX = "wdqk-index-nowtime.txt"

	DEBUG = false
)

// 章节数据类型：章节名称、章节url
type sectionInfo struct {
	name string
	url  string
}

// 用于调整正文中的样式
var mTransprot = map[string]string{ //TODO：改写从配置文件读取
	"</a>":              "",
	"<p>":               "",
	"</span>":           "",
	"</p>":              "\n",
	"<br />":            "\n",
	"<br/>":             "\n",
	`<p class="p1">`:    "",
	`<span class="s1">`: "",
	"\n　\n":             "\n",
	"\n　　\n":            "\n",
	"\n\n":              "\n",
	"\n\n\n":            "\n",
}

func debug(msg string) {

	if DEBUG {
		//log.Println("[debug msg]", msg)
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("[debug msg]%s:%3d(%s) %v\n", file, line, runtime.FuncForPC(pc).Name(), msg)
		}
	}
}

// 从书籍url获取所有章节的信息
func GetSectionUrl(url string) (sections []sectionInfo, err error) {

	//读取页面消息主体
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// debug(string(body))

	//去掉HTML消息体的空白字符,替换为空格。
	re := regexp.MustCompile(`\s`)
	bodystr := re.ReplaceAllString(string(body), " ")
	// debug(bodystr)

	//查找<div class="box"><ul>.*</ul></div>间内容
	re = regexp.MustCompile(`<div class="box">.*<div class="box">`)
	li_href := re.Find([]byte(bodystr))
	// debug(string(li_href))

	//解析<li><a href="http://www.gd560.com/1.html" title="第一章 林动">第一章 林动</a></li>
	re = regexp.MustCompile(`<li><a href="(.+?)" title="(.+?)">第`)
	sectInfosTemp := re.FindAllSubmatch(li_href, -1)
	debug(string(sectInfosTemp[0][0]))    //<li><a href="http://www.gd560.com/1.html"
	debug(string(sectInfosTemp[0][1]))    //http://www.gd560.com/1.html
	debug(string(sectInfosTemp[0][1][0])) //h
	debug(string(sectInfosTemp[0][1][1])) //t
	debug(string(sectInfosTemp[0][2]))    //第一章
	// debug(string(sectInfosTemp[0][3])) //panic: runtime error: index out of range
	debug(string(sectInfosTemp[1][1])) //http://www.gd560.com/2.html

	for _, sect_info := range sectInfosTemp {
		debug("name:" + string(sect_info[2]) + " downpag:" + string(sect_info[1]))
		// debug("[page url][0]" + string(sect_info[0])) //page body li (s) :<li>...
		// debug("[page url][1]" + string(sect_info[1])) //http://www.gd560.com/1.html
		// debug("[page url][2]" + string(sect_info[2])) //第一千两百四十四章 再回异魔域
		//only 3 mumber,sect_info[3] cause runtime err!
		name, url := string(sect_info[2]), string(sect_info[1])
		sections = append(sections, sectionInfo{name, url})
		// mSectionInfo[string(sect_info[2])] = string(sect_info[1]) 
		// map数据结构不能满足要求，因为range该map时获得值的顺序是随机的。
	}

	return sections, nil
}

// 获取url对应的章节正文
func GetSectionText(url string) (byteText []byte, err error) {

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("err:", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	byteText = HtmlToText(body)
	return
}

// 转化、去除章节内容中的html编码
func HtmlToText(byteHtml []byte) (byteText []byte) {

	strHtml := html.UnescapeString(string(byteHtml))

	re := regexp.MustCompile(`\s`) //quanql:
	strHtml = re.ReplaceAllString(strHtml, " ")

	strHtml = strings.ToLower(strHtml)

	re = regexp.MustCompile(`<a href="(.+?)" target="_blank">`)
	strHtml = re.ReplaceAllString(strHtml, "")

	re = regexp.MustCompile(`<a target="_blank" href="(.+?)">`)
	strHtml = re.ReplaceAllString(strHtml, "")

	debug(strHtml)

	// re = regexp.MustCompile(`<div class=top>(.+?)<p align="center">`)
	// byteText = re.Find([]byte(strHtml))
	// 这种方式不能满足要求

	re = regexp.MustCompile(`<div class=top>(.+?)<p align="center">`)
	arrTextBody := re.FindAllSubmatch([]byte(strHtml), -1)
	byteText = arrTextBody[0][1]
	// debug(string(arrTextBody[0][0]))
	// debug(string(arrTextBody[0][1]))

	// 这部分代码不能往前调了，会影响div部分的正则查询
	strHtml = string(byteText)
	for old, new := range mTransprot {
		strHtml = strings.Replace(strHtml, old, new, -1)
	}
	byteText = []byte(strHtml)
	debug(string(byteText))

	return byteText
}

// 保存所有章节正文到同一个文件
func SaveBook(sections []sectionInfo) error {

	var book string

	fout, err := os.Create(FILEBOOK)
	if err != nil {
		fmt.Println("err:", err)
		return err
	}

	defer func() {
		fout.Close()
	}()

	// 将所有章节正文保存到内存
	for _, sectinfo := range sections {
		re := regexp.MustCompile(`(.+?)title="`)
		name := re.ReplaceAllString(sectinfo.name, "")

		fmt.Printf("downloading section[%s]...\n", name)

		body, err := GetSectionText(sectinfo.url)
		if err != nil {
			fmt.Println("err:", err)
			return err
		}

		book = book + fmt.Sprintf("%s\n%s\n\n", name, string(body))
	}

	// 将内存中的章节正文保存到文件
	fmt.Println("saving the book...")
	_, err = fout.WriteString(book)
	if err != nil {
		fmt.Println("err:", err)
		return err
	}

	return nil
}

// 保存一份与正文对应的章节目录，以便排查下载的书籍是否完整
func SaveBookCatalog(sections []sectionInfo) error {

	fmt.Println("saving the novel catalog...")

	var index string

	fout, err := os.Create(FILEINDEX)
	if err != nil {
		fmt.Println("err:", err)
		return err
	}
	defer fout.Close()

	// 将所有章节目录信息保存到内存
	for i, sectinfo := range sections {
		index = index + fmt.Sprintf("id:%5d | section:%s\n", i, sectinfo.name)
	}

	_, err = fout.WriteString(index)
	if err != nil {
		fmt.Println("savebookCatalog err:", err)
		return err
	}

	return nil
}

func main() {
	timestamp := time.Now() //TODO：用时间戳替换文件名中的nowtime
	fmt.Println("[开始]", timestamp)
	// FILEBOOK = strings.Replace(FILEBOOK, "nowtime", timestamp, -1)
	// FILEINDEX = strings.Replace(FILEINDEX, "nowtime", timestamp, -1)
	sections, err := GetSectionUrl(BASEURL)
	if err != nil {
		fmt.Println("main err:", err)
		os.Exit(1)
	}
	SaveBookCatalog(sections)
	SaveBook(sections)
	fmt.Println("[结束]", time.Now())
}
