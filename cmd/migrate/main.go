package main

import (
	"fmt"
	"os"
	"strconv"

	"lockbox/internal/config"
	"lockbox/internal/database/migration"
	"lockbox/internal/logger"
)

func main() {
	env, err := config.NewMigrationEnv()
	if err != nil {
		panic(fmt.Sprintf("migrate error: %v", err))
	}

	log, err := logger.NewLogger(&env.App)
	if err != nil {
		panic(fmt.Sprintf("log error: %v", err))
	}

	mig := migration.New(env)

	if len(os.Args) < 1 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "bootstrap":
		if err := mig.Bootstrap(); err != nil {
			log.Error("bootstrap error", err, nil)
			os.Exit(1)
		}
		fmt.Println("bootstrap success!")
	case "up":
		if err := mig.Up(); err != nil {
			log.Error("up error", err, nil)
			os.Exit(1)
		}
		fmt.Println("up success!")

	case "down":
		if err := mig.Down(); err != nil {
			log.Error("down error", nil, nil)
			os.Exit(1)
		}
		fmt.Println(" down success!")

	case "version":
		v, _, err := mig.Version()
		if err != nil {
			log.Error("version error", err, nil)
			os.Exit(1)
		}
		fmt.Printf("current version: %d\n", v)
	case "force":
		if len(os.Args) != 3 {
			log.Error("force should be: force <version> ", err, nil)
			os.Exit(1)
		}

		v, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Error("version should be a number", err, nil)
			os.Exit(1)
		}

		if err := mig.Force(v); err != nil {
			log.Error("force error", err, nil)
		}
		fmt.Printf("force %d success!", v)

	default:
		log.Info("command not found. Please use:", nil)
		usage()
	}
}

func usage() {
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  migrate bootstrap")
	fmt.Println("  migrate up")
	fmt.Println("  migrate down")
	fmt.Println("  migrate version")
	fmt.Println("  migrate force <version>")
	fmt.Println()
}
