package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/levigross/grequests"
	"github.com/tidwall/gjson"
	"log"
	"time"
)

type Comment struct {
	uname   string
	likes   string
	content string
	date    string
}

var DBB *sql.DB

func insertDB(comments Comment) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/GOLANG?charset=utf8mb4&parseTime=True&loc=Local"
	DBB, _ = sql.Open("mysql", dsn)
	sql_ := "INSERT INTO bilibili(uname,likes,content,date) VALUE(?,?,?,?)"
	res, err := DBB.Exec(sql_, comments.uname, comments.likes, comments.content, comments.date)
	if err != nil {
		panic(err.Error())
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(lastID)
}

func getComment() {
	var pg int64 = 1
	for {
		urls := fmt.Sprintf("https://api.bilibili.com/x/v2/reply?pn=%d&type=1&oid=54737593", pg)
		res, _ := grequests.Get(urls, nil)
		body := res.String()
		if !gjson.Valid(body) {
			log.Fatalln("error")
		}
		length := gjson.Get(body, "data.replies.#").Int()
		l := 1000 / length
		if pg > l {
			break
		} else {
			var i int64
			for i = 1; i < length; i++ {
				getMainComment(body, i)
				commentPath := fmt.Sprintf("data.replies.%d.replies", i)
				childLength := gjson.Get(body, commentPath).Int()
				var j int64
				for j = 1; j < childLength; j++ {
					getChildComment(body, i, j)
				}
			}
		}
		pg = pg + 1
	}
	deleteDB()
}

var bili Comment

func getMainComment(body string, i int64) {
	pathUser := fmt.Sprintf("data.replies.%d.member", i)
	pathArticle := fmt.Sprintf("data.replies.%d.content", i)
	pathLike := fmt.Sprintf("data.replies.%d.like", i)
	pathTime := fmt.Sprintf("data.replies.%d", i)
	dataUser := gjson.Get(body, pathUser)
	dataLike := gjson.Get(body, pathLike)
	dataArticle := gjson.Get(body, pathArticle)
	resTime := gjson.Get(body, pathTime).Map()["ctime"]
	tm := time.Unix(resTime.Int(), 0)
	dataTime := tm.Format("2006-01-02 15:04:05")
	bili.uname = dataUser.Map()["uname"].String()
	bili.likes = dataLike.String()
	bili.content = dataArticle.Map()["message"].String()
	bili.date = dataTime
	insertDB(bili)
}

func getChildComment(body string, i int64, j int64) {
	childPath := fmt.Sprintf("data.replies.%d.replies.%d", i, j)
	childData := gjson.Get(body, childPath)
	childUserName := childData.Map()["member"]
	goalUserName := childUserName.Map()["uname"]
	childLike := gjson.Get(body, childPath)
	childJson := childData.Map()["content"]
	goalComment := childJson.Map()["message"]
	childTime1 := gjson.Get(body, childPath).Map()["ctime"]
	childTm := time.Unix(childTime1.Int(), 0)
	childTime := childTm.Format("2006-01-02 15:04:05")
	bili.uname = goalUserName.String()
	bili.likes = childLike.Map()["like"].String()
	bili.content = goalComment.String()
	bili.date = childTime
	insertDB(bili)

}

func deleteDB() {
	deleteSQL := "DELETE FROM bilibili where uname=''"
	res, err := DBB.Exec(deleteSQL)
	if err != nil {
		panic(err.Error())
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		log.Fatal("err")
	}
	fmt.Print(lastID)
}

func main() {
	getComment()
	DBB.Close()
}
