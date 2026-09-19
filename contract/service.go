// Package contract 保存 Pulse 可被任意层引用的无依赖契约。
// 本包不得导入其他包，也不得包含数据库模型、运行时对象或请求级状态。
package contract

// ServiceName 是 Pulse 服务在配置、ServiceContext 和内部服务通信中的稳定名称。
const ServiceName = "pulse"

// ProductName 是面向评委与黑客松提交的产品名。
const ProductName = "Pulse Screener"

// CMCAPIKeyEnv 是 CoinMarketCap Pro API 密钥的唯一环境变量名。
const CMCAPIKeyEnv = "CMC_API_KEY"

// AllowSampleEnv 在缺少真实密钥时允许加载离线样例行情。
const AllowSampleEnv = "PULSE_ALLOW_SAMPLE"
