# MCP Server -- Read Papers in the local Specified Directory
[中文文档](https://github.com/sirix-v/pdf-mcp-server/blob/master/README-zh.md)

### Compile
```
go build -o pdf-mcp-server main.go
```

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
      "url": "http://127.0.0.1:8080/sse"
    }
  }
}
```
### Set Cursor Rules
Example:
```
When handling PDF-related requests, call the find_pdf tool of the mcp server named "pdf-mcp-server" to list all papers in the /Users/sirix/lunwen directory, search for the paper with the corresponding file name in this directory, and use the read_pdf tool of the pdf-mcp-server to read it.
``` 
