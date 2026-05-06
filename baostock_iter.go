package baostock

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// QueryHistoryKDataPlusIter 查询历史K线数据（迭代器版本，Go 1.23+）
//
// 使用示例:
//
//	for fields, record := range client.QueryHistoryKDataPlusIter(ctx,
//	    &baostock.HistoryKDataRequest{
//	        Code:      "sh.600000",
//	        Fields:    strings.Join(baostock.DailyKLineCommonFields, ","),
//	        StartDate: "2020-01-01",
//	        EndDate:   "2023-12-31",
//	        Frequency: baostock.FrequencyDaily,
//	    }) {
//	    if !fields {
//	        break // 迭代被 yield 终止或发生错误
//	    }
//	    fmt.Printf("日期: %s, 收盘: %s\n", record[0], record[5])
//	}
//
// 提前终止迭代：
//   - 在循环体中使用 break 或 return
//   - yield 返回 false 时自动停止
func (c *Client) QueryHistoryKDataPlusIter(ctx context.Context, req *HistoryKDataRequest) func(yield func(fields []string, record []string) bool) {
	return func(yield func(fields []string, record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil, nil) // 错误时调用一次 yield(false) 终止
			return
		}

		if err := validateStockCode(req.Code); err != nil {
			yield(nil, nil)
			return
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
				return
			default:
			}

			// 构建请求消息
			msgBody := fmt.Sprintf("query_history_k_data_plus%s%s%s%d%s%d%s%s%s%s%s%s%s%s%s%s%s%s",
				MessageSplit, c.userID, MessageSplit, pageNum, MessageSplit,
				DefaultPerPageCount, MessageSplit, code, MessageSplit,
				req.Fields, MessageSplit, startDate, MessageSplit,
				endDate, MessageSplit, req.Frequency, MessageSplit, req.AdjustFlag)

			resp, err := c.sendMessage(ctx, MsgTypeGetKDataPlusRequest, msgBody)
			if err != nil {
				return
			}

			result, err := parseHistoryKDataResponse(resp)
			if err != nil {
				return
			}

			if result.ErrorCode != ErrSuccess {
				return
			}

			// 第一页保存字段列表
			if pageNum == 1 {
				allFields = result.Fields
			}

			// 处理当前页的每条记录
			for _, record := range result.Data {
				if !yield(allFields, record) {
					return // yield 返回 false，停止迭代
				}
			}

			// 检查是否还有更多页
			if len(result.Data) < DefaultPerPageCount {
				break
			}

			pageNum++
		}
	}
}

// QueryTradeDatesIter 查询交易日（迭代器版本）
func (c *Client) QueryTradeDatesIter(ctx context.Context, startDate, endDate string) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
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
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		// 使用迭代器风格的 JSON 解析
		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryAllStockIter 查询指定日期的所有股票（迭代器版本）
func (c *Client) QueryAllStockIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		if date == "" {
			date = time.Now().Format("2006-01-02")
		}

		msgBody := fmt.Sprintf("query_all_stock%s%s%s1%s%d%s%s",
			MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, date)

		resp, err := c.sendMessage(ctx, MsgTypeQueryAllStockRequest, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryStockBasicIter 查询股票基本信息（迭代器版本）
func (c *Client) QueryStockBasicIter(ctx context.Context, code, codeName string) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		if code != "" {
			if err := validateStockCode(code); err != nil {
				yield(nil)
				return
			}
		}

		normalizedCode := normalizeStockCode(code)

		msgBody := fmt.Sprintf("query_stock_basic%s%s%s1%s%d%s%s%s%s",
			MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, normalizedCode, MessageSplit, codeName)

		resp, err := c.sendMessage(ctx, MsgTypeQueryStockBasicRequest, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryStockIndustryIter 查询行业分类（迭代器版本）
func (c *Client) QueryStockIndustryIter(ctx context.Context, code, date string) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		if code != "" {
			if err := validateStockCode(code); err != nil {
				yield(nil)
				return
			}
		}

		normalizedCode := normalizeStockCode(code)

		msgBody := fmt.Sprintf("query_stock_industry%s%s%s1%s%d%s%s%s%s",
			MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, normalizedCode, MessageSplit, date)

		resp, err := c.sendMessage(ctx, MsgTypeQueryStockIndustryRequest, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryHS300StocksIter 查询沪深300成分股（迭代器版本）
func (c *Client) QueryHS300StocksIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.queryIndexStocksIter(ctx, MsgTypeQueryHS300StocksRequest, "query_hs300_stocks", date)
}

// QuerySZ50StocksIter 查询上证50成分股（迭代器版本）
func (c *Client) QuerySZ50StocksIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.queryIndexStocksIter(ctx, MsgTypeQuerySZ50StocksRequest, "query_sz50_stocks", date)
}

// QueryZZ500StocksIter 查询中证500成分股（迭代器版本）
func (c *Client) QueryZZ500StocksIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.queryIndexStocksIter(ctx, MsgTypeQueryZZ500StocksRequest, "query_zz500_stocks", date)
}

