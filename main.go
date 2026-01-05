package main

import (
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	_ "jh_game_service/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"jh_game_service/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.New())
}
