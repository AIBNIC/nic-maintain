package core

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/scrapli/scrapligo/driver/network"
	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/platform"
)

func Econ_connect(ip, user, pwd, secret string) *network.Driver {
	fmt.Printf("开始连接 %s ...\n", ip)

	p, err := platform.NewPlatform(
		"cisco_iosxe",
		ip,
		options.WithAuthNoStrictKey(),
		options.WithAuthUsername(user),
		options.WithAuthPassword(pwd),
		options.WithAuthSecondary(secret),
		options.WithDefaultDesiredPriv("privilege-exec"),
		options.WithTransportType("standard"),
		options.WithStandardTransportExtraKexs([]string{"diffie-hellman-group1-sha1"}),
		options.WithStandardTransportExtraCiphers([]string{"3des-cbc"}),
	)

	if err != nil {
		log.Fatalf("failed to create platform; error: %+v", err)
	}

	d, err := p.GetNetworkDriver()
	if err != nil {
		log.Fatalf("failed to fetch network driver from the platform; error: %+v", err)
	}

	err = d.Open()
	if err != nil {
		fmt.Printf("%+v\n由于连接 %s 失败，正在跳转到下一台设备中\n", err, ip)
		return nil
	}

	fmt.Printf("已连接 %s ...\n", ip)

	return d
}

func Econ_inspection(d *network.Driver) []any {
	fmt.Println("开始巡检...")

	r, err := d.SendCommands([]string{
		"show memory",
		"show cpu",
	})
	if err != nil {
		log.Fatalf("unable to run command: %v", err)
	}

	reg_memory := regexp.MustCompile(`(\d.*)%`)
	reg_cpu := regexp.MustCompile(`.*? (\d.*)%`)

	var output_memory []string
	var output_cpu [][]string

	for i, r := range r.Responses {
		if i == 0 {
			output_memory = reg_memory.FindStringSubmatch(r.Result)
			fmt.Println("内存使用率：", output_memory[0])
		} else if i == 1 {
			output_cpu = reg_cpu.FindAllStringSubmatch(r.Result, -1)
			fmt.Printf(
				"CPU使用率： 五秒内使用：%s， 一分钟内使用：%s， 五分钟内使用：%s\n",
				output_cpu[0][1]+"%",
				output_cpu[1][1]+"%",
				output_cpu[2][1]+"%",
			)
		}
	}

	prompt, err := d.Channel.GetPrompt()
	if err != nil {
		log.Fatalf("failed to get prompt; error: %+v", err)
	}

	cpu_5s, _ := strconv.ParseFloat(output_cpu[0][1], 32)
	cpu_1m, _ := strconv.ParseFloat(output_cpu[1][1], 32)
	cpu_5m, _ := strconv.ParseFloat(output_cpu[2][1], 32)
	mem, _ := strconv.ParseFloat(output_memory[1], 32)

	return []any{
		string(prompt),
		cpu_5s, cpu_1m, cpu_5m,
		mem,
	}
}
