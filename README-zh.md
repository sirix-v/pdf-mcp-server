# mcp服务器--阅读对应目录下的论文

### 编译
```
go build -o pdf-mcp-server main.go
```

### 安装

将pdf-mcp-server这个二进制文件所在的路径添加到环境变量里

以mac为例

```
export PATH=$PATH:你的路径/pdf-mcp-server
source ~/.zshrc
```


### 启动mcp服务器

```
pdf-mcp-server -pdfdir=你存放论文的路径
比如：pdf-mcp-server -pdfdir=/Users/sirix/lunwen
```

### cursor设置
```
{
  "mcpServers": {
    "pdf-mcp-server": {
      "url": "http://127.0.0.1:8080/sse"
    }
  }
}
```
### 设置Cursor Rules
例子
```
处理与PDF相关的请求时，调用名为“pdf-mcp-server”的mcp服务器的find_pdf工具列出/Users/sirix/lunwen目录下的所有论文，,在这个目录下查找对应文件名称的论文，并且使用pdf-mcp-server服务器的read_pdf工具进行阅读
```
