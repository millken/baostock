// Command baostock-demo 演示 baostock-go 库的基本用法。
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/millken/baostock"
)

func main() {
	// 创建新客户端
	client := baostock.NewClient()
	// 登录服务器
	fmt.Println("正在连接 BaoStock...")
	if err := client.Login(context.Background()); err != nil {
		log.Fatalf("登录失败: %v", err)
	}
	defer func() {
		if err := client.Logout(context.Background()); err != nil {
			log.Printf("登出警告: %v", err)
		}
	}()
	fmt.Println("登录成功！")

	total := 0
	_ = client.QueryStockBasic(context.Background(), "", "",
		func(record []string) error {
			_, _ = record[0], record[1]
			total++
			return nil
		})
	fmt.Printf("查询到 %d 条股票基本信息\n", total)

	// // 示例1: 查询历史K线数据
	// fmt.Println("\n=== 示例1: 历史K线数据 ===")
	// queryKDataExample(client)

	// // 示例2: 查询交易日
	// fmt.Println("\n=== 示例2: 交易日查询 ===")
	// queryTradeDatesExample(client)

	// // 示例3: 查询所有股票
	// fmt.Println("\n=== 示例3: 所有股票查询 ===")
	// queryAllStocksExample(client)

	// // 示例4: 查询沪深300成分股
	// fmt.Println("\n=== 示例4: 沪深300成分股 ===")
	// queryIndexStocksExample(client)
}

// queryKDataExample 演示K线数据查询（流式）
func queryKDataExample(client *baostock.Client) {
	recordCount := 0
	var fields []string

	err := client.QueryHistoryKDataPlus(context.Background(),
		&baostock.HistoryKDataRequest{
			Code:       "sh.600000", // 浦发银行
			Fields:     strings.Join(baostock.DailyKLineCommonFields, ","),
			StartDate:  "2023-12-01",
			EndDate:    "2023-12-31",
			Frequency:  baostock.FrequencyDaily,     // 日线
			AdjustFlag: baostock.AdjustFlagNoAdjust, // 不复权
		},
		func(f []string, record []string) error {
			if len(fields) == 0 {
				fields = f
				fmt.Printf("字段: %v\n", fields)
			}
			recordCount++

			// 只显示前5条
			if recordCount <= 5 {
				fmt.Printf("%s  %s: 开盘=%s, 最高=%s, 最低=%s, 收盘=%s, 成交量=%s\n",
					record[0], record[1], record[2], record[3], record[4], record[5], record[6])
			}
			return nil
		})

	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}

	fmt.Printf("\n总记录数: %d\n", recordCount)
}

// queryTradeDatesExample 演示交易日查询（流式）
func queryTradeDatesExample(client *baostock.Client) {
	totalDates := 0
	tradingDays := 0

	err := client.QueryTradeDates(context.Background(), "2023-12-01", "2023-12-31",
		func(record []string) error {
			totalDates++
			if len(record) > 1 && record[1] == "1" {
				tradingDays++
			}
			return nil
		})

	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}

	fmt.Printf("总日期数: %d\n", totalDates)
	fmt.Printf("2023年12月交易日天数: %d\n", tradingDays)
}

// queryAllStocksExample 演示所有股票查询（流式）
func queryAllStocksExample(client *baostock.Client) {
	totalCount := 0
	count := 0

	err := client.QueryAllStock(context.Background(), "2023-12-29",
		func(record []string) error {
			totalCount++
			if count < 10 && len(record) > 1 {
				fmt.Printf("  %s: %s\n", record[0], record[1])
				count++
			}
			return nil
		})

	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}

	fmt.Printf("股票总数: %d\n", totalCount)
	if totalCount > 10 {
		fmt.Println("\n前10只股票已显示...")
	}
}

// queryIndexStocksExample 演示指数成分股查询（流式）
func queryIndexStocksExample(client *baostock.Client) {
	totalCount := 0
	count := 0

	err := client.QueryHS300Stocks(context.Background(), "2023-12-29",
		func(record []string) error {
			totalCount++
			if count < 10 && len(record) > 1 {
				fmt.Printf("  %s: %s\n", record[0], record[1])
				count++
			}
			return nil
		})

	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}

	fmt.Printf("沪深300总数: %d\n", totalCount)
	if totalCount > 10 {
		fmt.Println("\n前10只股票已显示...")
	}
}
