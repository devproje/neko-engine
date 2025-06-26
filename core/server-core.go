package core

import (
	"fmt"
	"os"

	"github.com/devproje/neko-engine/common"
	"github.com/devproje/neko-engine/core/route"
	"github.com/gin-gonic/gin"
)

func NewServerCore(sl *common.ServiceLoader) {
	app := gin.Default()
	route.CoreRouter(app, sl)

	if err := app.Run(":8080"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
}
