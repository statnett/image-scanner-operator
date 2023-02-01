package main

import (
	"flag"
	"log"

	"github.com/spf13/pflag"
	"github.com/statnett/controller-runtime-viper/pkg/zap"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/statnett/image-scanner-operator/internal/config"
	"github.com/statnett/image-scanner-operator/internal/operator"
)

func main() {
	cfg := config.Config{}
	opts := zap.Options{Development: true}

	opr := operator.Operator{}

	opts.BindFlags(flag.CommandLine)

	if err := opr.BindFlags(&cfg, flag.CommandLine); err != nil {
		log.Fatal(err)
	}

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)
	klog.SetLogger(logger)

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
