# BaoStock 证券数据接口协议文档

## 1. 概述

BaoStock 是一个免费、开源的证券数据平台，提供完整的A股历史数据、实时行情、财务数据等。本文档详细描述了其客户端与服务器之间的通信协议。

**服务器信息:**
- 主机: `public-api.baostock.com`
- 端口: `10030`
- 传输协议: TCP
- 客户端版本: `00.9.10`

---

## 2. 消息格式

### 2.1 整体结构

```
[消息头]\x1[消息体]\x1[CRC32校验码]\n
```

**说明:**
- `\x1` (ASCII 0x01) - 字段分隔符 (MESSAGE_SPLIT)
- `\n` (ASCII 0x0A) - 消息结束符 (DELIMITER)
- 部分响应使用 `<![CDATA[]]>\n` 作为压缩消息的结束符

### 2.2 消息头结构 (21字节)

```
[版本号(6)]\x1[消息类型(2)]\x1[消息体长度(10)]
```

| 字段 | 长度 | 说明 | 示例 |
|------|------|------|------|
| 版本号 | 7字节 | 客户端版本号 | `00.9.10` |
| 分隔符 | 1字节 | `\x1` | - |
| 消息类型 | 2字节 | 请求/响应类型代码 | `00`, `01` 等 |
| 分隔符 | 1字节 | `\x1` | - |
| 消息体长度 | 10字节 | 消息体长度，左补0 | `0000000123` |

**消息头示例:** `00.9.10\x100\x10000000045`

---

## 3. 消息类型代码

### 3.1 登录/登出

| 代码 | 类型 | 说明 |
|------|------|------|
| `00` | 请求 | 登录请求 |
| `01` | 响应 | 登录响应 |
| `02` | 请求 | 登出请求 |
| `03` | 响应 | 登出响应 |
| `04` | 错误 | 错误信息 |

### 3.2 K线数据

| 代码 | 类型 | 说明 |
|------|------|------|
| `11` | 请求 | 获取历史K线数据请求 |
| `12` | 响应 | 获取历史K线数据响应 |
| `95` | 请求 | 获取历史K线数据Plus请求（行情压缩） |
| `96` | 响应 | 获取历史K线数据Plus响应（行情压缩） |

### 3.3 财务数据

| 代码 | 类型 | 说明 |
|------|------|------|
| `13`/`14` | 请求/响应 | 股息分红 |
| `15`/`16` | 请求/响应 | 复权因子数据 |
| `17`/`18` | 请求/响应 | 季频盈利能力 |
| `19`/`20` | 请求/响应 | 季频营运能力 |
| `21`/`22` | 请求/响应 | 季频成长能力 |
| `23`/`24` | 请求/响应 | 季频杜邦指数 |
| `25`/`26` | 请求/响应 | 季频偿债能力 |
| `27`/`28` | 请求/响应 | 季频现金流量 |

### 3.4 公司公告

| 代码 | 类型 | 说明 |
|------|------|------|
| `29`/`30` | 请求/响应 | 业绩报告 |
| `31`/`32` | 请求/响应 | 业绩预告 |

### 3.5 元数据查询

| 代码 | 类型 | 说明 |
|------|------|------|
| `33`/`34` | 请求/响应 | 交易日信息 |
| `35`/`36` | 请求/响应 | 所有证券信息 |
| `45`/`46` | 请求/响应 | 证券基本资料 |
| `59`/`60` | 请求/响应 | 行业分类 |
| `81`/`82` | 请求/响应 | 概念分类 |
| `83`/`84` | 请求/响应 | 地域分类 |

### 3.6 指数成分股

| 代码 | 类型 | 说明 |
|------|------|------|
| `61`/`62` | 请求/响应 | 沪深300成分股 |
| `63`/`64` | 请求/响应 | 上证50成分股 |
| `65`/`66` | 请求/响应 | 中证500成分股 |

### 3.7 特殊股票分类

| 代码 | 类型 | 说明 |
|------|------|------|
| `67`/`68` | 请求/响应 | 终止上市股票 |
| `69`/`70` | 请求/响应 | 暂停上市股票 |
| `71`/`72` | 请求/响应 | ST股票列表 |
| `73`/`74` | 请求/响应 | *ST股票列表 |
| `85`/`86` | 请求/响应 | 中小板分类 |
| `87`/`88` | 请求/响应 | 创业板分类 |
| `89`/`90` | 请求/响应 | 沪港通 |
| `91`/`92` | 请求/响应 | 深港通 |
| `93`/`94` | 请求/响应 | 风险警示板 |

