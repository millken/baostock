package baostock

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestClientLogin 测试客户端登录
func TestClientLogin(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	if !client.loggedIn {
		t.Error("登录后 loggedIn 状态应该为 true")
	}

	if client.userID == "" {
		t.Error("登录后 userID 不应为空")
	}
}

// TestClientQueryHistoryKDataPlus 测试查询历史K线数据（流式）
func TestClientQueryHistoryKDataPlus(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	recordCount := 0
	var receivedFields []string

	err := client.QueryHistoryKDataPlus(context.Background(),
		&HistoryKDataRequest{
			Code:       "sh.600000",
			Fields:     "date,code,open,high,low,close,volume",
			StartDate:  "2023-12-01",
			EndDate:    "2023-12-31",
			Frequency:  FrequencyDaily,
			AdjustFlag: AdjustFlagNoAdjust,
		},
		func(fields []string, record []string) error {
			if len(receivedFields) == 0 {
				receivedFields = fields
			}
			recordCount++
			return nil
		})

	if err != nil {
		t.Fatalf("查询K线数据失败: %v", err)
	}

	if recordCount == 0 {
		t.Error("应该返回数据")
	}

	if len(receivedFields) != 7 {
		t.Errorf("应该有7个字段, 实际: %d", len(receivedFields))
	}
}

// TestClientQueryHistoryKDataPlusEarlyStop 测试流式查询提前终止
func TestClientQueryHistoryKDataPlusEarlyStop(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	maxRecords := 5
	recordCount := 0

	err := client.QueryHistoryKDataPlus(context.Background(),
		&HistoryKDataRequest{
			Code:       "sh.600000",
			Fields:     "date,code,open,high,low,close,volume",
			StartDate:  "2023-12-01",
			EndDate:    "2023-12-31",
			Frequency:  FrequencyDaily,
			AdjustFlag: AdjustFlagNoAdjust,
		},
		func(fields []string, record []string) error {
			recordCount++
			if recordCount >= maxRecords {
				return errors.New("stop early")
			}
			return nil
		})

	if err == nil {
		t.Error("应该返回提前终止的错误")
	}

	if err != nil && err.Error() != "stop early" {
		t.Errorf("返回的错误应该是 'stop early', 实际: %v", err)
	}

	if recordCount != maxRecords {
		t.Errorf("应该在处理 %d 条记录后停止, 实际处理了: %d", maxRecords, recordCount)
	}
}

// TestClientQueryHistoryKDataPlusCancel 测试流式查询 context 取消
func TestClientQueryHistoryKDataPlusCancel(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	recordCount := 0
	cancelCalled := false

	err := client.QueryHistoryKDataPlus(ctx,
		&HistoryKDataRequest{
			Code:       "sh.600000",
			Fields:     "date,code,open,high,low,close,volume",
			StartDate:  "2020-01-01", // 大范围数据，用于测试取消
			EndDate:    "2023-12-31",
			Frequency:  FrequencyDaily,
			AdjustFlag: AdjustFlagNoAdjust,
		},
		func(fields []string, record []string) error {
			recordCount++
			if recordCount >= 10 && !cancelCalled {
				cancelCalled = true
				cancel()
			}
			return nil
		})

	if cancelCalled && err != context.Canceled && err != nil {
		t.Errorf("调用 cancel 后应该返回 context.Canceled, 实际: %v", err)
	}

	if recordCount < 10 {
		t.Errorf("应该至少处理10条记录, 实际: %d", recordCount)
	}
}

// TestClientQueryTradeDates 测试查询交易日
func TestClientQueryTradeDates(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

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
		t.Fatalf("查询交易日失败: %v", err)
	}

	if totalDates == 0 {
		t.Error("应该返回交易日数据")
	}

	if tradingDays == 0 {
		t.Error("12月应该有交易日")
	}
}

// TestClientQueryAllStock 测试查询所有股票
func TestClientQueryAllStock(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	totalCount := 0

	err := client.QueryAllStock(context.Background(), "2023-12-29",
		func(record []string) error {
			totalCount++
			return nil
		})

	if err != nil {
		t.Fatalf("查询所有股票失败: %v", err)
	}

	if totalCount < 1000 {
		t.Errorf("股票数量太少: %d", totalCount)
	}
}

