package main

import (
	"database/sql"
	"fmt"
	"github.com/antchfx/htmlquery"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type User struct {
	Title   string
	Date    string
	Writer  string
	Reader  string
	Article string
}

var DB *sql.DB

func getHtml(url_ string) string {
	req, _ := http.NewRequest("GET", url_, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/"+
		"537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")
	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(resp.Body)
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil && data == nil {
		log.Fatalln(err)
	}
	return fmt.Sprintf("%s", data)
}

func spider(row string) User {
	res := getHtml("https://news.fzu.edu.cn/" + row)
	root1, _ := htmlquery.Parse(strings.NewReader(res))
	trTitle := htmlquery.Find(root1, "//*[@id=\"main\"]/div[2]/form/div[1]/p[1]")
	trDate := htmlquery.Find(root1, "//*[@id=\"fbsj\"]")
	trWriter := htmlquery.Find(root1, "//*[@id=\"author\"]")
	goal1 := strings.Split(row, "/")
	goal2 := strings.Split(goal1[2], ".")
	goal := goal2[0]
	url_ := "https://news.fzu.edu.cn/system/resource/code/news/click/dynclicks.jsp?clickid=" + goal + "&owner=1744991928&clicktype=wbnews"
	html_ := getHtml(url_)
	root2, _ := htmlquery.Parse(strings.NewReader(html_))
	trReader := htmlquery.Find(root2, "/html/body/text()")
	trArticle := htmlquery.Find(root1, "//*[@id=\"news_content_display\"]")
	var users User
	users.Title = htmlquery.InnerText(trTitle[0])
	users.Date = htmlquery.InnerText(trDate[0])
	users.Writer = htmlquery.InnerText(trWriter[0])
	users.Reader = htmlquery.InnerText(trReader[0])
	users.Article = htmlquery.InnerText(trArticle[0])
	return users
}

func InsertDB(user User) {
	sql_ := "INSERT INTO fzu(Title, Writer, Reader, Date, Article) VALUES(?,?,?,?,?)" //插入数据库
	res, err := DB.Exec(sql_, user.Title, user.Writer, user.Reader, user.Date, user.Article)
	if err != nil {
		panic(err.Error())
	} //检测错误
	lastID, err := res.LastInsertId() //LastInsertID方法处理res错误的返回
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(lastID) //数据存储完成
}

func nextPage() {
	wg := sync.WaitGroup{} //设置一个等待组
	for j := 135; j > 107; j-- {
		url := fmt.Sprintf("https://news.fzu.edu.cn/fdyw/%d.htm", j)
		var urls []string
		html_ := getHtml(url)
		root, _ := htmlquery.Parse(strings.NewReader(html_))
		for i := 1; i < 41; i++ {
			xpath_ := fmt.Sprintf("//*[@id=\"main\"]/div[2]/div[2]/ul/li[%d]/a/@href", i)
			tr := htmlquery.Find(root, xpath_)
			str1 := htmlquery.InnerText(tr[0])
			str2 := strings.Split(str1, "/")
			hotels := str2[1] + "/" + str2[2] + "/" + str2[3]
			urls = append(urls, hotels)
		}
		for _, row1 := range urls {
			go func(row string) { //定义一个goroutine
				wg.Add(1) //等待组的计数器+1
				InsertDB(spider(row))
				wg.Done() //等待组的计数器-1
			}(row1)
		}
		wg.Wait() //等待所有的任务完成
	}
}

func run() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/GOLANG?charset=utf8mb4&parseTime=True&loc=Local"
	DB, _ = sql.Open("mysql", dsn)
	wg := sync.WaitGroup{}
	url := "https://news.fzu.edu.cn/fdyw.htm"
	var NewsUrl []string
	html_ := getHtml(url)
	root, _ := htmlquery.Parse(strings.NewReader(html_))
	for i := 1; i < 41; i++ {
		xpath_ := fmt.Sprintf("//*[@id=\"main\"]/div[2]/div[2]/ul/li[%d]/a/@href", i)
		tr := htmlquery.Find(root, xpath_)
		NewsUrl = append(NewsUrl, htmlquery.InnerText(tr[0]))
	}
	for _, row1 := range NewsUrl {
		go func(row string) {
			wg.Add(1)
			InsertDB(spider(row))
			wg.Done()
		}(row1)
	}
	wg.Wait()
	nextPage()
}

func main() {
	run()
}