### 3.8 宏观经济数据

| 代码 | 类型 | 说明 |
|------|------|------|
| `47`/`48` | 请求/响应 | 存款利率 |
| `49`/`50` | 请求/响应 | 贷款利率 |
| `51`/`52` | 请求/响应 | 存款准备金率 |
| `53`/`54` | 请求/响应 | 货币供应量（月度） |
| `55`/`56` | 请求/响应 | 货币供应量（年底余额） |
| `57`/`58` | 请求/响应 | SHIBOR利率 |
| `75`/`76` | 请求/响应 | CPI居民消费价格指数 |
| `77`/`78` | 请求/响应 | PPI工业品出厂价格指数 |
| `79`/`80` | 请求/响应 | PMI采购经理人指数 |

### 3.9 实时行情

| 代码 | 类型 | 说明 |
|------|------|------|
| `37`/`38` | 请求/响应 | 实时行情登录 |
| `39`/`40` | 请求/响应 | 实时行情登出 |
| `41`/`42` | 请求/响应 | 订阅行情 |
| `43`/`44` | 请求/响应 | 取消订阅 |

---

## 4. 消息体格式详解

### 4.1 通用请求格式

```
方法名\x1用户ID\x1当前页码\x1每页条数\x1[其他参数...]
```

**参数说明:**
- 大多数API遵循此格式
- 页码通常从1开始
- 每页条数默认10000
- 参数使用逗号分隔后转换为`\x1`分隔

### 4.2 登录请求 (类型00)

**请求格式:**
```
login\x1用户ID\x1密码\x1选项
```

**示例:**
```
login\x1anonymous\x1123456\x10
```

**响应格式:**
```
错误码\x1错误信息\x1方法名\x1用户ID
```

**成功响应示例:**
```
0\x1login success!\x1login\x1anonymous
```

### 4.3 登出请求 (类型02)

**请求格式:**
```
logout\x1用户ID\x1时间戳(YYYYMMDDHHmmss)
```

**示例:**
```
logout\x1anonymous\x120240110143000
```

### 4.4 历史K线Plus请求 (类型95)

**请求格式:**
```
query_history_k_data_plus\x1用户ID\x1页码\x1每页条数\x1证券代码\x1字段列表\x1开始日期\x1结束日期\x1频率\x1复权标志
```

**参数说明:**
- 证券代码格式: `sh.600000` 或 `sz.000001` (9位)
- 字段列表: 逗号分隔
  - 常用字段: `date,code,open,high,low,close,preclose,volume,amount,adjustflag,turn,tradestatus,pctChg,isST`
- 频率: `d`=日, `w`=周, `m`=月, `5`=5分钟, `15`=15分钟, `30`=30分钟, `60`=60分钟
- 复权标志: `1`=后复权, `2`=前复权, `3`=不复权

**响应格式:**
```
错误码\x1错误信息\x1方法名\x1用户ID\x1当前页\x1每页条数\x1数据JSON\x1代码\x1字段\x1开始日期\x1结束日期\x1频率\x1复权标志
```

**请求示例:**
```
query_history_k_data_plus\x1anonymous\x11\x110000\x1sh.600000\x1date,code,open,high,low,close,volume\x12023-01-01\x12023-12-31\x1d\x13
```

### 4.5 查询交易日 (类型33)

**请求格式:**
```
query_trade_dates\x1用户ID\x1页码\x1每页条数\x1开始日期\x1结束日期
```

**响应格式:**
```
错误码\x1错误信息\x1方法名\x1用户ID\x1当前页\x1每页条数\x1数据JSON\x1开始日期\x1结束日期\x1字段
```

### 4.6 查询所有证券 (类型35)

**请求格式:**
```
query_all_stock\x1用户ID\x1页码\x1每页条数\x1日期
```

**响应格式:**
```
错误码\x1错误信息\x1方法名\x1用户ID\x1当前页\x1每页条数\x1数据JSON\x1日期\x1字段
```

### 4.7 查询证券基本资料 (类型45)

**请求格式:**
```
query_stock_basic\x1用户ID\x1页码\x1每页条数\x1证券代码\x1证券名称
```

**响应格式:**
```
错误码\x1错误信息\x1方法名\x1用户ID\x1当前页\x1每页条数\x1数据JSON\x1代码\x1名称\x1字段
```

### 4.8 季频盈利能力 (类型17)

