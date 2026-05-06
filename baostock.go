// Package baostock 提供 BaoStock 证券数据平台的 Go 语言客户端。
// BaoStock 是一个免费、开源的证券数据平台，提供完整的A股历史数据、实时行情、财务数据等。
//
// 使用示例:
//
//	client := baostock.NewClient()
//	if err := client.Login(context.Background()); err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Logout()
//
//	// 查询日线K线数据（流式处理，内存占用恒定）
//	err := client.QueryHistoryKDataPlus(context.Background(),
//	    &baostock.HistoryKDataRequest{
//	        Code:        "sh.600000",
//	        Fields:      strings.Join(baostock.DailyKLineCommonFields, ","),
//	        StartDate:   "2023-01-01",
//	        EndDate:     "2023-12-31",
//	        Frequency:   baostock.FrequencyDaily,
//	        AdjustFlag:  baostock.AdjustFlagNoAdjust,
//	    },
//	    func(fields []string, record []string) error {
//	        // 处理每条记录，如写入文件/数据库/发送到channel
//	        fmt.Printf("日期: %s, 收盘: %s\n", record[0], record[5])
//	        return nil // 返回 error 可停止迭代
//	    })
//
//	// 或自定义字段
//	err := client.QueryHistoryKDataPlus(context.Background(),
//	    &baostock.HistoryKDataRequest{
//	        Code:        "sh.600000",
//	        Fields:      "date,code,open,high,low,close,volume,amount,pctChg",
//	        StartDate:   "2023-01-01",
//	        EndDate:     "2023-12-31",
//	        Frequency:   baostock.FrequencyDaily,
//	        AdjustFlag:  baostock.AdjustFlagForward, // 前复权
//	    },
//	    func(fields []string, record []string) error {
//	        // 处理每条记录
//	        return nil
//	    })
//
// 支持的K线频率：
//   - FrequencyDaily:   日线（支持1990-12-19至今，18个字段）
//   - FrequencyWeek:    周线（11个字段）
//   - FrequencyMonth:   月线（11个字段）
//   - Frequency5Min:    5分钟线（10个字段，指数无数据）
//   - Frequency15Min:   15分钟线（10个字段，指数无数据）
//   - Frequency30Min:   30分钟线（10个字段，指数无数据）
//   - Frequency60Min:   60分钟线（10个字段，指数无数据）
//
// 支持的复权类型：
//   - AdjustFlagForward:  前复权（推荐用于技术分析）
//   - AdjustFlagBackward: 后复权
//   - AdjustFlagNoAdjust: 不复权
//
// 可用字段常量定义在 fields.go 中：
//   - DailyKLineFields:         全部日线字段（18个）
//   - DailyKLineCommonFields:   日线常用字段（8个）
//   - WeeklyMonthlyKLineFields: 周月线字段（11个）
//   - MinuteKLineFields:        分钟线字段（10个）
//
// API文档: https://www.baostock.com/mainContent?file=stockKData.md
package baostock

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// 协议常量
const (
	// 服务器配置
	DefaultServerHost = "public-api.baostock.com"
	DefaultServerPort = 10030
	ClientVersion     = "00.9.10"

	// 消息分隔符
	MessageSplit = "\x01" // ASCII 0x01
	Delimiter    = "\n"   // ASCII 0x0A

	// 消息结构
	MessageHeaderLength     = 21
	MessageHeaderBodyLength = 10

	// 分页
	DefaultPerPageCount = 10000

	// 证券代码
	StockCodeLength = 9

	// 默认值
	DefaultStartDate = "2015-01-01"
)

