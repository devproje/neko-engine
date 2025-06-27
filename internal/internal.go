package internal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devproje/neko-engine/common"
	"github.com/devproje/neko-engine/internal/route"
	"github.com/gin-gonic/gin"
)

func NewInternalServer(sl *common.ServiceLoader) {
	app := gin.Default()
	route.InternalRouter(app, sl)

	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(24 * time.Hour)
			time.Sleep(time.Until(next))

			fmt.Println("Resetting count at midnight...")
			if err := sl.Acc.Acc.ResetCount(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}
	}()

	if err := app.Run(":8081"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc

	fmt.Println("Shutting down internal server...")
}