**请求格式:**
```
query_profit_data\x1用户ID\x1页码\x1每页条数\x1证券代码\x1年份\x1季度
```

**响应格式:**
```
错误码\x1错误信息\x1方法名\x1用户ID\x1当前页\x1每页条数\x1数据JSON\x1代码\x1年份\x1季度\x1字段
```

### 4.9 季频营运能力 (类型19)

**请求格式:**
```
query_operation_data\x1用户ID\x1页码\x1每页条数\x1证券代码\x1年份\x1季度
```

### 4.10 季频成长能力 (类型21)

**请求格式:**
```
query_growth_data\x1用户ID\x1页码\x1每页条数\x1证券代码\x1年份\x1季度
```

### 4.11 季频偿债能力 (类型25)

**请求格式:**
```
query_balance_data\x1用户ID\x1页码\x1每页条数\x1证券代码\x1年份\x1季度
```

### 4.12 季频现金流量 (类型27)

**请求格式:**
```
query_cash_flow_data\x1用户ID\x1页码\x1每页条数\x1证券代码\x1年份\x1季度
```

### 4.13 季频杜邦指数 (类型23)

**请求格式:**
```
query_dupont_data\x1用户ID\x1页码\x1每页条数\x1证券代码\x1年份\x1季度
```

### 4.14 股息分红 (类型13)

**请求格式:**
```
query_dividend_data\x1用户ID\x1页码\x1每页条数\x1证券代码\x1年份\x1年份类型
```

**年份类型:** `report`=预案公告年份, `operate`=除权除息年份

### 4.15 复权因子 (类型15)

**请求格式:**
```
query_adjust_factor\x1用户ID\x1页码\x1每页条数\x1证券代码\x1开始日期\x1结束日期
```

### 4.16 公司业绩报告 (类型29)

**请求格式:**
```
query_performance_express_report\x1用户ID\x1页码\x1每页条数\x1证券代码\x1开始日期\x1结束日期
```

### 4.17 公司业绩预告 (类型31)

**请求格式:**
```
query_forecast_report\x1用户ID\x1页码\x1每页条数\x1证券代码\x1开始日期\x1结束日期
```

### 4.18 行业分类 (类型59)

**请求格式:**
```
query_stock_industry\x1用户ID\x1页码\x1每页条数\x1证券代码\x1日期
```

### 4.19 概念分类 (类型81)

**请求格式:**
```
query_stock_concept\x1用户ID\x1页码\x1每页条数\x1证券代码\x1日期
```

### 4.20 地域分类 (类型83)

**请求格式:**
```
query_stock_area\x1用户ID\x1页码\x1每页条数\x1证券代码\x1日期
```

### 4.21 沪深300成分股 (类型61)

**请求格式:**
```
query_hs300_stocks\x1用户ID\x1页码\x1每页条数\x1日期
```

### 4.22 上证50成分股 (类型63)

**请求格式:**
```
query_sz50_stocks\x1用户ID\x1页码\x1每页条数\x1日期
```

### 4.23 中证500成分股 (类型65)

**请求格式:**
```
query_zz500_stocks\x1用户ID\x1页码\x1每页条数\x1日期
```

### 4.24 存款利率 (类型47)

**请求格式:**
```
query_deposit_rate_data\x1用户ID\x1页码\x1每页条数\x1开始日期\x1结束日期
```

### 4.25 贷款利率 (类型49)

**请求格式:**
```
query_loan_rate_data\x1用户ID\x1页码\x1每页条数\x1开始日期\x1结束日期
```

### 4.26 存款准备金率 (类型51)

**请求格式:**
```
query_required_reserve_ratio_data\x1用户ID\x1页码\x1每页条数\x1开始日期\x1结束日期\x1年份类型
```

**年份类型:** `0`=公告日期, `1`=生效日期

### 4.27 货币供应量月度 (类型53)

**请求格式:**
```
query_money_supply_data_month\x1用户ID\x1页码\x1每页条数\x1开始年月\x1结束年月
```

**日期格式:** `yyyy-MM`

### 4.28 货币供应量年度 (类型55)

**请求格式:**
```
query_money_supply_data_year\x1用户ID\x1页码\x1每页条数\x1开始年份\x1结束年份
```

**日期格式:** `yyyy`

### 4.29 CPI指数 (类型75)

**请求格式:**
```
query_cpi_data\x1用户ID\x1页码\x1每页条数\x1开始日期\x1结束日期
```

### 4.30 PPI指数 (类型77)

