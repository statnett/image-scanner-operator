package main

import (
	"flag"

	"github.com/statnett/image-scanner-operator/internal/config"
	"github.com/statnett/image-scanner-operator/internal/operator"
)

func main() {
	cfg := config.Config{}
	cfg.Zap.Development = true

	opr := operator.Operator{}
	opr.BindConfig(&cfg, flag.CommandLine)
	opr.ValidateConfig(cfg)
	opr.Start(cfg)
}
