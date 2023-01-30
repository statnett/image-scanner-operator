package main

import (
	"flag"
	"log"

	"github.com/spf13/pflag"

	"github.com/statnett/image-scanner-operator/internal/config"
	"github.com/statnett/image-scanner-operator/internal/operator"
)

func main() {
	cfg := config.Config{}
	cfg.Zap.Development = true

	opr := operator.Operator{}
	if err := opr.BindFlags(&cfg, flag.CommandLine); err != nil {
		log.Fatal(err)
	}
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	if err := opr.UnmarshalConfig(&cfg); err != nil {
		log.Fatal(err)
	}

	if err := opr.ValidateConfig(cfg); err != nil {
		log.Fatal(err)
	}

	if err := opr.Start(cfg); err != nil {
		log.Fatal(err)
	}
}
