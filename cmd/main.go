package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ChanceYoungWebDevelopment/ActionTargetSkillAssessment/internal/config"
	"github.com/ChanceYoungWebDevelopment/ActionTargetSkillAssessment/internal/monitor"
	"github.com/ChanceYoungWebDevelopment/ActionTargetSkillAssessment/internal/web"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	log.Printf("at-ping starting: hosts=%d interval=%v timeout=%v port=%d privileged=%v window=%d push=%v",
		len(cfg.Hosts), cfg.Interval, cfg.Timeout, cfg.Port, cfg.Privileged, cfg.Window, cfg.PushInterval)
	addr := "localhost:" + strconv.Itoa(cfg.Port)
	log.Printf("Web Dash: http://%s\n", addr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mgr := monitor.NewManager(cfg.Window, cfg.DownAfter)
	if err := mgr.Start(ctx, cfg); err != nil {
		log.Fatalf("monitor start: %v", err)
	}



	srv := web.NewServer(web.Options{
		Addr:         cfg.ListenAddr(),
		PushInterval: cfg.PushInterval,
		//StaticDir: cfg.WebDir //optional dev override
	}, mgr)

	go func() {
		if err := srv.Serve(); err != nil {
			log.Printf("http server exit: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shCtx)
	log.Println("at-ping stopped")
}
