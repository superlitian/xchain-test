package main

import (
	"xchain_test/service"
)

func main() {

	// 如果重新部署合约，需要在 service.go 中开启 common.Deploy()

	service.Start()
}
