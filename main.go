package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/scshark/Hato/internal"
	"github.com/scshark/Hato/internal/conf"
	"github.com/scshark/Hato/internal/routers"
	"github.com/scshark/Hato/pkg/debug"
	"github.com/scshark/Hato/pkg/util"
)

var (
	noDefaultFeatures bool
	features          suites
)

type suites []string

func (s *suites) String() string {
	return strings.Join(*s, ",")
}

func (s *suites) Set(value string) error {
	for _, item := range strings.Split(value, ",") {
		*s = append(*s, strings.TrimSpace(item))
	}
	return nil
}

func init() {
	flagParse()

	conf.Initialize(features, noDefaultFeatures)
	internal.Initialize()
}

func flagParse() {
	flag.BoolVar(&noDefaultFeatures, "no-default-features", false, "whether use default features")
	flag.Var(&features, "features", "use special features")
	flag.Parse()
}

func main() {
	gin.SetMode(conf.ServerSetting.RunMode)

	router := routers.NewRouter()
	s := &http.Server{
		Addr:           conf.ServerSetting.HttpIp + ":" + conf.ServerSetting.HttpPort,
		Handler:        router,
		ReadTimeout:    conf.ServerSetting.ReadTimeout,
		WriteTimeout:   conf.ServerSetting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	util.PrintHelloBanner(debug.VersionInfo())
	fmt.Fprintf(color.Output, "PaoPao service listen on %s\n",
		color.GreenString(fmt.Sprintf("http://%s:%s", conf.ServerSetting.HttpIp, conf.ServerSetting.HttpPort)),
	)

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("run app failed: %s", err)
	}
}
