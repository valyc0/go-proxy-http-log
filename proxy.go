package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var (
	listenAddr  string
	targetAddr  string
	logFilePath string
	logFile     *os.File
	targetURL   *url.URL
)

func init() {
	flag.StringVar(&listenAddr, "listen", ":8090", "Address and port to listen on")
	flag.StringVar(&targetAddr, "target", "http://example.com", "Target server address")
	flag.StringVar(&logFilePath, "log", "proxy.log", "Path to log file")
	flag.Parse()

	if len(os.Args) == 1 {
		printUsageAndExit()
	}

	var err error
	targetURL, err = url.Parse(targetAddr)
	if err != nil {
		log.Fatalf("Invalid target address: %v", err)
	}

	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

func main() {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.Host = targetURL.Host
		},
		ModifyResponse: logResponse,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		proxy.ServeHTTP(w, r)
	})

	fmt.Printf("Starting proxy server on %s\n", listenAddr)
	fmt.Printf("Forwarding requests to %s\n", targetAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func logRequest(r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("Error dumping request: %v", err)
		return
	}
	log.Printf("Request:\n%s\n", string(dump))
}

func logResponse(resp *http.Response) error {
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Printf("Error dumping response: %v", err)
		return nil
	}
	log.Printf("Response:\n%s\n", string(dump))
	return nil
}

func printUsageAndExit() {
	fmt.Println("Usage of the proxy server:")
	fmt.Println("  -listen : Address and port to listen on (default: :8090)")
	fmt.Println("  -target : Target server address to forward requests to (default: http://example.com)")
	fmt.Println("  -log    : Path to log file (default: proxy.log)")
	fmt.Println("\nExample:")
	fmt.Println("  ./proxy -listen :8080 -target http://localhost:3000 -log mylogfile.log")
	os.Exit(0)
}