**请求格式:**
```
query_ppi_data\x1用户ID\x1页码\x1每页条数\x1开始日期\x1结束日期
```

### 4.31 PMI指数 (类型79)

**请求格式:**
```
query_pmi_data\x1用户ID\x1页码\x1每页条数\x1开始日期\x1结束日期
```

### 4.32 中小板分类 (类型85)

**请求格式:**
```
query_ame_stocks\x1用户ID\x1页码\x1每页条数\x1日期
```

### 4.33 创业板分类 (类型87)

**请求格式:**
```
query_gem_stocks\x1用户ID\x1页码\x1每页条数\x1日期
```

### 4.34 沪港通 (类型89)

**请求格式:**
```
query_shhk_stocks\x1用户ID\x1页码\x1每页条数\x1日期
```

### 4.35 深港通 (类型91)

**请求格式:**
```
query_szhk_stocks\x1用户ID\x1页码\x1每页条数\x1日期
```

### 4.36 风险警示板 (类型93)

**请求格式:**
```
query_stocks_in_risk\x1用户ID\x1页码\x1每页条数\x1日期
```

### 4.37 ST股票 (类型71)

**请求格式:**
```
query_st_stocks\x1用户ID\x1页码\x1每页条数\x1日期
```

### 4.38 *ST股票 (类型73)

**请求格式:**
```
query_starst_stocks\x1用户ID\x1页码\x1每页条数\x1日期
```

### 4.39 终止上市股票 (类型67)

**请求格式:**
```
query_terminated_stocks\x1用户ID\x1页码\x1每页条数\x1日期
```

### 4.40 暂停上市股票 (类型69)

**请求格式:**
```
query_suspended_stocks\x1用户ID\x1页码\x1每页条数\x1日期
```

---

## 5. 数据格式

### 5.1 JSON数据格式

数据字段使用JSON格式传输，结构为:

```json
{
    "record": [
        ["2023-01-03", "sh.600000", "3.15", "3.17", "3.13", "3.16", "1234567"],
        ["2023-01-04", "sh.600000", "3.16", "3.18", "3.14", "3.15", "2345678"]
    ]
}
```

### 5.2 分页机制

- 默认每页10000条记录
- 使用 `next()` 方法获取下一页数据
- 当 `cur_row_num` 达到 `per_page_count` 时自动请求下一页

---

## 6. 压缩与校验

### 6.1 CRC32校验

```python
import zlib

# 计算校验码
crc32str = zlib.crc32(bytes(消息头 + 消息体, encoding='utf-8'))
```

### 6.2 数据压缩

- 支持zlib压缩的消息类型: `MESSAGE_TYPE_GETKDATAPLUS_RESPONSE` (类型96)
- 压缩消息结束符: `<![CDATA[]]>\n`
- 解压方法: `zlib.decompress()`

**压缩消息处理流程:**
1. 检查消息头中的消息类型是否在 `COMPRESSED_MESSAGE_TYPE_TUPLE` 中
2. 如果是压缩类型，获取消息体长度
3. 使用 `zlib.decompress()` 解压消息体
4. 拼接消息头和解压后的消息体

---

## 7. 错误代码

### 7.1 成功代码

| 错误码 | 说明 |
|--------|------|
| `0` | 成功 |

### 7.2 登录相关错误 (100010xx)

| 错误码 | 说明 |
|--------|------|
| `10001001` | 用户未登录 |
| `10001002` | 用户名或密码错误 |
| `10001003` | 获取用户信息失败 |
| `10001004` | 客户端版本号过期 |
| `10001005` | 账号登录数达到上限 |
| `10001006` | 用户权限不足 |
| `10001007` | 需要登录激活 |
| `10001008` | 用户名为空 |
| `10001009` | 密码为空 |
| `10001011` | 黑名单用户 |

### 7.3 网络相关错误 (100020xx)

| 错误码 | 说明 |
|--------|------|
| `10002001` | 网络错误 |
| `10002002` | 网络连接失败 |
| `10002003` | 网络连接超时 |
| `10002004` | 网络接收时连接断开 |
| `10002005` | 网络发送失败 |
| `10002006` | 网络发送超时 |
| `10002007` | 网络接收错误 |
| `10002008` | 网络接收超时 |

### 7.4 客户端相关错误 (100040xx)