// 消息类型代码
const (
	// 登录/登出
	MsgTypeLoginRequest   = "00"
	MsgTypeLoginResponse  = "01"
	MsgTypeLogoutRequest  = "02"
	MsgTypeLogoutResponse = "03"
	MsgTypeError          = "04"

	// K线数据
	MsgTypeGetKDataRequest      = "11"
	MsgTypeGetKDataResponse     = "12"
	MsgTypeGetKDataPlusRequest  = "95"
	MsgTypeGetKDataPlusResponse = "96"

	// 财务数据
	MsgTypeQueryDividendDataRequest  = "13"
	MsgTypeQueryDividendDataResponse = "14"
	MsgTypeAdjustFactorRequest       = "15"
	MsgTypeAdjustFactorResponse      = "16"
	MsgTypeProfitDataRequest         = "17"
	MsgTypeProfitDataResponse        = "18"
	MsgTypeOperationDataRequest      = "19"
	MsgTypeOperationDataResponse     = "20"
	MsgTypeQueryGrowthDataRequest    = "21"
	MsgTypeQueryGrowthDataResponse   = "22"
	MsgTypeQueryDupontDataRequest    = "23"
	MsgTypeQueryDupontDataResponse   = "24"
	MsgTypeQueryBalanceDataRequest   = "25"
	MsgTypeQueryBalanceDataResponse  = "26"
	MsgTypeQueryCashFlowDataRequest  = "27"
	MsgTypeQueryCashFlowDataResponse = "28"

	// 公司公告
	MsgTypeQueryPerformanceExpressReportRequest  = "29"
	MsgTypeQueryPerformanceExpressReportResponse = "30"
	MsgTypeQueryForecastReportRequest            = "31"
	MsgTypeQueryForecastReportResponse           = "32"

	// 元数据查询
	MsgTypeQueryTradeDatesRequest  = "33"
	MsgTypeQueryTradeDatesResponse = "34"
	MsgTypeQueryAllStockRequest    = "35"
	MsgTypeQueryAllStockResponse   = "36"
	MsgTypeQueryStockBasicRequest  = "45"
	MsgTypeQueryStockBasicResponse = "46"

	// 板块信息
	MsgTypeQueryStockIndustryRequest  = "59"
	MsgTypeQueryStockIndustryResponse = "60"
	MsgTypeQueryStockConceptRequest   = "81"
	MsgTypeQueryStockConceptResponse  = "82"
	MsgTypeQueryStockAreaRequest      = "83"
	MsgTypeQueryStockAreaResponse     = "84"

	// 指数成分股
	MsgTypeQueryHS300StocksRequest  = "61"
	MsgTypeQueryHS300StocksResponse = "62"
	MsgTypeQuerySZ50StocksRequest   = "63"
	MsgTypeQuerySZ50StocksResponse  = "64"
	MsgTypeQueryZZ500StocksRequest  = "65"
	MsgTypeQueryZZ500StocksResponse = "66"

	// 特殊股票
	MsgTypeQueryTerminatedStocksRequest  = "67"
	MsgTypeQueryTerminatedStocksResponse = "68"
	MsgTypeQuerySuspendedStocksRequest   = "69"
	MsgTypeQuerySuspendedStocksResponse  = "70"
	MsgTypeQuerySTStocksRequest          = "71"
	MsgTypeQuerySTStocksResponse         = "72"
	MsgTypeQueryStarSTStocksRequest      = "73"
	MsgTypeQueryStarSTStocksResponse     = "74"

	// 市场板块
	MsgTypeQueryAMEStocksRequest  = "85"
	MsgTypeQueryAMEStocksResponse = "86"
	MsgTypeQueryGEMStocksRequest  = "87"
	MsgTypeQueryGEMStocksResponse = "88"

	// 市场互联互通
	MsgTypeQuerySHHKStocksRequest  = "89"
	MsgTypeQuerySHHKStocksResponse = "90"
	MsgTypeQuerySZHKStocksRequest  = "91"
	MsgTypeQuerySZHKStocksResponse = "92"

	// 风险警示
	MsgTypeQueryStockInRiskRequest  = "93"
	MsgTypeQueryStockInRiskResponse = "94"

	// 宏观经济数据
	MsgTypeQueryDepositRateDataRequest           = "47"
	MsgTypeQueryDepositRateDataResponse          = "48"
	MsgTypeQueryLoanRateDataRequest              = "49"
	MsgTypeQueryLoanRateDataResponse             = "50"
	MsgTypeQueryRequiredReserveRatioDataRequest  = "51"
	MsgTypeQueryRequiredReserveRatioDataResponse = "52"
	MsgTypeQueryMoneySupplyDataMonthRequest      = "53"
	MsgTypeQueryMoneySupplyDataMonthResponse     = "54"
	MsgTypeQueryMoneySupplyDataYearRequest       = "55"
	MsgTypeQueryMoneySupplyDataYearResponse      = "56"
	MsgTypeQuerySHIBORDataRequest                = "57"
	MsgTypeQuerySHIBORDataResponse               = "58"
	MsgTypeQueryCPIDataRequest                   = "75"
	MsgTypeQueryCPIDataResponse                  = "76"
	MsgTypeQueryPPIDataRequest                   = "77"
	MsgTypeQueryPPIDataResponse                  = "78"
	MsgTypeQueryPMIDataRequest                   = "79"
	MsgTypeQueryPMIDataResponse                  = "80"

	// 实时行情
	MsgTypeLoginRealTimeRequest    = "37"
	MsgTypeLoginRealTimeResponse   = "38"
	MsgTypeLogoutRealTimeRequest   = "39"
	MsgTypeLogoutRealTimeResponse  = "40"
	MsgTypeSubscriptionsRequest    = "41"
	MsgTypeSubscriptionsResponse   = "42"
	MsgTypeCancelSubscribeRequest  = "43"
	MsgTypeCancelSubscribeResponse = "44"
)

// Frequency 表示K线数据频率
//
// 可获取的数据范围：
//   - 日线: 1990-12-19 至当前时间
//   - 周线: 每周最后一个交易日可获取
//   - 月线: 每月最后一个交易日可获取
//   - 分钟线: 支持5/15/30/60分钟，指数无分钟线数据
//
// 不同频率支持的字段：
//   - 日线: 全部18个字段（包含 peTTM, psTTM, pbMRQ 等估值指标）
//   - 周月线: 11个字段（不包含 preclose, tradestatus, 估值指标）
//   - 分钟线: 10个字段（增加 time 字段，不包含估值指标）
type Frequency string

