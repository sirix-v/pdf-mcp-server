package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"strings"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	"github.com/ledongthuc/pdf"
)

type listPDFsReq struct {
	Directory string `json:"directory" description:"directory path to list PDF files"`
}

type readPDFReq struct {
	FilePath string `json:"file_path" description:"path to the PDF file to read"`
}

var defaultPDFDir string

func init() {
	flag.StringVar(&defaultPDFDir, "pdfdir", "/Users/xiaojun/lunwen", "默认PDF目录")
}

func main() {
	flag.Parse()
	messageEndpointURL := "/message"

	sseTransport, mcpHandler, err := transport.NewSSEServerTransportAndHandler(messageEndpointURL)
	if err != nil {
		log.Panicf("new sse transport and hander with error: %v", err)
	}

	mcpServer, err := server.NewServer(sseTransport,
		server.WithServerInfo(protocol.Implementation{
			Name:    "pdf-mcp-server",
			Version: "1.0.0",
		}),
	)
	if err != nil {
		panic(err)
	}

	// 注册列出PDF文件的工具
	listPDFsTool, err := protocol.NewTool("list_pdfs", "List all PDF files in the specified directory", listPDFsReq{})
	if err != nil {
		panic(fmt.Sprintf("Failed to create list_pdfs tool: %v", err))
	}
	mcpServer.RegisterTool(listPDFsTool, listPDFs)

	// 注册读取PDF文件的工具
	readPDFTool, err := protocol.NewTool("read_pdf", "Read content of a PDF file", readPDFReq{})
	if err != nil {
		panic(fmt.Sprintf("Failed to create read_pdf tool: %v", err))
	}
	mcpServer.RegisterTool(readPDFTool, readPDF)

	router := http.NewServeMux()
	router.HandleFunc("/sse", mcpHandler.HandleSSE().ServeHTTP)
	router.HandleFunc(messageEndpointURL, mcpHandler.HandleMessage().ServeHTTP)

	httpServer := &http.Server{
		Addr:        ":8080",
		Handler:     router,
		IdleTimeout: time.Minute,
	}

	errCh := make(chan error, 3)
	go func() {
		errCh <- mcpServer.Run()
	}()

	go func() {
		if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	if err = signalWaiter(errCh); err != nil {
		panic(fmt.Sprintf("signal waiter: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	httpServer.RegisterOnShutdown(func() {
		if err = mcpServer.Shutdown(ctx); err != nil {
			panic(err)
		}
	})

	if err = httpServer.Shutdown(ctx); err != nil {
		panic(err)
	}
}

func listPDFs(_ context.Context, request *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	req := new(listPDFsReq)
	if err := protocol.VerifyAndUnmarshal(request.RawArguments, &req); err != nil {
		return nil, err
	}

	dir := req.Directory
	if dir == "" {
		dir = defaultPDFDir
	}

	var pdfFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".pdf" {
			pdfFiles = append(pdfFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Found %d PDF files:\n%s", len(pdfFiles), formatFileList(pdfFiles)),
			},
		},
	}, nil
}

func readPDF(_ context.Context, request *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	req := new(readPDFReq)
	if err := protocol.VerifyAndUnmarshal(request.RawArguments, &req); err != nil {
		return nil, err
	}

	// 检查文件是否存在
	if _, err := os.Stat(req.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", req.FilePath)
	}

	// 检查文件扩展名
	if filepath.Ext(req.FilePath) != ".pdf" {
		return nil, fmt.Errorf("file is not a PDF: %s", req.FilePath)
	}

	// 解析PDF内容
	f, r, err := pdf.Open(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open pdf: %v", err)
	}
	defer f.Close()

	var sb strings.Builder
	total := 0
	maxLen := 1 << 40 // 最多返回1TB
	for pageNum := 1; pageNum <= r.NumPage(); pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		if total+len(content) > maxLen {
			sb.WriteString(content[:maxLen-total])
			break
		}
		sb.WriteString(content)
		total += len(content)
	}

	text := sb.String()
	if len(text) == 0 {
		text = "未能提取到PDF文本内容。"
	}

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: text,
			},
		},
	}, nil
}

func formatFileList(files []string) string {
	if len(files) == 0 {
		return "No PDF files found"
	}

	result := ""
	for i, file := range files {
		result += fmt.Sprintf("%d. %s\n", i+1, file)
	}
	return result
}

func signalWaiter(errCh chan error) error {
	signalToNotify := []os.Signal{syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM}
	if signal.Ignored(syscall.SIGHUP) {
		signalToNotify = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, signalToNotify...)

	select {
	case sig := <-signals:
		switch sig {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
			log.Printf("Received signal: %s\n", sig)
			return nil
		}
	case err := <-errCh:
		return err
	}

	return nil
}
