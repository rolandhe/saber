package main

import "github.com/rolandhe/saber/nfour"

func main() {
	conf := nfour.NewSrvConf(func(task *nfour.Task) ([]byte, error) {
		return []byte("echo:" + string(task.PayLoad)), nil
	}, func(err error) []byte {
		return []byte(err.Error())
	}, 1000)

	nfour.Startup(11011, conf)
}
