远程图片转化webp服务
客户端向服务端发送图片,服务端转化图片返回,客户端接收并保存
主要解决71服务器版本太低没法直接升级服务器

go build -o webpclient client.go
go build -o webpserver server.go