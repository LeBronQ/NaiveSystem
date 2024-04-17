package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"
	"math/rand"
)

func main() {
	rand.Seed(time.Now().UnixNano())


    // 生成一个范围在 0.0 到 1.0 之间的随机浮点数
    for i := 0; i < 1000; i++ {
    	randomFloat := rand.Float64()
	s := strconv.FormatFloat(randomFloat*100, 'f', -1, 64)
	cmd := exec.Command("sudo", "tc", "qdisc", "change", "dev", "lo", "root", "netem", "loss", s)
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error:", err)
		}
}}