// TestClientQueryHS300Stocks 测试查询沪深300成分股
func TestClientQueryHS300Stocks(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	totalCount := 0

	err := client.QueryHS300Stocks(context.Background(), "2023-12-29",
		func(record []string) error {
			totalCount++
			return nil
		})

	if err != nil {
		t.Fatalf("查询沪深300失败: %v", err)
	}

	if totalCount != 300 {
		t.Errorf("沪深300应该有300只股票, 实际: %d", totalCount)
	}
}

// TestClientQuerySZ50Stocks 测试查询上证50成分股
func TestClientQuerySZ50Stocks(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	totalCount := 0

	err := client.QuerySZ50Stocks(context.Background(), "2023-12-29",
		func(record []string) error {
			totalCount++
			return nil
		})

	if err != nil {
		t.Fatalf("查询上证50失败: %v", err)
	}

	if totalCount != 50 {
		t.Errorf("上证50应该有50只股票, 实际: %d", totalCount)
	}
}

// TestClientQueryStockBasic 测试查询股票基本信息
func TestClientQueryStockBasic(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	recordCount := 0

	err := client.QueryStockBasic(context.Background(), "sh.600000", "",
		func(record []string) error {
			recordCount++
			return nil
		})

	if err != nil {
		t.Fatalf("查询股票基本信息失败: %v", err)
	}

	if recordCount == 0 {
		t.Error("应该返回股票基本信息")
	}
}

// TestClientQueryProfitData 测试查询季频盈利能力
func TestClientQueryProfitData(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	data, err := client.QueryProfitData(context.Background(),
		&QuarterlyDataRequest{
			Code:    "sh.600000",
			Year:    2023,
			Quarter: 4,
		})
	if err != nil {
		t.Fatalf("查询盈利能力数据失败: %v", err)
	}

	if data.ErrorCode != ErrSuccess {
		t.Fatalf("查询失败: %s", data.ErrorMsg)
	}
}

// TestClientQueryDividendData 测试查询股息分红
func TestClientQueryDividendData(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	data, err := client.QueryDividendData(context.Background(),
		&DividendDataRequest{
			Code:     "sh.600000",
			Year:     2023,
			YearType: "report",
		})
	if err != nil {
		t.Fatalf("查询股息分红失败: %v", err)
	}

	if data.ErrorCode != ErrSuccess {
		t.Fatalf("查询失败: %s", data.ErrorMsg)
	}
}

// TestClientQuerySTStocks 测试查询ST股票
func TestClientQuerySTStocks(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	recordCount := 0

	err := client.QuerySTStocks(context.Background(), "2023-12-29",
		func(record []string) error {
			recordCount++
			return nil
		})

	if err != nil {
		t.Fatalf("查询ST股票失败: %v", err)
	}

	// ST股票数量可能为0，所以只检查查询成功
	// 只要没有错误就算成功
}

// TestClientWithCustomConfig 测试自定义配置
func TestClientWithCustomConfig(t *testing.T) {
	config := &Config{
		Host:     "public-api.baostock.com",
		Port:     10030,
		Username: "anonymous",
		Password: "123456",
		Timeout:  30 * time.Second,
	}

	client := NewClientWithConfig(config)

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	if !client.loggedIn {
		t.Error("登录后 loggedIn 状态应该为 true")
	}
}

// Example functions for documentation

func ExampleClient_basicUsage() {
	client := NewClient()

	client.Login(context.Background())
	defer client.Logout(context.Background())

	// 流式查询历史K线数据
	_ = client.QueryHistoryKDataPlus(context.Background(),
		&HistoryKDataRequest{
			Code:       "sh.600000",
			Fields:     "date,code,open,high,low,close,volume",
			StartDate:  "2023-01-01",
			EndDate:    "2023-12-31",
			Frequency:  FrequencyDaily,
			AdjustFlag: AdjustFlagNoAdjust,
		},
		func(fields []string, record []string) error {
			// 处理每条记录
			return nil
		})
}

