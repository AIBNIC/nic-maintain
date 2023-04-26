package main

import (
	"fmt"
	"nic-maintain/core"
)

func main() {
	// 读取数据
	Switch_list := core.Info_to_list()

	// 阈值，超出下值则写入摘要
	Threshold_cpu := 30
	Threshold_memory := 70

	var run_num int = 0
	for {
		if run_num == 0 {
			fmt.Println("请输入要执行的数字  1.巡检 2.备份 3.退出：")
			fmt.Scan(&run_num)
		} else if run_num == 1 {
			var Econ_inspection_box [][]any
			Switch_list_choice := core.Choice_list(Switch_list)
			for _, i := range Switch_list_choice {
				device := core.Econ_connect(i[0], i[1], i[2], i[3])
				if device == nil {
					Econ_inspection_box = append(
						Econ_inspection_box,
						[]any{i[0], "1connect_faild", -1, -1, -1, -1},
					)
				} else {
					Econ_inspection_info := append([]any{i[0]}, core.Econ_inspection(device)...)
					Econ_inspection_box = append(Econ_inspection_box, Econ_inspection_info)

					defer device.Close()
				}
			}

			core.Turn_xlsx(Econ_inspection_box, Threshold_cpu, Threshold_memory)
			run_num = 0
		} else if run_num == 2 {
			Switch_list_choice := core.Choice_list(Switch_list)
			tftp_ip, err := core.Tftp_server()
			if err != nil {
				fmt.Println("错误：无法获取本机 IP")
			} else {
				core.Start_tftp_process()

				// 失败次数
				var error_time int = 0

				for _, i := range Switch_list_choice {
					device := core.Econ_connect(i[0], i[1], i[2], i[3])
					if device == nil {
						error_time += 1
					} else {
						core.Econ_backup(device, tftp_ip)
						defer device.Close()
					}
				}

				core.Stop_tftp_process()
				fmt.Printf("\n尝试连接交换机，失败 %d 个\n", error_time)
			}

			run_num = 0
		} else if run_num == 3 {
			break
		} else {
			fmt.Println("输入有误！请重新输入")
			run_num = 0
		}
	}
}