const (
	Frequency5Min  Frequency = "5"  // 5分钟K线（指数无数据）
	Frequency15Min Frequency = "15" // 15分钟K线（指数无数据）
	Frequency30Min Frequency = "30" // 30分钟K线（指数无数据）
	Frequency60Min Frequency = "60" // 60分钟K线（指数无数据）
	FrequencyDaily Frequency = "d"  // 日K线，支持1990-12-19至今
	FrequencyWeek  Frequency = "w"  // 周K线，每周最后一个交易日可获取
	FrequencyMonth Frequency = "m"  // 月K线，每月最后一个交易日可获取
)

// AdjustFlag 表示复权类型
//
// BaoStock 使用"涨跌幅复权法"进行复权：
//   - 后复权(1): 以当前为基准，历史价格向前调整
//   - 前复权(2): 以当前为基准，未来价格向后调整（推荐用于技术分析）
//   - 不复权(3): 原始价格，不进行复权处理
//
// 注意：不同系统间采用复权方式可能不一致，导致数据与同花顺、通达信等存在差异
type AdjustFlag string

const (
	AdjustFlagBackward AdjustFlag = "1" // 后复权：以当前为基准，历史价格向前调整
	AdjustFlagForward  AdjustFlag = "2" // 前复权：以当前为基准，未来价格向后调整（推荐）
	AdjustFlagNoAdjust AdjustFlag = "3" // 不复权：使用原始价格
)

// Config 表示客户端配置
type Config struct {
	Host     string        // 服务器地址
	Port     int           // 服务器端口
	Username string        // 用户名
	Password string        // 密码
	Timeout  time.Duration // 超时时间
	APIKey   string        // API Key（可选，用于高级功能）
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:     DefaultServerHost,
		Port:     DefaultServerPort,
		Username: "anonymous",
		Password: "123456",
		Timeout:  30 * time.Second,
	}
}

// Client 表示 BaoStock 客户端
type Client struct {
	config    *Config
	conn      net.Conn
	reader    *bufio.Reader
	connected bool
	loggedIn  bool
	userID    string
	apiKey    string
}

// NewClient 使用默认配置创建新的 BaoStock 客户端
func NewClient() *Client {
	return NewClientWithConfig(DefaultConfig())
}

// NewClientWithConfig 使用自定义配置创建新的 BaoStock 客户端
func NewClientWithConfig(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}
	return &Client{
		config: config,
	}
}