func ExampleClient_queryTradeDates() {
	client := NewClient()

	client.Login(context.Background())
	defer client.Logout(context.Background())

	_ = client.QueryTradeDates(context.Background(), "2023-01-01", "2023-01-31",
		func(record []string) error {
			_ = record[0]
			_ = record[1]
			return nil
		})
}

func ExampleClient_queryAllStock() {
	client := NewClient()

	client.Login(context.Background())
	defer client.Logout(context.Background())

	count := 0
	_ = client.QueryAllStock(context.Background(), "2023-12-29",
		func(record []string) error {
			count++
			_, _, _ = record[0], record[1], record[6]
			return nil
		})
	_ = count
}

func TestExampleClientQueryStockBasic(t *testing.T) {
	client := NewClient()

	client.Login(context.Background())
	defer client.Logout(context.Background())
	total := 0
	_ = client.QueryStockBasic(context.Background(), "sh.600000", "",
		func(record []string) error {
			_, _, _ = record[0], record[1], record[7]
			total++
			return nil
		})
	fmt.Printf("查询到 %d 条股票基本信息\n", total)
}

func ExampleClient_queryHS300Stocks() {
	client := NewClient()

	client.Login(context.Background())
	defer client.Logout(context.Background())

	_ = client.QueryHS300Stocks(context.Background(), "2023-12-29",
		func(record []string) error {
			_, _ = record[0], record[1]
			return nil
		})
}

func ExampleClient_queryDepositRate() {
	client := NewClient()

	client.Login(context.Background())
	defer client.Logout(context.Background())

	_ = client.QueryDepositRateData(context.Background(), "2020-01-01", "2023-12-31",
		func(record []string) error {
			_, _ = record[0], record[1]
			return nil
		})
}

func ExampleClient_withCustomConfig() {
	config := &Config{
		Host:     "public-api.baostock.com",
		Port:     10030,
		Username: "anonymous",
		Password: "123456",
		Timeout:  60 * time.Second,
	}

	client := NewClientWithConfig(config)

	client.Login(context.Background())
	defer client.Logout(context.Background())

	_ = client
}

func Example_errorHandling() {
	client := NewClient()

	err := client.Login(context.Background())
	if err != nil {
		var bsErr *Error
		if errors.As(err, &bsErr) {
			_ = bsErr.Code
			_ = bsErr.Message
		}
		return
	}
	defer client.Logout(context.Background())

	_ = client.QueryHistoryKDataPlus(context.Background(),
		&HistoryKDataRequest{
			Code:       "sh.600000",
			Fields:     "date,code,open,high,low,close,volume",
			StartDate:  "2023-01-01",
			EndDate:    "2023-12-31",
			Frequency:  FrequencyDaily,
			AdjustFlag: AdjustFlagNoAdjust,
		},
		func(fields []string, record []string) error {
			return nil
		})
}

func ExampleClient_queryProfitData() {
	client := NewClient()

	client.Login(context.Background())
	defer client.Logout(context.Background())

	data, _ := client.QueryProfitData(context.Background(),
		&QuarterlyDataRequest{
			Code:    "sh.600000",
			Year:    2023,
			Quarter: 4,
		})

	_ = data.ErrorCode
	_ = data.Fields
	for _, row := range data.Data {
		_ = row
	}
}

func ExampleClient_queryDividendData() {
	client := NewClient()

	client.Login(context.Background())
	defer client.Logout(context.Background())

	data, _ := client.QueryDividendData(context.Background(),
		&DividendDataRequest{
			Code:     "sh.600000",
			Year:     2023,
			YearType: "report",
		})

	_ = data.ErrorCode
	for _, row := range data.Data {
		_ = row
	}
}

func ExampleClient_querySTStocks() {
	client := NewClient()

	client.Login(context.Background())
	defer client.Logout(context.Background())

	_ = client.QuerySTStocks(context.Background(), "2023-12-29",
		func(record []string) error {
			_, _ = record[0], record[1]
			return nil
		})
}