// queryIndexStocksIter 指数成分股查询辅助方法（迭代器版本）
func (c *Client) queryIndexStocksIter(ctx context.Context, msgType, methodName, date string) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		if date == "" {
			date = time.Now().Format("2006-01-02")
		}

		msgBody := fmt.Sprintf("%s%s%s%s1%s%d%s%s",
			methodName, MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, date)

		resp, err := c.sendMessage(ctx, msgType, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryDepositRateDataIter 查询存款利率数据（迭代器版本）
func (c *Client) QueryDepositRateDataIter(ctx context.Context, startDate, endDate string) func(yield func(record []string) bool) {
	return c.queryEconomicDataIter(ctx, MsgTypeQueryDepositRateDataRequest, "query_deposit_rate_data", startDate, endDate, "")
}

// QueryLoanRateDataIter 查询贷款利率数据（迭代器版本）
func (c *Client) QueryLoanRateDataIter(ctx context.Context, startDate, endDate string) func(yield func(record []string) bool) {
	return c.queryEconomicDataIter(ctx, MsgTypeQueryLoanRateDataRequest, "query_loan_rate_data", startDate, endDate, "")
}

// QueryCPIDataIter 查询CPI数据（迭代器版本）
func (c *Client) QueryCPIDataIter(ctx context.Context, startDate, endDate string) func(yield func(record []string) bool) {
	return c.queryEconomicDataIter(ctx, MsgTypeQueryCPIDataRequest, "query_cpi_data", startDate, endDate, "")
}

// QueryPPIDataIter 查询PPI数据（迭代器版本）
func (c *Client) QueryPPIDataIter(ctx context.Context, startDate, endDate string) func(yield func(record []string) bool) {
	return c.queryEconomicDataIter(ctx, MsgTypeQueryPPIDataRequest, "query_ppi_data", startDate, endDate, "")
}

// QueryPMIDataIter 查询PMI数据（迭代器版本）
func (c *Client) QueryPMIDataIter(ctx context.Context, startDate, endDate string) func(yield func(record []string) bool) {
	return c.queryEconomicDataIter(ctx, MsgTypeQueryPMIDataRequest, "query_pmi_data", startDate, endDate, "")
}

// queryEconomicDataIter 经济数据查询辅助方法（迭代器版本）
func (c *Client) queryEconomicDataIter(ctx context.Context, msgType, methodName, startDate, endDate, extraParam string) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		msgBody := fmt.Sprintf("%s%s%s%s1%s%d%s%s%s%s",
			methodName, MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, startDate, MessageSplit, endDate)

		if extraParam != "" {
			msgBody += MessageSplit + extraParam
		}

		resp, err := c.sendMessage(ctx, msgType, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryAdjustFactorIter 查询复权因子数据（迭代器版本）
func (c *Client) QueryAdjustFactorIter(ctx context.Context, req *AdjustFactorRequest) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		if err := validateStockCode(req.Code); err != nil {
			yield(nil)
			return
		}

		startDate := req.StartDate
		if startDate == "" {
			startDate = DefaultStartDate
		}

		endDate := req.EndDate
		if endDate == "" {
			endDate = GetCurrentDate()
		}

		normalizedCode := normalizeStockCode(req.Code)

		msgBody := fmt.Sprintf("query_adjust_factor%s%s%s%s1%d%s%s%s%s%s%s",
			MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, normalizedCode, MessageSplit, startDate, MessageSplit, endDate)

		resp, err := c.sendMessage(ctx, MsgTypeAdjustFactorRequest, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryPerformanceExpressReportIter 查询业绩快报（迭代器版本）
func (c *Client) QueryPerformanceExpressReportIter(ctx context.Context, req *ReportDataRequest) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		if err := validateStockCode(req.Code); err != nil {
			yield(nil)
			return
		}

		startDate := req.StartDate
		if startDate == "" {
			startDate = DefaultStartDate
		}

		endDate := req.EndDate
		if endDate == "" {
			endDate = GetCurrentDate()
		}

		msgBody := fmt.Sprintf("query_performance_express_report%s%s%s%s1%d%s%s%s%s%s%s",
			MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, req.Code, MessageSplit, startDate, MessageSplit, endDate)

		resp, err := c.sendMessage(ctx, MsgTypeQueryPerformanceExpressReportRequest, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryForecastReportIter 查询业绩预告（迭代器版本）
func (c *Client) QueryForecastReportIter(ctx context.Context, req *ReportDataRequest) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		if err := validateStockCode(req.Code); err != nil {
			yield(nil)
			return
		}

		startDate := req.StartDate
		if startDate == "" {
			startDate = DefaultStartDate
		}

		endDate := req.EndDate
		if endDate == "" {
			endDate = GetCurrentDate()
		}

		msgBody := fmt.Sprintf("query_forecast_report%s%s%s%s1%d%s%s%s%s%s%s",
			MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, req.Code, MessageSplit, startDate, MessageSplit, endDate)

		resp, err := c.sendMessage(ctx, MsgTypeQueryForecastReportRequest, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryStockConceptIter 查询概念分类（迭代器版本）
func (c *Client) QueryStockConceptIter(ctx context.Context, code, date string) func(yield func(record []string) bool) {
	return c.queryStockClassificationIter(ctx, MsgTypeQueryStockConceptRequest, code, date, "query_stock_concept")
}

// QueryStockAreaIter 查询地域分类（迭代器版本）
func (c *Client) QueryStockAreaIter(ctx context.Context, code, date string) func(yield func(record []string) bool) {
	return c.queryStockClassificationIter(ctx, MsgTypeQueryStockAreaRequest, code, date, "query_stock_area")
}

// queryStockClassificationIter 股票分类查询辅助方法（迭代器版本）
func (c *Client) queryStockClassificationIter(ctx context.Context, msgType, code, date, methodName string) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		if code != "" {
			if err := validateStockCode(code); err != nil {
				yield(nil)
				return
			}
		}

		msgBody := fmt.Sprintf("%s%s%s%s1%s%d%s%s%s%s",
			methodName, MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, code, MessageSplit, date)

		resp, err := c.sendMessage(ctx, msgType, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryTerminatedStocksIter 查询终止上市股票（迭代器版本）
func (c *Client) QueryTerminatedStocksIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.querySpecialStocksIter(ctx, MsgTypeQueryTerminatedStocksRequest, "query_terminated_stocks", date)
}

// QuerySuspendedStocksIter 查询暂停上市股票（迭代器版本）
func (c *Client) QuerySuspendedStocksIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.querySpecialStocksIter(ctx, MsgTypeQuerySuspendedStocksRequest, "query_suspended_stocks", date)
}

// QuerySTStocksIter 查询ST股票（迭代器版本）
func (c *Client) QuerySTStocksIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.querySpecialStocksIter(ctx, MsgTypeQuerySTStocksRequest, "query_st_stocks", date)
}

// QueryStarSTStocksIter 查询*ST股票（迭代器版本）
func (c *Client) QueryStarSTStocksIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.querySpecialStocksIter(ctx, MsgTypeQueryStarSTStocksRequest, "query_starst_stocks", date)
}

// QueryAMEStocksIter 查询中小板股票（迭代器版本）
func (c *Client) QueryAMEStocksIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.querySpecialStocksIter(ctx, MsgTypeQueryAMEStocksRequest, "query_ame_stocks", date)
}

// QueryGEMStocksIter 查询创业板股票（迭代器版本）
func (c *Client) QueryGEMStocksIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.querySpecialStocksIter(ctx, MsgTypeQueryGEMStocksRequest, "query_gem_stocks", date)
}

