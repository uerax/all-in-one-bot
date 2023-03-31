package tg

import (
	"fmt"
	"strings"
)

// Crypto Start
func addCryptoGrowthMonitor(id int64, args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 2 {
		fmt.Printf("addCryptoGrowthMonitor 参数有误: %s", args)
		return
	}
	
	api.CryptoApi.AddHighMonitor(id, arg[0], arg[1])

}

func addCryptoDeclineMonitor(id int64, args string) {
	arg := strings.Split(args, " ")
	if len(arg) != 2 {
		fmt.Printf("addCryptoDeclineMonitor 参数有误: %s", args)
		return
	}
	
	api.CryptoApi.AddLowMonitor(id, arg[0], arg[1])
}

func getCryptoPrice(id int64, args string) {
	arg := strings.Split(args, " ")
	if args == "" {
		arg = nil
	}
	ctp := api.CryptoApi.GetPrice(id, arg...)
	sb := strings.Builder{}
	sb.WriteString("查询的结果如下:")
	for k, v := range ctp {
		sb.WriteString("\n")
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(v)
	}
	api.SendMsg(id, sb.String())	
}

func deleteCryptoMinitor(id int64, args string) {
	arg := strings.Split(args, ",")
	if args == "" {
		return
	}
	api.CryptoApi.DeleteMonitor(id, arg...)
}

// ChatGPT
func chatGPT(id int64, args string) {
	if args == "" {
		return
	}

	api.ChatGPTApi.Ask(id, args)
}

// Default
func execute(id int64, args string) {
	if args == "" {
		return
	}

	api.ChatGPTApi.Ask(id, args)
}