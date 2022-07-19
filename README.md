# azeroth

慢sql分析

- 基于MySQL自带慢sql统计功能输出的slow-query.log分析
- 使用教程：https://wiki.haizhi.com/pages/viewpage.action?pageId=83895026
- 编译
  - 在入口main.go同级目录下执行
  - ``CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o azeroth_linux main.go`` (linux运行环境)
  - ``CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o azeroth_mac main.go`` (mac运行环境)
  - ``CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o azeroth_windows.exe main.go`` (windows运行环境)