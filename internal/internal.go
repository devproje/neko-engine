package internal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/devproje/neko-engine/common"
	"github.com/devproje/neko-engine/internal/route"
	"github.com/gin-gonic/gin"
)

func NewInternalServer(sl *common.ServiceLoader) {
	app := gin.Default()
	route.InternalRouter(app, sl)

	if err := app.Run(":8081"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc

	fmt.Println("Shutting down internal server...")
}
