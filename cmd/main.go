package main

import (
	"flag"
	"log"

	"github.com/statnett/image-scanner-operator/internal/config"
	"github.com/statnett/image-scanner-operator/internal/operator"
)

func main() {
	cfg := config.Config{}
	cfg.Zap.Development = true

	opr := operator.Operator{}
	if err := opr.BindConfig(&cfg, flag.CommandLine); err != nil {
		log.Fatal(err)
	}

	if err := opr.ValidateConfig(cfg); err != nil {
		log.Fatal(err)
	}

	if err := opr.Start(cfg); err != nil {
		log.Fatal(err)
	}
}