// Connect 建立与服务器的连接
func (c *Client) Connect(ctx context.Context) error {
	if c.conn != nil {
		return nil // 已连接
	}

	dialer := &net.Dialer{
		Timeout: c.config.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", c.config.Host, c.config.Port))
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	c.conn = conn
	c.reader = bufio.NewReaderSize(conn, 8192)
	c.connected = true
	return nil
}

// Disconnect 关闭与服务器的连接
func (c *Client) Disconnect() error {
	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	c.reader = nil
	c.connected = false
	c.loggedIn = false
	return err
}

// SetAPIKey 设置 API Key，用于高级功能
func (c *Client) SetAPIKey(key string) {
	c.apiKey = key
}

// Login 与 BaoStock 服务器进行身份验证
func (c *Client) Login(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	options := "0"
	if c.apiKey != "" {
		options = c.apiKey
	}

	msgBody := fmt.Sprintf("login%s%s%s%s%s%s", MessageSplit, c.config.Username, MessageSplit, c.config.Password, MessageSplit, options)

	resp, err := c.sendMessage(ctx, MsgTypeLoginRequest, msgBody)
	if err != nil {
		return err
	}

	result := &LoginResponse{}
	if err := parseLoginResponse(resp, result); err != nil {
		return err
	}

	if result.ErrorCode != ErrSuccess {
		return &Error{Code: result.ErrorCode, Message: result.ErrorMsg}
	}

	c.userID = result.UserID
	c.loggedIn = true
	return nil
}

// Logout 终止会话
func (c *Client) Logout(ctx context.Context) error {
	if !c.loggedIn {
		return nil
	}

	timestamp := time.Now().Format("20060102150405")
	msgBody := fmt.Sprintf("logout%s%s%s%s", MessageSplit, c.userID, MessageSplit, timestamp)

	_, err := c.sendMessage(ctx, MsgTypeLogoutRequest, msgBody)
	if err != nil {
		return err
	}

	c.loggedIn = false
	return c.Disconnect()
}

// sendMessage 向服务器发送消息并返回响应
func (c *Client) sendMessage(ctx context.Context, msgType, msgBody string) (*Response, error) {
	if !c.connected {
		return nil, errors.New("未连接到服务器")
	}

	// 构建消息头
	header := fmt.Sprintf("%s%s%s%s%s", ClientVersion, MessageSplit, msgType, MessageSplit, formatLength(len(msgBody)))

	// 计算CRC32
	fullMsg := header + msgBody
	crc32 := calculateCRC32(fullMsg)

	// 构建完整请求
	request := fmt.Sprintf("%s%s%d%s", fullMsg, MessageSplit, crc32, Delimiter)

	// 发送请求
	if _, err := c.conn.Write([]byte(request)); err != nil {
		return nil, fmt.Errorf("发送消息失败: %w", err)
	}

	// 接收响应（支持 context 取消）
	response, err := c.receiveResponse(ctx)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// receiveResponse 接收并解析来自服务器的响应
func (c *Client) receiveResponse(ctx context.Context) (*Response, error) {
	var buffer bytes.Buffer

	for {
		// 检查 context 是否已取消
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		buffer.WriteString(line)

		// 检查消息是否结束
		str := buffer.String()
		// 对于压缩响应，必须等待 <![CDATA[]]>\n 标记
		// 对于非压缩响应，只需要 \n
		// 先检查是否是压缩响应类型（通过检查消息头）
		if len(str) >= MessageHeaderLength {
			header := str[:MessageHeaderLength]
			headerParts := strings.Split(header, MessageSplit)
			if len(headerParts) >= 2 {
				msgType := headerParts[1]
				// 检查是否是压缩消息类型
				isCompressed := (msgType == MsgTypeGetKDataPlusResponse)

				if isCompressed && strings.HasSuffix(str, "<![CDATA[]]>\n") {
					break
				} else if !isCompressed && strings.HasSuffix(str, "\n") {
					break
				}
			}
		}
	}

	responseStr := buffer.String()

	// 解析消息头
	if len(responseStr) < MessageHeaderLength {
		return nil, errors.New("无效的响应: 消息过短")
	}

	header := responseStr[:MessageHeaderLength]
	body := responseStr[MessageHeaderLength:]

	// 检查是否压缩
	headerParts := strings.Split(header, MessageSplit)
	if len(headerParts) < 3 {
		return nil, errors.New("无效的响应头")
	}

	msgType := headerParts[1]

	// 处理压缩响应
	if msgType == MsgTypeGetKDataPlusResponse {
		// 对于压缩响应，数据格式为: header + [压缩数据]\x1[CRC32]<![CDATA[]]>\n
		// 需要从完整响应中找到 <![CDATA[ 的位置
		before, _, ok := strings.Cut(responseStr, "<![CDATA[")
		if !ok {
			// 未找到 <![CDATA[ 标记，尝试找到数据结束位置
			// 响应格式: header + compressed_data + \x01 + CRC32 + \n
			// 从后往前找第一个 \x01 作为CRC32分隔符
			lastNewline := strings.LastIndex(responseStr, Delimiter)

			if lastNewline > MessageHeaderLength {
				// 在换行符前找最后一个 \x01
				lastSplit := strings.LastIndex(responseStr[:lastNewline], MessageSplit)
				if lastSplit > MessageHeaderLength {
					compressedData := []byte(responseStr[MessageHeaderLength:lastSplit])

					decompressed, err := decompressData(compressedData)
					if err != nil {
						return nil, fmt.Errorf("解压失败: %w", err)
					}

					body = string(decompressed)
				} else {
					return nil, errors.New("压缩响应格式错误: 无法确定数据范围")
				}
			} else {
				return nil, errors.New("压缩响应格式错误: 响应过短")
			}
		} else {
			// 找到 <![CDATA[ 前面的 \x01 分隔符（CRC32前的分隔符）
			cdataEnd := strings.LastIndex(before, MessageSplit)
			if cdataEnd == -1 {
				return nil, errors.New("压缩响应格式错误: 未找到CRC32分隔符")
			}

			compressedData := []byte(responseStr[MessageHeaderLength:cdataEnd])

			decompressed, err := decompressData(compressedData)
			if err != nil {
				return nil, fmt.Errorf("解压失败: %w", err)
			}

			body = string(decompressed)
		}
	}

	return &Response{
		Header:     header,
		Body:       body,
		FullString: responseStr,
	}, nil
}

// Response 表示服务器响应
type Response struct {
	Header     string // 消息头
	Body       string // 消息体
	FullString string // 完整响应字符串
}

// LoginResponse 表示登录响应
type LoginResponse struct {
	ErrorCode string
	ErrorMsg  string
	Method    string
	UserID    string
}

// HistoryKDataRequest 表示历史K线数据请求
//
// 可用字段列表：
//
// 日线字段（包含停牌证券，18个）：
//   date       - 交易所行情日期，格式：YYYY-MM-DD
//   code       - 证券代码，格式：sh.600000 或 sz.000001
//   open       - 开盘价，精度：小数点后4位，单位：人民币元
//   high       - 最高价，精度：小数点后4位，单位：人民币元
//   low        - 最低价，精度：小数点后4位，单位：人民币元
//   close      - 收盘价，精度：小数点后4位，单位：人民币元
//   preclose   - 昨日收盘价，精度：小数点后4位，单位：人民币元
//   volume     - 成交数量，单位：股
//   amount     - 成交金额，精度：小数点后4位，单位：人民币元
//   adjustflag - 复权状态：1=后复权，2=前复权，3=不复权
//   turn       - 换手率，精度：小数点后6位，单位：%
//   tradestatus- 交易状态：1=正常交易，0=停牌
//   pctChg     - 涨跌幅（百分比），精度：小数点后6位
//   peTTM      - 滚动市盈率，精度：小数点后6位
//   psTTM      - 滚动市销率，精度：小数点后6位
//   pcfNcfTTM  - 滚动市现率，精度：小数点后6位
//   pbMRQ      - 市净率，精度：小数点后6位
//   isST       - 是否ST股：1=是，0=否
//
// 周月线字段（11个）：
//   date, code, open, high, low, close, volume, amount,
//   adjustflag, turn, pctChg
//
// 分钟线字段（5/15/30/60分钟，10个）：
//   date       - 交易所行情日期，格式：YYYY-MM-DD
//   time       - 交易所行情时间，格式：YYYYMMDDHHMMSSsss
//   code       - 证券代码
//   open, high, low, close, volume, amount, adjustflag
//
// 注意：
//   - 指数没有分钟线数据
//   - 周线每周最后一个交易日才可以获取
//   - 月线每月最后一个交易日才可以获取
//   - 分钟线数据不包含指数
//
// 示例：
//   // 日线查询（使用预定义字段集合）
//   req.Fields = strings.Join(baostock.DailyKLineFields, ",")
//   或自定义字段:
//   req.Fields = "date,code,open,high,low,close,volume,amount"
type HistoryKDataRequest struct {
	Code       string     // 证券代码，如 "sh.600000" 或 "sz.000001"
	Fields     string     // 字段列表，用逗号分隔，如 "date,code,open,high,low,close"
	StartDate  string     // 开始日期（包含），格式 "YYYY-MM-DD"，为空时取 2015-01-01
	EndDate    string     // 结束日期（包含），格式 "YYYY-MM-DD"，为空时取最近交易日
	Frequency  Frequency  // K线频率：d=日线，w=周线，m=月线，5/15/30/60=分钟线
	AdjustFlag AdjustFlag // 复权标志：1=后复权，2=前复权，3=不复权（默认）
}

// HistoryKDataResponse 表示历史K线数据响应
type HistoryKDataResponse struct {
	ErrorCode    string     // 错误代码
	ErrorMsg     string     // 错误信息
	Method       string     // 方法名
	UserID       string     // 用户ID
	CurPageNum   string     // 当前页码
	PerPageCount string     // 每页条数
	Data         [][]string // 数据
	Code         string     // 证券代码
	Fields       []string   // 字段列表
	StartDate    string     // 开始日期
	EndDate      string     // 结束日期
	Frequency    string     // K线频率
	AdjustFlag   string     // 复权标志
}

// QueryHistoryKDataPlus 查询历史K线数据（流式）
//
// 通过API接口获取A股历史交易数据，支持日K线、周K线、月K线以及5/15/30/60分钟K线数据。
// 可获取1990-12-19至当前时间的数据，支持前复权、后复权、不复权三种类型。
//
// 此方法会自动分页获取数据，对每条记录调用回调函数。优势：
//   - 边下载边处理，内存占用恒定（只保留一页数据）
//   - 支持通过回调函数提前终止（返回 error）
//   - 支持 context 取消，可随时中断下载
//
// Fields 参数说明（不同频率支持的字段不同）：
//
// 日线字段（18个）:
//   date,code,open,high,low,close,preclose,volume,amount,adjustflag,
//   turn,tradestatus,pctChg,peTTM,psTTM,pcfNcfTTM,pbMRQ,isST
//
// 周月线字段（11个）:
//   date,code,open,high,low,close,volume,amount,adjustflag,turn,pctChg
//
// 分钟线字段（10个）:
//   date,time,code,open,high,low,close,volume,amount,adjustflag
//
// 使用预定义字段集合示例:
//   req.Fields = strings.Join(baostock.DailyKLineCommonFields, ",")
//
// 或自定义字段（逗号分隔）:
//   req.Fields = "date,code,open,high,low,close,volume,amount"
//
// callback 参数：
//   - fields: 字段名列表
//   - record: 单条记录数据（与 fields 一一对应）
//   - 返回 error 可停止迭代（返回的 error 会由本函数返回）
//
// 示例：
//   err := client.QueryHistoryKDataPlus(context.Background(),
//       &baostock.HistoryKDataRequest{
//           Code:      "sh.600000",
//           Fields:    strings.Join(baostock.DailyKLineCommonFields, ","),
//           StartDate: "2020-01-01",
//           EndDate:   "2023-12-31",
//           Frequency: baostock.FrequencyDaily,
//       },
//       func(fields []string, record []string) error {
//           // 处理每条记录，如写入文件/数据库/发送到channel
//           fmt.Printf("日期: %s, 收盘: %s\n", record[0], record[5])
//           return nil // 返回 error 可停止迭代
//       })
func (c *Client) QueryHistoryKDataPlus(ctx context.Context, req *HistoryKDataRequest, callback func(fields []string, record []string) error) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	if err := validateStockCode(req.Code); err != nil {
		return err
	}

	startDate := req.StartDate
	if startDate == "" {
		startDate = DefaultStartDate
	}

	endDate := req.EndDate
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	code := normalizeStockCode(req.Code)
	pageNum := 1
	var allFields []string

	for {
		// 检查 context 是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 构建请求消息（带上页码）
		msgBody := fmt.Sprintf("query_history_k_data_plus%s%s%s%d%s%d%s%s%s%s%s%s%s%s%s%s%s%s",
			MessageSplit, c.userID, MessageSplit, pageNum, MessageSplit,
			DefaultPerPageCount, MessageSplit, code, MessageSplit,
			req.Fields, MessageSplit, startDate, MessageSplit,
			endDate, MessageSplit, req.Frequency, MessageSplit, req.AdjustFlag)

		resp, err := c.sendMessage(ctx, MsgTypeGetKDataPlusRequest, msgBody)
		if err != nil {
			return err
		}

		result, err := parseHistoryKDataResponse(resp)
		if err != nil {
			return err
		}

		if result.ErrorCode != ErrSuccess {
			return &Error{Code: result.ErrorCode, Message: result.ErrorMsg}
		}

		// 第一页保存字段列表
		if pageNum == 1 {
			allFields = result.Fields
		}

		// 处理当前页的每条记录
		for _, record := range result.Data {
			if err := callback(allFields, record); err != nil {
				return err
			}
		}

		// 检查是否还有更多页
		// 如果返回的数据量小于每页条数，说明是最后一页
		if len(result.Data) < DefaultPerPageCount {
			break
		}

		pageNum++
	}

	return nil
}

// QueryTradeDates 查询交易日（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QueryTradeDates(ctx context.Context, startDate, endDate string, callback func(record []string) error) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	if startDate == "" {
		startDate = DefaultStartDate
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	msgBody := fmt.Sprintf("query_trade_dates%s%s%s1%s%d%s%s%s%s",
		MessageSplit, c.userID, MessageSplit, MessageSplit,
		DefaultPerPageCount, MessageSplit, startDate, MessageSplit, endDate)

	resp, err := c.sendMessage(ctx, MsgTypeQueryTradeDatesRequest, msgBody)
	if err != nil {
		return err
	}

	bodyParts := strings.Split(resp.Body, MessageSplit)
	if len(bodyParts) < 7 {
		return errors.New("无效的响应")
	}

	errorCode := bodyParts[0]
	if errorCode != ErrSuccess {
		errorMsg := ""
		if len(bodyParts) > 1 {
			errorMsg = bodyParts[1]
		}
		return &Error{Code: errorCode, Message: errorMsg}
	}

	// 流式解析 JSON 数据（避免一次性加载 10000 条记录到内存）
	if err := streamJsonRecords(bodyParts[6], callback); err != nil {
		return err
	}

	return nil
}