// QuerySHHKStocksIter 查询沪港通股票（迭代器版本）
func (c *Client) QuerySHHKStocksIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.querySpecialStocksIter(ctx, MsgTypeQuerySHHKStocksRequest, "query_shhk_stocks", date)
}

// QuerySZHKStocksIter 查询深港通股票（迭代器版本）
func (c *Client) QuerySZHKStocksIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.querySpecialStocksIter(ctx, MsgTypeQuerySZHKStocksRequest, "query_szhk_stocks", date)
}

// QueryStocksInRiskIter 查询风险警示板股票（迭代器版本）
func (c *Client) QueryStocksInRiskIter(ctx context.Context, date string) func(yield func(record []string) bool) {
	return c.querySpecialStocksIter(ctx, MsgTypeQueryStockInRiskRequest, "query_stocks_in_risk", date)
}

// querySpecialStocksIter 特殊股票查询辅助方法（迭代器版本）
func (c *Client) querySpecialStocksIter(ctx context.Context, msgType, methodName string, date string) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		if date == "" {
			date = GetCurrentDate()
		}

		msgBody := fmt.Sprintf("%s%s%s%s1%s%d%s%s",
			methodName, MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, date)

		resp, err := c.sendMessage(ctx, msgType, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryRequiredReserveRatioDataIter 查询存款准备金率数据（迭代器版本）
func (c *Client) QueryRequiredReserveRatioDataIter(ctx context.Context, startDate, endDate, yearType string) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		if yearType == "" {
			yearType = "0"
		}

		msgBody := fmt.Sprintf("query_required_reserve_ratio_data%s%s%s%s1%d%s%s%s%s%s%s",
			MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, startDate, MessageSplit, endDate, MessageSplit, yearType)

		resp, err := c.sendMessage(ctx, MsgTypeQueryRequiredReserveRatioDataRequest, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryMoneySupplyDataMonthIter 查询月度货币供应量数据（迭代器版本）
func (c *Client) QueryMoneySupplyDataMonthIter(ctx context.Context, startDate, endDate string) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		msgBody := fmt.Sprintf("query_money_supply_data_month%s%s%s1%s%d%s%s%s%s",
			MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, startDate, MessageSplit, endDate)

		resp, err := c.sendMessage(ctx, MsgTypeQueryMoneySupplyDataMonthRequest, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}

// QueryMoneySupplyDataYearIter 查询年度货币供应量数据（迭代器版本）
func (c *Client) QueryMoneySupplyDataYearIter(ctx context.Context, startDate, endDate string) func(yield func(record []string) bool) {
	return func(yield func(record []string) bool) {
		if err := c.ensureLogin(); err != nil {
			yield(nil)
			return
		}

		msgBody := fmt.Sprintf("query_money_supply_data_year%s%s%s1%s%d%s%s%s%s",
			MessageSplit, c.userID, MessageSplit, MessageSplit,
			DefaultPerPageCount, MessageSplit, startDate, MessageSplit, endDate)

		resp, err := c.sendMessage(ctx, MsgTypeQueryMoneySupplyDataYearRequest, msgBody)
		if err != nil {
			yield(nil)
			return
		}

		bodyParts := strings.Split(resp.Body, MessageSplit)
		if len(bodyParts) < 7 {
			yield(nil)
			return
		}

		errorCode := bodyParts[0]
		if errorCode != ErrSuccess {
			yield(nil)
			return
		}

		dec := newRecordIterator(bodyParts[6])
		for dec.Next() {
			if !yield(dec.Record()) {
				return
			}
		}
	}
}
