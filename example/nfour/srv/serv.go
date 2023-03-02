package main

import (
	"github.com/rolandhe/saber/nfour"
	"github.com/rolandhe/saber/nfour/duplex"
)

func main() {
	conf := nfour.NewSrvConf(func(task *nfour.Task) ([]byte, error) {
		return []byte("echo:" + string(task.PayLoad)), nil
	}, func(err error) []byte {
		return []byte(err.Error())
	}, 1000)

	duplex.Startup(11011, conf)
}