// QueryAllStock 查询指定日期的所有股票（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QueryAllStock(ctx context.Context, date string, callback func(record []string) error) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	msgBody := fmt.Sprintf("query_all_stock%s%s%s1%s%d%s%s",
		MessageSplit, c.userID, MessageSplit, MessageSplit,
		DefaultPerPageCount, MessageSplit, date)

	resp, err := c.sendMessage(ctx, MsgTypeQueryAllStockRequest, msgBody)
	if err != nil {
		return err
	}

	bodyParts := strings.Split(resp.Body, MessageSplit)
	if len(bodyParts) < 7 {
		return errors.New("无效的响应")
	}

	errorCode := bodyParts[0]
	if errorCode != ErrSuccess {
		errorMsg := ""
		if len(bodyParts) > 1 {
			errorMsg = bodyParts[1]
		}
		return &Error{Code: errorCode, Message: errorMsg}
	}

	// 流式解析 JSON 数据
	if err := streamJsonRecords(bodyParts[6], callback); err != nil {
		return err
	}

	return nil
}

// QueryStockBasic 查询股票基本信息（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QueryStockBasic(ctx context.Context, code, codeName string, callback func(record []string) error) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	if code != "" {
		if err := validateStockCode(code); err != nil {
			return err
		}
	}

	// 规范化股票代码
	normalizedCode := normalizeStockCode(code)

	msgBody := fmt.Sprintf("query_stock_basic%s%s%s1%s%d%s%s%s%s",
		MessageSplit, c.userID, MessageSplit, MessageSplit,
		DefaultPerPageCount, MessageSplit, normalizedCode, MessageSplit, codeName)

	resp, err := c.sendMessage(ctx, MsgTypeQueryStockBasicRequest, msgBody)
	if err != nil {
		return err
	}

	bodyParts := strings.Split(resp.Body, MessageSplit)
	if len(bodyParts) < 7 {
		return errors.New("无效的响应")
	}

	errorCode := bodyParts[0]
	if errorCode != ErrSuccess {
		errorMsg := ""
		if len(bodyParts) > 1 {
			errorMsg = bodyParts[1]
		}
		return &Error{Code: errorCode, Message: errorMsg}
	}

	// 流式解析 JSON 数据
	if err := streamJsonRecords(bodyParts[6], callback); err != nil {
		return err
	}

	return nil
}

