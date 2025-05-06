# MCP Server -- Read Papers in the Specified Directory

### Installation

Add the path where the pdf-mcp-server binary is located to your environment variable.

Take macOS as an example:

```
export PATH=$PATH:/your/path/pdf-mcp-server
source ~/.zshrc
```

### Start the MCP Server

```
pdf-mcp-server -pdfdir=your_paper_directory
For example: pdf-mcp-server -pdfdir=/Users/sirix/lunwen
```

### Cursor Settings
```
{
  "mcpServers": {
    "pdf-mcp-server": {
      "url": "http://127.0.0.1:8080/sse",
      "directory": "/Users/sirix/lunwen/"
    }
  }
}
```
### Set Cursor Rules
Example:
```
When handling PDF-related requests, call the list_pdfs tool of the mcp server named "pdf-mcp-server" to list all papers in the /Users/sirix/lunwen directory, search for the paper with the corresponding file name in this directory, and use the read_pdf tool of the pdf-mcp-server to read it.
``` 
### 设置Cursor Rules
例子
```
处理与PDF相关的请求时，调用名为“pdf-mcp-server”的mcp服务器的list_pdfs工具列出directory下的所有论文，,在这个目录下查找对应文件名称的论文，并且使用pdf-mcp-server服务器的read_pdf工具进行阅读
```
