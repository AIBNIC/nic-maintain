package core

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Info_to_list() [][]string {
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

func Choice_list(switch_list [][]string) [][]string {
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
