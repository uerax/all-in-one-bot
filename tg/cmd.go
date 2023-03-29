package tg

import (
	"fmt"
	"strings"
)

// must Coin Value
func addCryptoGrowthMonitor(id int64, args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 2 {
		fmt.Printf("addCryptoGrowthMonitor 参数有误: %s", args)
		return
	}
	
	api.CryptoApi.Add(id, arg[0], arg[1])
	
}