package core

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"

	"github.com/scrapli/scrapligo/driver/network"
	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/platform"
)

func customOnOpen(d *network.Driver) error {
	_, err := d.SendCommand("terminal length 0")
	if err != nil {
		return err
	}

	// Ruijie devices
	// See: https://github.com/scrapli/scrapligo/issues/113#issuecomment-1367100153
	_, err = d.SendCommand("terminal width 256")
	if err != nil {
		return err
	}

	return nil
}

func Econ_connect(ip, user, pwd, secret string) *network.Driver {
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
		options.WithNetworkOnOpen(customOnOpen),
	)

	if err != nil {
		log.Printf("failed to create platform; error: %+v", err)
		return nil
	}

	d, err := p.GetNetworkDriver()
	if err != nil {
		log.Printf("failed to fetch network driver from the platform; error: %+v", err)
		return nil
	}

	err = d.Open()
	if err != nil {
		fmt.Printf("\n连接 %s 失败\n%+v\n", ip, err)
		return nil
	}

	fmt.Printf("\n已连接到 %s\n", ip)
	return d
}

func Econ_inspection(d *network.Driver) []any {
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

	prompt, err := d.GetPrompt()
	if err != nil {
		log.Fatalf("failed to get prompt; error: %+v", err)
	}

	reg_prompt := regexp.MustCompile(`(?im)^([\w.\-@/:]{1,63})#$`)
	output_prompt := reg_prompt.FindStringSubmatch(prompt)

	var res []any
	if cpu_5s, err := strconv.ParseFloat(output_cpu[0][1], 32); err != nil {
		res = append(res, math.NaN())
	} else {
		res = append(res, cpu_5s)
	}
	if cpu_1m, err := strconv.ParseFloat(output_cpu[1][1], 32); err != nil {
		res = append(res, math.NaN())
	} else {
		res = append(res, cpu_1m)
	}
	if cpu_5m, err := strconv.ParseFloat(output_cpu[2][1], 32); err != nil {
		res = append(res, math.NaN())
	} else {
		res = append(res, cpu_5m)
	}
	if mem, err := strconv.ParseFloat(output_memory[1], 32); err != nil {
		res = append(res, math.NaN())
	} else {
		res = append(res, mem)
	}

	return append([]any{output_prompt[1]}, res...)
}

func Econ_backup(d *network.Driver, tftp_ip string) {
	prompt, err := d.GetPrompt()
	if err != nil {
		log.Fatalf("failed to get prompt; error: %+v", err)
	}

	reg_prompt := regexp.MustCompile(`(?im)^([\w.\-@/:]{1,63})#$`)
	output_prompt := reg_prompt.FindStringSubmatch(prompt)

	fmt.Printf("另存为 %s.text\n", output_prompt[1])

	r, err := d.SendCommands([]string{
		"write", // 提前保存一遍再备份
		"copy flash:config.text tftp://" + tftp_ip + "/" + output_prompt[1] + ".text",
	})
	if err != nil {
		log.Fatalf("unable to run command: %v", err)
	}

	for _, r := range r.Responses {
		fmt.Println(r.Result)
	}
}