// ExampleClient_queryHistoryKDataPlus 流式查询K线数据示例
func ExampleClient_queryHistoryKDataPlus() {
	client := NewClient()

	client.Login(context.Background())
	defer client.Logout(context.Background())

	// 使用预定义字段集合
	_ = client.QueryHistoryKDataPlus(context.Background(),
		&HistoryKDataRequest{
			Code:       "sh.600000",
			Fields:     strings.Join(DailyKLineCommonFields, ","),
			StartDate:  "2023-01-01",
			EndDate:    "2023-12-31",
			Frequency:  FrequencyDaily,
			AdjustFlag: AdjustFlagForward,
		},
		func(fields []string, record []string) error {
			// 处理每条记录，如写入文件或数据库
			_ = fields
			_ = record
			return nil
		})
}

// TestErrorString 测试错误类型的字符串输出
func TestErrorString(t *testing.T) {
	err := &Error{Code: ErrSuccess, Message: "success"}
	str := err.Error()
	if str == "" {
		t.Error("错误字符串不应为空")
	}

	err2 := &Error{Code: "99999", Message: "test error"}
	str2 := err2.Error()
	if str2 == "" {
		t.Error("错误字符串不应为空")
	}
}

// TestConstants 测试常量值
func TestConstants(t *testing.T) {
	if DefaultServerHost == "" {
		t.Error("DefaultServerHost 不应为空")
	}
	if DefaultServerPort == 0 {
		t.Error("DefaultServerPort 不应为0")
	}
	if ClientVersion == "" {
		t.Error("ClientVersion 不应为空")
	}
	if DefaultPerPageCount == 0 {
		t.Error("DefaultPerPageCount 不应为0")
	}
}

// TestFrequency 测试频率常量
func TestFrequency(t *testing.T) {
	tests := []struct {
		name     string
		freq     Frequency
		expected string
	}{
		{"5Min", Frequency5Min, "5"},
		{"15Min", Frequency15Min, "15"},
		{"30Min", Frequency30Min, "30"},
		{"60Min", Frequency60Min, "60"},
		{"Daily", FrequencyDaily, "d"},
		{"Weekly", FrequencyWeek, "w"},
		{"Monthly", FrequencyMonth, "m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.freq) != tt.expected {
				t.Errorf("频率 %s 应该是 %s, 实际是 %s", tt.name, tt.expected, tt.freq)
			}
		})
	}
}

// TestAdjustFlag 测试复权标志常量
func TestAdjustFlag(t *testing.T) {
	tests := []struct {
		name     string
		flag     AdjustFlag
		expected string
	}{
		{"Backward", AdjustFlagBackward, "1"},
		{"Forward", AdjustFlagForward, "2"},
		{"NoAdjust", AdjustFlagNoAdjust, "3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.flag) != tt.expected {
				t.Errorf("复权标志 %s 应该是 %s, 实际是 %s", tt.name, tt.expected, tt.flag)
			}
		})
	}
}

// TestDefaultConfig 测试默认配置
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Host != DefaultServerHost {
		t.Errorf("默认Host应该是 %s, 实际是 %s", DefaultServerHost, config.Host)
	}
	if config.Port != DefaultServerPort {
		t.Errorf("默认Port应该是 %d, 实际是 %d", DefaultServerPort, config.Port)
	}
	if config.Username != "anonymous" {
		t.Errorf("默认Username应该是 anonymous, 实际是 %s", config.Username)
	}
	if config.Password != "123456" {
		t.Errorf("默认Password应该是 123456, 实际是 %s", config.Password)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("默认Timeout应该是 30秒, 实际是 %v", config.Timeout)
	}
}

// TestNewClient 测试创建客户端
func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient() 不应返回 nil")
	}

	if client.conn != nil {
		t.Error("新客户端的连接应该为 nil")
	}

	if client.loggedIn {
		t.Error("新客户端的 loggedIn 应该为 false")
	}
}

// TestNewClientWithConfig 测试使用配置创建客户端
func TestNewClientWithConfig(t *testing.T) {
	config := &Config{
		Host:     "testhost",
		Port:     9999,
		Username: "testuser",
		Password: "testpass",
		Timeout:  10 * time.Second,
	}

	client := NewClientWithConfig(config)

	if client == nil {
		t.Fatal("NewClientWithConfig() 不应返回 nil")
	}

	if client.config != config {
		t.Error("客户端配置应该与传入的配置相同")
	}
}

