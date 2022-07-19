package router

import (
	"bytes"
	"encoding/csv"
	"flash-turtle/core"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
	"strings"
)

func IndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func DataList2(c *gin.Context) {
	c.HTML(http.StatusOK, "data_list.html", gin.H{})
}

func DownLoadData(c *gin.Context) {

	logPath := c.Query("log_path")
	resResult := core.ReadEachLineReader(logPath)

	headList := []string{"abstract_sql", "real_sql", "start_time", "query_time", "lock_time", "rows_sent", "rows_examined"}

	dataBytes := new(bytes.Buffer)
	dataBytes.WriteString("\xEF\xBB\xBF")
	wr := csv.NewWriter(dataBytes)

	wr.Write(headList)

	for k, v := range resResult {

		for _, sqlStatClazz := range v {
			wr.Write([]string{k, sqlStatClazz.Sql, sqlStatClazz.StartTime, sqlStatClazz.QueryTime, sqlStatClazz.LockTime, sqlStatClazz.RowsSent, sqlStatClazz.RowsExamined})
		}
	}

	wr.Flush() // 此时才会将缓冲区数据写入 }

	fmt.Println("hello world")
	wr.Flush()

	c.Writer.Header().Set("Content-type", "application/octet-stream")
	//c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", "analyzes.csv"))
	c.String(200, dataBytes.String())
}

func DataFetch(c *gin.Context) {

	logPath := c.PostForm("log_path")

	var resList []core.BaseVo
	var resResult = map[string][]core.SqlStatClass{}
	//resResult = core.ReadEachLineReader(`/Users/xiaofeili/code/azeroth/haft-year.log`)
	resResult = core.ReadEachLineReader(logPath)

	for k, v := range resResult {

		var queryTimeTotal = 0.0
		var tmpTotal = 0

		for _, sqlStatClazz := range v {

			queryTimeIntValueRes, _ := decimal.NewFromString(strings.Replace(sqlStatClazz.QueryTime, " ", "", -1))
			queryTimeIntValue, _ := queryTimeIntValueRes.Float64()

			queryTimeTotal += queryTimeIntValue
			tmpTotal++
		}

		queryTimeArg := queryTimeTotal / float64(len(v))

		var vo = core.BaseVo{
			k,
			queryTimeArg,
			int32(tmpTotal),
			v,
		}
		resList = append(resList, vo)
	}

	c.JSON(http.StatusOK, resList)
}

func InitRouter(r *gin.Engine) *gin.Engine {

	//静态资源
	r.Static("/static", "./static")

	//页面路由
	r.LoadHTMLGlob("templates/*")
	r.GET("/index", IndexPage) //加入的页面
	r.GET("/", IndexPage)      //加入的页面

	r.GET("/data_list", DataList2)

	//接口路由
	r.POST("/fetch/data", DataFetch)
	r.GET("/download/data", DownLoadData)

	return r
}
