package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/ezhdanovskiy/postgres-go-experiments/internal/config"
	"github.com/ezhdanovskiy/postgres-go-experiments/internal/listener"
	"github.com/ezhdanovskiy/postgres-go-experiments/internal/notifier"
)

func main() {
	var component = flag.String("component", "", "Run one component")
	flag.Parse()

	run(*component)
	log.Printf("Bye bye!")
}

type Application interface {
	Run() error
	Stop()
}

func run(comp string) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("new logger: %s", err)
	}
	logger := zapLogger.Sugar()

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatalf("new logger: %s", err)
	}

	var app Application

	switch comp {
	case "notifier":
		ntf, err := notifier.NewNotifier(logger, cfg)
		if err != nil {
			logger.Fatalf("%s", err)
		}
		app = ntf
	case "listener":
		ntf, err := listener.NewListener(logger, cfg)
		if err != nil {
			logger.Fatalf("%s", err)
		}
		app = ntf
	default:
		flag.Usage()
		os.Exit(2)
	}

	go shutdownMonitor(app)

	err = app.Run()
	if err != nil {
		logger.Fatalf("app run: %s", err)
	}
}

func shutdownMonitor(app Application) {
	stopping := make(chan os.Signal)
	signal.Notify(stopping, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-stopping

	app.Stop()
}
