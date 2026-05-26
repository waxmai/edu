package main

import (
	"fmt"
	"log"

	"edu-schedule-system/configs"
	"edu-schedule-system/internal/pkg/env"
)

func main() {
	env.Init()
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("env/config check failed: %v", err)
	}

	redisMode := "disabled"
	if cfg.Redis.Enabled {
		redisMode = "enabled"
	}
	fmt.Printf("env/config ok: env=%s port=%s mysql_read=%s mysql_write=%s redis=%s auth_mode=%s\n",
		env.Active().Value(),
		cfg.Server.Port,
		cfg.MySQL.Read.Addr,
		cfg.MySQL.Write.Addr,
		redisMode,
		cfg.Auth.Mode,
	)
}