| 错误码 | 说明 |
|--------|------|
| `10004001` | 解析数据错误 |
| `10004002` | gzip解压失败 |
| `10004003` | 客户端未知错误 |
| `10004004` | 数组越界 |
| `10004005` | 传入参数为空 |
| `10004006` | 参数错误 |
| `10004007` | 起始日期格式不正确 |
| `10004008` | 截止日期格式不正确 |
| `10004009` | 起始日期大于终止日期 |
| `10004010` | 日期格式不正确 |
| `10004011` | 无效的证券代码 |
| `10004012` | 无效的指标 |
| `10004013` | 超出日期支持范围 |
| `10004014` | 不支持的混合证券品种 |
| `10004015` | 不支持的证券代码品种 |
| `10004016` | 交易条数超过上限 |
| `10004017` | 不支持的交易信息 |
| `10004018` | 指标重复 |
| `10004019` | 消息格式不正确 |
| `10004020` | 错误的消息类型 |

### 7.5 系统错误 (100050xx)

| 错误码 | 说明 |
|--------|------|
| `10005001` | 系统级别错误 |

---

## 8. 通信流程

### 8.1 基本流程

```
1. 创建TCP连接 (public-api.baostock.com:10030)
2. 发送登录请求
3. 接收登录响应
4. 发送业务请求
5. 接收业务响应
6. 处理分页数据 (如需要)
7. 发送登出请求
8. 关闭连接
```

### 8.2 登录示例

**请求:**
```
00.9.10\x100\x1000000033\x1login\x1anonymous\x1123456\x10\x1-1234567890\n
```

**响应 (成功):**
```
00.9.10\x101\x1000000048\x10\x1login success!\x1login\x1anonymous\n
```

### 8.3 K线数据请求示例

**请求:**
```
00.9.10\x195\x1000000160\x1query_history_k_data_plus\x1anonymous\x11\x110000\x1sh.600000\x1date,code,open,high,low,close,volume\x12023-01-01\x12023-12-31\x1d\x13\x11234567890\n
```

---

## 9. 常量定义

```python
# 服务器配置
BAOSTOCK_SERVER_IP = "public-api.baostock.com"
BAOSTOCK_SERVER_PORT = 10030
BAOSTOCK_CLIENT_VERSION = "00.9.10"

# 消息分隔符
MESSAGE_SPLIT = "\x1"           # 字段分隔符
DELIMITER = "\n"                # 消息结束符
ATTRIBUTE_SPLIT = ","           # 属性分隔符

# 消息结构
MESSAGE_HEADER_LENGTH = 21      # 消息头长度
MESSAGE_HEADER_BODYLENGTH = 10  # 消息体长度字段位数

# 分页
BAOSTOCK_PER_PAGE_COUNT = 10000 # 默认每页条数
BAOSTOCK_REALTIME_LIMIT_COUNT = 500  # 实时行情代码限制

# 证券代码
STOCK_CODE_LENGTH = 9           # 证券代码长度

# 默认值
DEFAULT_START_DATE = "2015-01-01"  # 默认开始时间

# 压缩消息类型
COMPRESSED_MESSAGE_TYPE_TUPLE = ("96",)  # 支持压缩的消息类型
```

---

## 10. 实现注意事项

1. **连接复用**: 登录后保持连接，复用socket进行多次请求
2. **分页处理**: 大量数据需要分页获取，使用next()获取后续页
3. **日期格式**: 所有日期使用 `YYYY-MM-DD` 格式
4. **证券代码**: 必须为9位，格式如 `sh.600000`
5. **字符编码**: 使用UTF-8编码
6. **错误处理**: 检查error_code是否为"0"判断请求成功
7. **消息长度**: 消息体长度必须左补0到10位
8. **CRC32计算**: 对消息头+消息体（不含分隔符和CRC）计算
9. **接收缓冲**: 使用8192字节缓冲区接收数据，循环直到收到完整消息
10. **消息结束判断**: 检查消息尾部的 `<![CDATA[]]>\n` 或 `\n` 判断接收完成

---

## 11. Python示例代码

### 11.1 基本连接

```python
import socket
import zlib

def send_message(sock, msg_type, msg_body):
    version = "00.9.10"
    header = f"{version}\x1{msg_type}\x1{str(len(msg_body)).zfill(10)}"
    full_msg = header + msg_body
    crc32 = zlib.crc32(bytes(full_msg, 'utf-8'))
    request = full_msg + f"\x1{crc32}\n"
    sock.send(bytes(request, 'utf-8'))

    response = b""
    while True:
        recv = sock.recv(8192)
        response += recv
        if response.endswith(b"<![CDATA[]]>\n") or response.endswith(b"\n"):
            break
    return response.decode('utf-8')

# 连接服务器
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect(("public-api.baostock.com", 10030))

# 登录
login_body = "login\x1anonymous\x1123456\x10"
response = send_message(sock, "00", login_body)
print(response)
```

