package main

import (
	"flag"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var flagNodeUrl string
var flagMiner string
var flagAlgo string
var flagHost string
var flagPort uint
var flagServerID uint
var flagDebug bool

func main() {
	initLog()

	var err error

	flag.StringVar(&flagNodeUrl, "nodeurl", "http://127.0.0.1:9735", "RPC URL of node")
	flag.StringVar(&flagMiner, "miner", "", "address of miner account")
	flag.StringVar(&flagAlgo, "algo", "SHA256D", "algo name, such as SHA256D/X11/SCRYPT")

	flag.UintVar(&flagServerID, "serverid", 1, "id of server(1~254)")
	flag.StringVar(&flagHost, "host", "0.0.0.0", "host of server listen")
	flag.UintVar(&flagPort, "port", 3333, "port of server listen(1024~65534)")

	flag.BoolVar(&flagDebug, "debug", false, "enable debug")

	flag.Parse()

	if flagServerID < 1 || flagServerID > 254 {
		log.Errorf("invalid serverid")
		return
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	//log.SetReportCaller(true)
	if flagDebug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	jr := NewJobRepository()
	err = jr.Start()
	if err != nil {
	}

	nc := NewNodeClient(flagNodeUrl, flagMiner, flagAlgo)
	err = nc.Start()
	if err != nil {
	}

	strCfg := &StratumConfig{ServerID: uint8(flagServerID), Host: flagHost, Port: uint32(flagPort), MaxConn: 10}
	ss := NewStratumServer(strCfg)
	err = ss.Start()
	if err != nil {
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	<-c

	nc.Stop()
	jr.Stop()
	ss.Stop()
}

func initLog() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		dir = "/tmp"
	}

	fn := dir + "/gqlc-stratum.log"

	lw, err := rotatelogs.New(
		fn+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(fn),
	)

	lh := lfshook.NewHook(
		lw,
		&log.JSONFormatter{},
	)
	log.AddHook(lh)
}
