package main

import (
	"fmt"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	_ "jh_game_service/internal/logic"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"

	"jh_game_service/internal/cmd"
)

func main() {
	// 在最早的时刻设置时区
	ctx := gctx.New()
	timezone := g.Cfg().MustGet(ctx, "timezone").String()
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		fmt.Printf("加载时区失败: %v，使用默认时区\n", err)
		loc, _ = time.LoadLocation("Asia/Shanghai")
	}
	time.Local = loc
	fmt.Printf("时区设置: %s\n", timezone)

	cmd.Main.Run(ctx)
}