### 11.2 完整数据查询示例

```python
import socket
import zlib
import json

class BaoStockClient:
    def __init__(self):
        self.sock = None
        self.user_id = "anonymous"
        self.version = "00.8.90"

    def connect(self):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect(("public-api.baostock.com", 10030))

    def send_request(self, msg_type, msg_body):
        header = f"{self.version}\x1{msg_type}\x1{str(len(msg_body)).zfill(10)}"
        full_msg = header + msg_body
        crc32 = zlib.crc32(bytes(full_msg, 'utf-8'))
        request = full_msg + f"\x1{crc32}\n"
        self.sock.send(bytes(request, 'utf-8'))

        response = b""
        while True:
            recv = self.sock.recv(8192)
            response += recv
            if response.endswith(b"\n"):
                break

        # 检查是否需要解压
        response_str = response.decode('utf-8')
        if msg_type == "96":  # 压缩响应
            header = response[0:21]
            body_length = int(header.split('\x1')[2])
            body = zlib.decompress(response[21:21+body_length]).decode('utf-8')
            return header.decode('utf-8') + body

        return response_str

    def login(self):
        msg_body = f"login\x1{self.user_id}\x1123456\x10"
        response = self.send_request("00", msg_body)
        parts = response.split('\x1')
        if parts[4] == '0':
            print("登录成功")
            return True
        else:
            print(f"登录失败: {parts[5]}")
            return False

    def query_k_data(self, code, fields, start_date, end_date, frequency='d', adjustflag='3'):
        msg_body = f"query_history_k_data_plus\x1{self.user_id}\x11\x110000\x1{code}\x1{fields}\x1{start_date}\x1{end_date}\x1{frequency}\x1{adjustflag}"
        response = self.send_request("95", msg_body)

        # 解析响应
        parts = response.split('\x1')
        error_code = parts[4]

        if error_code == '0':
            data_json = parts[10]
            data = json.loads(data_json)
            return data['record']
        else:
            print(f"查询失败: {parts[5]}")
            return None

    def logout(self):
        import datetime
        timestamp = datetime.datetime.now().strftime('%Y%m%d%H%M%S')
        msg_body = f"logout\x1{self.user_id}\x1{timestamp}"
        response = self.send_request("02", msg_body)
        print("登出")

    def close(self):
        if self.sock:
            self.sock.close()

# 使用示例
client = BaoStockClient()
client.connect()
if client.login():
    data = client.query_k_data(
        code="sh.600000",
        fields="date,code,open,high,low,close,volume",
        start_date="2023-01-01",
        end_date="2023-12-31"
    )
    if data:
        for row in data:
            print(row)
    client.logout()
client.close()
```

---

## 12. API 快速参考

| 功能 | 消息类型 | Python函数名 |
|------|----------|-------------|
| 登录 | 00/01 | login() |
| 登出 | 02/03 | logout() |
| 历史K线 | 95/96 | query_history_k_data_plus() |
| 交易日 | 33/34 | query_trade_dates() |
| 所有股票 | 35/36 | query_all_stock() |
| 股票信息 | 45/46 | query_stock_basic() |
| 行业分类 | 59/60 | query_stock_industry() |
| 盈利能力 | 17/18 | query_profit_data() |
| 营运能力 | 19/20 | query_operation_data() |
| 成长能力 | 21/22 | query_growth_data() |
| 偿债能力 | 25/26 | query_balance_data() |
| 现金流量 | 27/28 | query_cash_flow_data() |
| 复权因子 | 15/16 | query_adjust_factor() |
| 股息分红 | 13/14 | query_dividend_data() |
| 业绩报告 | 29/30 | query_performance_express_report() |
| 业绩预告 | 31/32 | query_forecast_report() |
| 沪深300 | 61/62 | query_hs300_stocks() |
| 上证50 | 63/64 | query_sz50_stocks() |
| 中证500 | 65/66 | query_zz500_stocks() |
| 存款利率 | 47/48 | query_deposit_rate_data() |
| 贷款利率 | 49/50 | query_loan_rate_data() |
| 存款准备金率 | 51/52 | query_required_reserve_ratio_data() |
| CPI | 75/76 | query_cpi_data() |
| PPI | 77/78 | query_ppi_data() |
| PMI | 79/80 | query_pmi_data() |
