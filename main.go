package main

import (
	"fmt"
	"nic-maintain/core"
)

func main() {
	Switch_list := core.Info_to_list()
	var run_num int

	Threshold_cpu := 30
	Threshold_memory := 70

	for {
		fmt.Println("请输入要执行的数字  1.巡检 2.备份 3.退出：")
		fmt.Scanf("%d", &run_num)

		if run_num == 1 {
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
		} else if run_num == 2 {
			// TODO
		} else if run_num == 3 {
			break
		} else {
			fmt.Println("输入有误！请重新输入")
		}
	}
}
