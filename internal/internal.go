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
			time.Sleep(1 * time.Second)
			if now.Hour() != 0 || now.Minute() != 0 || now.Second() != 0 {
				continue
			}

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
