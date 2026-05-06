package baostock_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/millken/baostock"
)

// ExampleClient_QueryAllStockIter 演示基本的迭代器用法
func ExampleClient_QueryAllStockIter() {
	client := baostock.NewClient()
	if err := client.Login(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer client.Logout(context.Background())

	// 使用迭代器查询所有股票
	for record := range client.QueryAllStockIter(context.Background(), "2023-12-29") {
		if record == nil {
			break // 错误或迭代结束
		}
		fmt.Printf("代码: %s, 名称: %s\n", record[0], record[1])
	}
}

// ExampleClient_QueryHistoryKDataPlusIter 演示查询历史K线数据的迭代器用法
func ExampleClient_QueryHistoryKDataPlusIter() {
	client := baostock.NewClient()
	if err := client.Login(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer client.Logout(context.Background())

	// 查询浦发银行的历史K线数据
	for fields, record := range client.QueryHistoryKDataPlusIter(context.Background(),
		&baostock.HistoryKDataRequest{
			Code:      "sh.600000",
			Fields:    strings.Join(baostock.DailyKLineCommonFields, ","),
			StartDate: "2023-01-01",
			EndDate:   "2023-12-31",
			Frequency: baostock.FrequencyDaily,
		}) {
		if fields == nil {
			break
		}
		fmt.Printf("日期: %s, 收盘: %s\n", record[0], record[5])
	}
}

// ExampleClient_QueryAllStockIter_earlyBreak 演示提前终止迭代
func ExampleClient_QueryAllStockIter_earlyBreak() {
	client := baostock.NewClient()
	if err := client.Login(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer client.Logout(context.Background())

	count := 0
	// 只获取前10条记录
	for record := range client.QueryAllStockIter(context.Background(), "2023-12-29") {
		if record == nil {
			break
		}
		count++
		fmt.Printf("代码: %s, 名称: %s\n", record[0], record[1])
		if count >= 10 {
			break // 提前终止
		}
	}
}

// ExampleClient_QueryAllStockIter_withFilter 演示在迭代中过滤数据
func ExampleClient_QueryAllStockIter_withFilter() {
	client := baostock.NewClient()
	if err := client.Login(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer client.Logout(context.Background())

	// 只获取上海主板股票（以 sh.600 开头）
	for record := range client.QueryAllStockIter(context.Background(), "2023-12-29") {
		if record == nil {
			break
		}
		code := record[0]
		if strings.HasPrefix(code, "sh.600") || strings.HasPrefix(code, "sh.601") ||
			strings.HasPrefix(code, "sh.603") || strings.HasPrefix(code, "sh.605") {
			fmt.Printf("上海主板: %s - %s\n", code, record[1])
		}
	}
}

// ExampleClient_QueryAllStockIter_collectToSlice 演示收集数据到切片
func ExampleClient_QueryAllStockIter_collectToSlice() {
	client := baostock.NewClient()
	if err := client.Login(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer client.Logout(context.Background())

	// 收集前100条股票代码
	var codes []string
	for record := range client.QueryAllStockIter(context.Background(), "2023-12-29") {
		if record == nil {
			break
		}
		codes = append(codes, record[0])
		if len(codes) >= 100 {
			break
		}
	}
	fmt.Printf("收集了 %d 只股票\n", len(codes))
}

// ExampleClient_QueryHS300StocksIter 演示查询指数成分股
func ExampleClient_QueryHS300StocksIter() {
	client := baostock.NewClient()
	if err := client.Login(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer client.Logout(context.Background())

	// 查询沪深300成分股
	for record := range client.QueryHS300StocksIter(context.Background(), "2023-12-29") {
		if record == nil {
			break
		}
		fmt.Printf("代码: %s, 名称: %s\n", record[0], record[1])
	}
}

// ExampleClient_QueryDepositRateDataIter 演示查询经济数据
func ExampleClient_QueryDepositRateDataIter() {
	client := baostock.NewClient()
	if err := client.Login(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer client.Logout(context.Background())

	// 查询存款利率数据
	for record := range client.QueryDepositRateDataIter(context.Background(), "2020-01-01", "2023-12-31") {
		if record == nil {
			break
		}
		fmt.Printf("日期: %s, 利率: %s\n", record[0], record[1])
	}
}

// ExampleClient_QueryStockIndustryIter 演示查询行业分类
func ExampleClient_QueryStockIndustryIter() {
	client := baostock.NewClient()
	if err := client.Login(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer client.Logout(context.Background())

	// 查询贵州茅台的行业分类
	for record := range client.QueryStockIndustryIter(context.Background(), "sh.600519", "2023-12-29") {
		if record == nil {
			break
		}
		fmt.Printf("代码: %s, 行业: %s\n", record[0], record[1])
	}
}

// ExampleClient_QuerySTStocksIter 演示查询特殊股票
func ExampleClient_QuerySTStocksIter() {
	client := baostock.NewClient()
	if err := client.Login(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer client.Logout(context.Background())

	// 查询ST股票
	for record := range client.QuerySTStocksIter(context.Background(), "2023-12-29") {
		if record == nil {
			break
		}
		fmt.Printf("代码: %s, 名称: %s\n", record[0], record[1])
	}
}

// ExampleClient_comparison 对比回调和迭代器的使用方式
func ExampleClient_comparison() {
	client := baostock.NewClient()
	if err := client.Login(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer client.Logout(context.Background())

	// 旧方式：使用回调
	err := client.QueryAllStock(context.Background(), "2023-12-29",
		func(record []string) error {
			fmt.Printf("[回调] 代码: %s\n", record[0])
			return nil
		})
	if err != nil {
		log.Fatal(err)
	}

	// 新方式：使用迭代器
	for record := range client.QueryAllStockIter(context.Background(), "2023-12-29") {
		if record == nil {
			break
		}
		fmt.Printf("[迭代器] 代码: %s\n", record[0])
	}
}
