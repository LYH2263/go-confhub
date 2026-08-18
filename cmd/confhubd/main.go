package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/LYH2263/go-confhub"
	"github.com/LYH2263/go-confhub/internal/api"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP 监听地址")
	web := flag.String("web", "web", "静态管理页目录")
	hmac := flag.String("hmac-secret", "", "可选全局 HMAC 密钥（原始字符串）")
	requireSign := flag.Bool("require-sign", false, "写入必须验签")
	flag.Parse()

	var opts []confhub.Option
	if *hmac != "" {
		opts = append(opts, confhub.WithHMACSecret([]byte(*hmac)))
	}
	if *requireSign {
		opts = append(opts, confhub.WithRequireSign(true))
	}
	h := confhub.New(opts...)
	defer h.Close()

	srv := api.New(h, api.Options{WebDir: *web, AllowCORS: true})
	httpSrv := &http.Server{Addr: *addr, Handler: srv}

	go func() {
		log.Printf("confhubd 监听 %s", *addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	_ = httpSrv.Close()
}