// ExampleNewClient 创建新客户端示例
func ExampleNewClient() {
	client := NewClient()
	_ = client
}

// ExampleNewClientWithConfig 使用自定义配置创建客户端示例
func ExampleNewClientWithConfig() {
	config := &Config{
		Host:     "public-api.baostock.com",
		Port:     10030,
		Username: "anonymous",
		Password: "123456",
		Timeout:  30 * time.Second,
	}
	client := NewClientWithConfig(config)
	_ = client
}

// TestClientSixDigitStockCode 测试6位股票代码自动规范化
func TestClientSixDigitStockCode(t *testing.T) {
	client := NewClient()

	if err := client.Login(context.Background()); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	defer client.Logout(context.Background())

	t.Run("上海6位代码", func(t *testing.T) {
		var returnedCode string
		err := client.QueryHistoryKDataPlus(context.Background(),
			&HistoryKDataRequest{
				Code:       "600000",
				Fields:     "date,code,open,high,low,close",
				StartDate:  "2023-12-01",
				EndDate:    "2023-12-31",
				Frequency:  FrequencyDaily,
				AdjustFlag: AdjustFlagNoAdjust,
			},
			func(fields []string, record []string) error {
				if len(record) > 1 {
					returnedCode = record[1]
				}
				return nil // 只获取第一条
			})

		if err != nil {
			t.Fatalf("查询K线数据失败: %v", err)
		}

		if returnedCode != "sh.600000" {
			t.Errorf("返回的代码应该是 sh.600000, 实际是: %s", returnedCode)
		}
	})

	t.Run("深圳6位代码", func(t *testing.T) {
		var returnedCode string
		err := client.QueryHistoryKDataPlus(context.Background(),
			&HistoryKDataRequest{
				Code:       "000001",
				Fields:     "date,code,open,high,low,close",
				StartDate:  "2023-12-01",
				EndDate:    "2023-12-31",
				Frequency:  FrequencyDaily,
				AdjustFlag: AdjustFlagNoAdjust,
			},
			func(fields []string, record []string) error {
				if len(record) > 1 {
					returnedCode = record[1]
				}
				return nil
			})

		if err != nil {
			t.Fatalf("查询K线数据失败: %v", err)
		}

		if returnedCode != "sz.000001" {
			t.Errorf("返回的代码应该是 sz.000001, 实际是: %s", returnedCode)
		}
	})

	t.Run("创业板6位代码", func(t *testing.T) {
		var returnedCode string
		err := client.QueryHistoryKDataPlus(context.Background(),
			&HistoryKDataRequest{
				Code:       "300001",
				Fields:     "date,code,open,high,low,close",
				StartDate:  "2023-12-01",
				EndDate:    "2023-12-31",
				Frequency:  FrequencyDaily,
				AdjustFlag: AdjustFlagNoAdjust,
			},
			func(fields []string, record []string) error {
				if len(record) > 1 {
					returnedCode = record[1]
				}
				return nil
			})

		if err != nil {
			t.Fatalf("查询K线数据失败: %v", err)
		}

		if returnedCode != "sz.300001" {
			t.Errorf("返回的代码应该是 sz.300001, 实际是: %s", returnedCode)
		}
	})
}

// TestNormalizeStockCode 测试股票代码规范化
func TestNormalizeStockCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"6位代码(上海)", "600000", "sh.600000"},
		{"6位代码(深圳主板)", "000001", "sz.000001"},
		{"6位代码(深圳中小板)", "002001", "sz.002001"},
		{"6位代码(深圳创业板)", "300001", "sz.300001"},
		{"6位代码(上海科创板)", "688001", "sh.688001"},
		{"sh.格式", "sh.600000", "sh.600000"},
		{"sz.格式", "sz.000001", "sz.000001"},
		{"sh后缀", "600000sh", "sh.600000"},
		{"无点号", "sh600000", "sh.600000"},
		{"sz后缀", "000001sz", "sz.000001"},
		{"混合大小写", "SH.600000", "sh.600000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeStockCode(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeStockCode(%q) = %q, 期望 %q", tt.input, result, tt.expected)
			}
		})
	}
}