// QueryStockIndustry 查询行业分类（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QueryStockIndustry(ctx context.Context, code, date string, callback func(record []string) error) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	if code != "" {
		if err := validateStockCode(code); err != nil {
			return err
		}
	}

	// 规范化股票代码
	normalizedCode := normalizeStockCode(code)

	msgBody := fmt.Sprintf("query_stock_industry%s%s%s1%s%d%s%s%s%s",
		MessageSplit, c.userID, MessageSplit, MessageSplit,
		DefaultPerPageCount, MessageSplit, normalizedCode, MessageSplit, date)

	resp, err := c.sendMessage(ctx, MsgTypeQueryStockIndustryRequest, msgBody)
	if err != nil {
		return err
	}

	bodyParts := strings.Split(resp.Body, MessageSplit)
	if len(bodyParts) < 7 {
		return errors.New("无效的响应")
	}

	errorCode := bodyParts[0]
	if errorCode != ErrSuccess {
		errorMsg := ""
		if len(bodyParts) > 1 {
			errorMsg = bodyParts[1]
		}
		return &Error{Code: errorCode, Message: errorMsg}
	}

	// 流式解析 JSON 数据
	if err := streamJsonRecords(bodyParts[6], callback); err != nil {
		return err
	}

	return nil
}

// QueryHS300Stocks 查询沪深300成分股（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QueryHS300Stocks(ctx context.Context, date string, callback func(record []string) error) error {
	return c.queryIndexStocks(ctx, MsgTypeQueryHS300StocksRequest, "query_hs300_stocks", date, callback)
}

