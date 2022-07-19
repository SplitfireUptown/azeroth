package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	strings "strings"
	"time"
) //导入fmt包

/**
通过反射实现json序列化，必须首字母大写
*/

type BaseVo struct {
	AbstractSql  string         `json:"abstract_sql"`
	QueryTimeArg float64        `json:"query_time_arg"`
	Total        int32          `json:"total"`
	SqlInfo      []SqlStatClass `json:"sql_info"`
}

type SqlStatClass struct {
	Sql          string `json:"sql"`
	StartTime    string `json:"start_time"`
	QueryTime    string `json:"query_time"`
	LockTime     string `json:"lock_time"`
	RowsSent     string `json:"rows_sent"`
	RowsExamined string `json:"rows_examined"`
}

//var resResult = map[string][]SqlStatClass{}

// 导出csv文件
//func downLoadExcel(filePath string) {
//	//filePath := `/Users/xiaofeili/code/azeroth-Sql/s1.log`
//	resResult := ReadEachLineReader(filePath)
//
//	f, _ := os.Create("Sql-analyzes-" + strconv.FormatInt(time.Now().Unix(), 10) + ".csv")
//	defer f.Close()
//
//	writer := csv.NewWriter(f)
//
//	writer.Write([]string{"abstract_sql", "real_sql", "query_time", "lock_Time"})
//
//	for k, v := range resResult {
//
//		for _, sqlStatClazz := range v {
//			writer.Write([]string{k, sqlStatClazz.Sql, sqlStatClazz.QueryTime, sqlStatClazz.LockTime})
//		}
//	}
//
//	writer.Flush() // 此时才会将缓冲区数据写入 }
//
//	fmt.Println("hello world")
//}

// ReadEachLineReader 读取慢sql文件的每一行
func ReadEachLineReader(filePath string) map[string][]SqlStatClass {

	var resResult = map[string][]SqlStatClass{}
	start1 := time.Now()
	FileHandle, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		return resResult
	}
	defer FileHandle.Close()
	lineReader := bufio.NewReader(FileHandle)

	var x = 0
	for {
		x++
		line, _, err := lineReader.ReadLine()
		if err == io.EOF {
			break
		}

		i := 3

		// set timestamp为一段sql分析的第一行
		if strings.HasPrefix(string(line), "SET timestamp=") {
			var startTime string
			var sqlLineStr string
			var abstractSql string
			var queryTimeLine string

			for {

				nextLine, _, err := lineReader.ReadLine()
				nextLineStr := string(nextLine)
				if err == io.EOF || nextLineStr == "" {
					break
				}

				if strings.HasPrefix(nextLineStr, "# User") {
					i--
					continue
				}

				if strings.HasPrefix(nextLineStr, "# Time:") {
					startTime = nextLineStr[8:len(nextLineStr)]
				}

				if strings.HasPrefix(nextLineStr, "select") || strings.HasPrefix(nextLineStr, "SELECT") {
					sqlLineStr = nextLineStr

					// 转换抽象sql
					abstractSql = Fingerprint(nextLineStr)
				}

				if strings.HasPrefix(nextLineStr, "# Query_time") {
					queryTimeLine = nextLineStr
				}
				i--
				if i <= 0 {
					break
				}
			}

			// 抽象sql为空舍弃
			if abstractSql == "" {
				continue
			}

			if queryTimeLine != "" {
				//fmt.Println("queryTimeLine -> " + string(queryTimeLine))
				queryTimeStr := queryTimeLine[strings.Index(queryTimeLine, "Query_time")+12 : strings.Index(queryTimeLine, "Lock_time")-1]
				lockTimeStr := queryTimeLine[strings.Index(queryTimeLine, "Lock_time")+11 : strings.Index(queryTimeLine, "Rows_sent")-1]
				rowsSentStr := queryTimeLine[strings.Index(queryTimeLine, "Rows_sent")+11 : strings.Index(queryTimeLine, "Rows_examined")-1]
				rowsExaminedStr := queryTimeLine[strings.Index(queryTimeLine, "Rows_examined")+15 : len(queryTimeLine)]

				var sqlStatTmp = SqlStatClass{
					sqlLineStr,
					startTime,
					queryTimeStr,
					lockTimeStr,
					rowsSentStr,
					rowsExaminedStr,
				}

				statsArray := resResult[abstractSql]
				if statsArray == nil {
					var statsNewArray = []SqlStatClass{
						sqlStatTmp,
					}
					resResult[abstractSql] = statsNewArray
				} else {
					statsArray = append(statsArray, sqlStatTmp)
					resResult[abstractSql] = statsArray
				}
			}
		}
	}
	fmt.Println("readEachLineReader spend : ", time.Now().Sub(start1))

	return resResult
}

// GetBetweenDates 根据开始日期和结束日期计算出时间段内所有日期
func GetBetweenDates(sdate, edate string) []string {
	d := []string{}
	timeFormatTpl := "2006-01-02 15:04:05"
	if len(timeFormatTpl) != len(sdate) {
		timeFormatTpl = timeFormatTpl[0:len(sdate)]
	}
	date, err := time.Parse(timeFormatTpl, sdate)
	if err != nil {
		// 时间解析，异常
		return d
	}
	date2, err := time.Parse(timeFormatTpl, edate)
	if err != nil {
		// 时间解析，异常
		return d
	}
	if date2.Before(date) {
		// 如果结束时间小于开始时间，异常
		return d
	}
	// 输出日期格式固定
	timeFormatTpl = "2006-01-02"
	date2Str := date2.Format(timeFormatTpl)
	d = append(d, date.Format(timeFormatTpl))
	for {
		date = date.AddDate(0, 0, 1)
		dateStr := date.Format(timeFormatTpl)
		d = append(d, dateStr)
		if dateStr == date2Str {
			break
		}
	}
	return d
}

func BuildTaskStr(taskArray []string) string {
	var finalTask = "grep -E '"

	for i, task := range taskArray {
		if i != 0 {
			finalTask = finalTask + "|" + task
		} else {
			finalTask += task
		}
	}

	finalTask += fmt.Sprintf(
		"' -B 5 %s > tmp%s.log",
		" /Users/xiaofeili/code/azeroth-Sql/haft-year.log",
		strconv.FormatInt(time.Now().Unix(), 10))

	return finalTask
}

func RunCommand(path, name string, arg ...string) (msg string, err error) {
	cmd := exec.Command(name, arg...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Dir = path
	err = cmd.Run()
	log.Println(cmd.Args)
	if err != nil {
		msg = fmt.Sprint(err) + ": " + stderr.String()
		err = errors.New(msg)
		log.Println(" ERROR CMD -> ", err.Error(), "cmd", cmd.Args)
	}
	return
}

func getCurrentPath() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}
