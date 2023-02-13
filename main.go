package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/scrapli/scrapligo/driver/network"
	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/platform"
	"github.com/xuri/excelize/v2"
)

func info_to_list() [][]string {
	_, err := os.Stat("Econnect_box")
	if err != nil && os.IsNotExist(err) {
		fmt.Println("Econnect_box文件夹创建中...")
		os.Mkdir("Econnect_box", os.ModePerm)
	}

	f, err := os.Open("Econnect_box/switch_info.csv")
	if err != nil {
		fmt.Println("switch_info.csv文件缺失，正在创建.....")
		err := os.WriteFile("Econnect_box/switch_info.csv", []byte("IP地址,用户名,密码,enable密码,第一行请勿更改！"), os.ModePerm)
		if err == nil {
			fmt.Println("创建成功！请添加完信息后重新打开")
		}
		os.Exit(1)
	}

	reader := csv.NewReader(f)
	switch_list, err := reader.ReadAll()

	if err != nil || len(switch_list) <= 1 {
		fmt.Println("你貌似还没输入信息，请添加信息后重新打开,例如：")
		fmt.Println("IP地址   用户名 密码 enable密码\n10.1.1.1 test test test\n10.1.1.2 test test test")
		os.Exit(1)
	}

	return switch_list
}

func choice_list(switch_list [][]string) [][]string {
	fmt.Println("序号 | IP地址 | 用户名 | 密码")

	for i, sw := range switch_list {
		if i != 0 {
			fmt.Printf("%d | %s | %s | %s\n", i, sw[0], sw[1], sw[2])
		}
	}

	fmt.Println("请选择需要执行的交换机[格式：1、1-5]：")

	for {
		var Switch_list_choice_num string
		fmt.Scanf("%s", &Switch_list_choice_num)

		choice_num := strings.Split(Switch_list_choice_num, "-")
		start_num, err := strconv.Atoi(choice_num[0])
		if err != nil {
			fmt.Println("输入有误！请重新输入正确的格式：1 或 1-5")
		} else {
			if start_num == 0 {
				fmt.Println("不在范围内！请重新输入")
			} else {
				if len(choice_num) == 1 {
					if start_num >= len(switch_list) {
						fmt.Println("超过指定范围")
					}

					switch_list = switch_list[start_num : start_num+1]

					break
				} else {
					if len(choice_num) == 2 {
						end_num, err := strconv.Atoi(choice_num[1])
						if end_num >= len(switch_list) || err != nil {
							fmt.Println("超过指定范围")
						}

						switch_list = switch_list[start_num : end_num+1]

						break
					} else {
						println("输入有误！请重新输入")
					}
				}
			}
		}
	}

	return switch_list
}

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

	cpu_5s, _ := strconv.Atoi(output_cpu[0][1])
	cpu_1m, _ := strconv.Atoi(output_cpu[1][1])
	cpu_5m, _ := strconv.Atoi(output_cpu[2][1])
	mem, _ := strconv.Atoi(output_memory[1])

	return []any{
		string(prompt),
		cpu_5s, cpu_1m, cpu_5m,
		mem,
	}
}

func turn_xlsx(data_list [][]any, Threshold_cpu, Threshold_memory int) {
	for i := range data_list {
		var total string

		if data_list[i][1] == "1connect_faild" {
			data_list[i][1] = "连接失败"
			data_list[i][2] = "连接失败"
			data_list[i][3] = "连接失败"
			data_list[i][4] = "连接失败"
			data_list[i][5] = "连接失败"
			data_list[i] = append(data_list[i], "连接失败,请检查网络配置！")
		} else {
			cpu_5s := data_list[i][2].(int)
			cpu_1m := data_list[i][3].(int)
			cpu_5m := data_list[i][4].(int)
			mem := data_list[i][5].(int)
			if cpu_5s >= Threshold_cpu {
				total += "cpu超出偏高 "
			} else {
				if cpu_1m >= Threshold_cpu {
					total += "cpu超出偏高 "
				} else {
					if cpu_5m >= Threshold_cpu {
						total += "cpu超出偏高 "
					} else {
						if mem >= Threshold_memory {
							total += "内存偏高 "
						}

						if total != "" {
							data_list[i] = append(data_list[i], total)
						} else {
							data_list[i] = append(data_list[i], "正常")
						}
					}
				}
			}
		}
	}

	data_list = append([][]any{
		{"管理IP", "主机名", "CPU - 5s", "CPU - 1m", "CPU - 5m", "内存使用", "情况摘要"},
	}, data_list...)

	data_list = append(data_list, []any{"固定参考值", "最大值", 100, 100, 100, 100})
	data_list = append(data_list, []any{"固定参考值", "最小值", 0, 0, 0, 0})

	time_folor := time.Now().Format("2006010215")
	folor := "./Econnect_box/" + time_folor + "/"

	_, err := os.Stat(folor)
	if err != nil && os.IsNotExist(err) {
		os.Mkdir(folor, os.ModePerm)
	}

	xlsx_name := time.Now().Format("20060102") + "-巡检表.xlsx"

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	index, _ := f.NewSheet("巡检表")

	for idx, row := range data_list {
		cell, _ := excelize.CoordinatesToCellName(1, idx+1)
		f.SetSheetRow("巡检表", cell, &row)
	}

	styleHeader, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#0070c0"},
		},
		Font: &excelize.Font{
			Color: "#ffffff",
		},
	})
	f.SetCellStyle("巡检表", "A1", "G1", styleHeader)

	f.SetConditionalFormat("巡检表", "C2:F"+fmt.Sprint(len(data_list)),
		[]excelize.ConditionalFormatOptions{
			{
				Type:     "data_bar",
				BarColor: "#92d050",
				Criteria: "=",
				MinType:  "min",
				MaxType:  "max",
			},
		},
	)

	f.SetColWidth("巡检表", "A", "A", 12)
	f.SetColWidth("巡检表", "B", "B", 18)
	f.SetColWidth("巡检表", "G", "G", 35)

	f.SetActiveSheet(index)
	if err := f.SaveAs(folor + xlsx_name); err != nil {
		fmt.Println(err, "文件被占用，请关闭后重试！")
	}
}

func main() {
	Switch_list := info_to_list()
	var run_num int

	Threshold_cpu := 30
	Threshold_memory := 70

	for {
		fmt.Println("请输入要执行的数字  1.巡检 2.备份 3.退出：")
		fmt.Scanf("%d", &run_num)

		if run_num == 1 {
			var Econ_inspection_box [][]any
			Switch_list_choice := choice_list(Switch_list)
			for _, i := range Switch_list_choice {
				device := Econ_connect(i[0], i[1], i[2], i[3])
				if device == nil {
					Econ_inspection_box = append(
						Econ_inspection_box,
						[]any{i[0], "1connect_faild", -1, -1, -1, -1},
					)
				} else {
					Econ_inspection_info := append([]any{i[0]}, Econ_inspection(device)...)
					Econ_inspection_box = append(Econ_inspection_box, Econ_inspection_info)

					defer device.Close()
				}
			}

			turn_xlsx(Econ_inspection_box, Threshold_cpu, Threshold_memory)
		} else if run_num == 2 {
			// TODO
		} else if run_num == 3 {
			break
		} else {
			fmt.Println("输入有误！请重新输入")
		}
	}
}