// QuerySZ50Stocks 查询上证50成分股（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QuerySZ50Stocks(ctx context.Context, date string, callback func(record []string) error) error {
	return c.queryIndexStocks(ctx, MsgTypeQuerySZ50StocksRequest, "query_sz50_stocks", date, callback)
}

// QueryZZ500Stocks 查询中证500成分股（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QueryZZ500Stocks(ctx context.Context, date string, callback func(record []string) error) error {
	return c.queryIndexStocks(ctx, MsgTypeQueryZZ500StocksRequest, "query_zz500_stocks", date, callback)
}

// queryIndexStocks 指数成分股查询辅助方法
func (c *Client) queryIndexStocks(ctx context.Context, msgType, methodName, date string, callback func(record []string) error) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	msgBody := fmt.Sprintf("%s%s%s%s1%s%d%s%s",
		methodName, MessageSplit, c.userID, MessageSplit, MessageSplit,
		DefaultPerPageCount, MessageSplit, date)

	resp, err := c.sendMessage(ctx, msgType, msgBody)
	if err != nil {
		return err
	}

	bodyParts := strings.Split(resp.Body, MessageSplit)
	if len(bodyParts) < 7 {
		return errors.New("无效的响应")
	}

	errorCode := bodyParts[0]
	if errorCode != ErrSuccess {
		errorMsg := ""
		if len(bodyParts) > 1 {
			errorMsg = bodyParts[1]
		}
		return &Error{Code: errorCode, Message: errorMsg}
	}

	// 流式解析 JSON 数据
	if err := streamJsonRecords(bodyParts[6], callback); err != nil {
		return err
	}

	return nil
}

// QueryDepositRateData 查询存款利率数据（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QueryDepositRateData(ctx context.Context, startDate, endDate string, callback func(record []string) error) error {
	return c.queryEconomicData(ctx, MsgTypeQueryDepositRateDataRequest, "query_deposit_rate_data", startDate, endDate, "", callback)
}

// QueryLoanRateData 查询贷款利率数据（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QueryLoanRateData(ctx context.Context, startDate, endDate string, callback func(record []string) error) error {
	return c.queryEconomicData(ctx, MsgTypeQueryLoanRateDataRequest, "query_loan_rate_data", startDate, endDate, "", callback)
}

// QueryCPIData 查询CPI数据（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QueryCPIData(ctx context.Context, startDate, endDate string, callback func(record []string) error) error {
	return c.queryEconomicData(ctx, MsgTypeQueryCPIDataRequest, "query_cpi_data", startDate, endDate, "", callback)
}

// QueryPPIData 查询PPI数据（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QueryPPIData(ctx context.Context, startDate, endDate string, callback func(record []string) error) error {
	return c.queryEconomicData(ctx, MsgTypeQueryPPIDataRequest, "query_ppi_data", startDate, endDate, "", callback)
}

// QueryPMIData 查询PMI数据（流式）
//
// callback 参数：单条记录数据
// 返回 error 可停止迭代
func (c *Client) QueryPMIData(ctx context.Context, startDate, endDate string, callback func(record []string) error) error {
	return c.queryEconomicData(ctx, MsgTypeQueryPMIDataRequest, "query_pmi_data", startDate, endDate, "", callback)
}

