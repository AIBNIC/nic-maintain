package core

import (
	"fmt"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

func Turn_xlsx(data_list [][]any, Threshold_cpu, Threshold_memory int) {
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