// queryEconomicData 经济数据查询辅助方法
func (c *Client) queryEconomicData(ctx context.Context, msgType, methodName, startDate, endDate, extraParam string, callback func(record []string) error) error {
	if err := c.ensureLogin(); err != nil {
		return err
	}

	msgBody := fmt.Sprintf("%s%s%s%s1%s%d%s%s%s%s",
		methodName, MessageSplit, c.userID, MessageSplit, MessageSplit,
		DefaultPerPageCount, MessageSplit, startDate, MessageSplit, endDate)

	if extraParam != "" {
		msgBody += MessageSplit + extraParam
	}

	resp, err := c.sendMessage(ctx, msgType, msgBody)
	if err != nil {
		return err
	}

	result := &struct {
		ErrorCode string `json:"-"`
		ErrorMsg  string `json:"-"`
		jsonData   string `json:"-"`
	}{}

	if err := parseStandardResponse(resp, result); err != nil {
		return err
	}

	if result.ErrorCode != ErrSuccess {
		return &Error{Code: result.ErrorCode, Message: result.ErrorMsg}
	}

	// 流式解析 JSON 数据
	if err := streamJsonRecords(result.jsonData, callback); err != nil {
		return err
	}

	return nil
}

// 辅助函数

func formatLength(length int) string {
	return fmt.Sprintf("%010d", length)
}

func validateStockCode(code string) error {
	// 先规范化股票代码（支持6位代码自动添加市场前缀）
	normalizedCode := normalizeStockCode(code)

	// 验证规范化后的代码
	normalizedCode = strings.ToLower(normalizedCode)
	if len(normalizedCode) != StockCodeLength {
		return fmt.Errorf("证券代码必须为%d位，当前: %s", StockCodeLength, code)
	}

	// 检查格式
	if !strings.HasPrefix(normalizedCode, "sh.") && !strings.HasPrefix(normalizedCode, "sz.") {
		return fmt.Errorf("证券代码必须以'sh.'或'sz.'开头，当前: %s", code)
	}

	return nil
}

func parseLoginResponse(resp *Response, result *LoginResponse) error {
	bodyParts := strings.Split(resp.Body, MessageSplit)
	if len(bodyParts) < 4 {
		return errors.New("无效的登录响应")
	}

	result.ErrorCode = bodyParts[0]
	result.ErrorMsg = bodyParts[1]
	result.Method = bodyParts[2]
	result.UserID = bodyParts[3]

	return nil
}

func parseHistoryKDataResponse(resp *Response) (*HistoryKDataResponse, error) {
	bodyParts := strings.Split(resp.Body, MessageSplit)
	if len(bodyParts) < 13 {
		return nil, errors.New("无效的历史K线数据响应")
	}

	result := &HistoryKDataResponse{
		ErrorCode:    bodyParts[0],
		ErrorMsg:     bodyParts[1],
		Method:       bodyParts[2],
		UserID:       bodyParts[3],
		CurPageNum:   bodyParts[4],
		PerPageCount: bodyParts[5],
		Code:         bodyParts[7],
		StartDate:    bodyParts[9],
		EndDate:      bodyParts[10],
		Frequency:    bodyParts[11],
		AdjustFlag:   bodyParts[12],
	}

	// 解析字段
	if len(bodyParts) > 8 {
		fieldsStr := bodyParts[8]
		result.Fields = strings.Split(fieldsStr, ",")
	}

	// 解析JSON数据
	if len(bodyParts) > 6 {
		dataJSON := bodyParts[6]
		if dataJSON != "" {
			var parsedData struct {
				Record [][]string `json:"record"`
			}
			if err := json.Unmarshal([]byte(dataJSON), &parsedData); err != nil {
				return nil, fmt.Errorf("解析数据JSON失败: %w", err)
			}
			result.Data = parsedData.Record
		}
	}

	return result, nil
}

func parseStandardResponse(resp *Response, result interface{}) error {
	bodyParts := strings.Split(resp.Body, MessageSplit)
	if len(bodyParts) < 6 {
		return errors.New("无效的响应")
	}

	// 使用反射或类型断言设置错误字段
	// 为简单起见，假设result有ErrorCode和ErrorMsg字段
	if r, ok := result.(interface{ setErrorCode(string) }); ok {
		r.setErrorCode(bodyParts[0])
	}
	if r, ok := result.(interface{ setErrorMsg(string) }); ok {
		r.setErrorMsg(bodyParts[1])
	}

	// 从bodyParts[6]解析JSON数据
	if len(bodyParts) > 6 && bodyParts[6] != "" {
		// 检查是否有 jsonData 字段用于流式解析
		if r, ok := result.(interface{ setJsonData(string) }); ok {
			r.setJsonData(bodyParts[6])
		}
		if err := json.Unmarshal([]byte(bodyParts[6]), result); err != nil {
			return fmt.Errorf("解析数据JSON失败: %w", err)
		}
	}

	return nil
}

// ensureLogin 检查并确保用户已登录
func (c *Client) ensureLogin() error {
	if !c.loggedIn {
		return &Error{Code: ErrNoLogin, Message: errorMessages[ErrNoLogin]}
	}
	return nil
}
