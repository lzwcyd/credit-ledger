package decimal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Decimal 高精度定点数，内部使用整数存储，精度为4位小数
// 例如：123.4567 存储为 1234567
type Decimal struct {
	value    int64
	precision int
}

const (
	// DefaultPrecision 默认精度（小数位数）
	DefaultPrecision = 4
)

// New 创建新的 Decimal
func New(value int64, precision int) Decimal {
	return Decimal{value: value, precision: precision}
}

// NewFromString 从字符串创建 Decimal
func NewFromString(s string) (Decimal, error) {
	return Parse(s, DefaultPrecision)
}

// NewFromFloat 从浮点数创建 Decimal
func NewFromFloat(f float64) Decimal {
	return FromFloat(f, DefaultPrecision)
}

// NewFromInt 从整数创建 Decimal
func NewFromInt(i int64) Decimal {
	return New(i * pow10(DefaultPrecision), DefaultPrecision)
}

// Parse 解析字符串为 Decimal
func Parse(s string, precision int) (Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Decimal{}, fmt.Errorf("empty string")
	}

	// 处理负号
	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	} else if s[0] == '+' {
		s = s[1:]
	}

	// 分割整数和小数部分
	parts := strings.Split(s, ".")
	intPart := parts[0]
	decPart := ""
	if len(parts) > 1 {
		decPart = parts[1]
	}

	// 补齐或截断小数部分
	if len(decPart) < precision {
		decPart = decPart + strings.Repeat("0", precision-len(decPart))
	} else if len(decPart) > precision {
		decPart = decPart[:precision]
	}

	// 组合为整数
	valueStr := intPart + decPart
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return Decimal{}, fmt.Errorf("invalid number: %s", s)
	}

	if negative {
		value = -value
	}

	return New(value, precision), nil
}

// FromFloat 从浮点数创建 Decimal
func FromFloat(f float64, precision int) Decimal {
	multiplier := pow10(precision)
	value := int64(math.Round(f * float64(multiplier)))
	return New(value, precision)
}

// Float64 转换为 float64
func (d Decimal) Float64() float64 {
	multiplier := pow10(d.precision)
	return float64(d.value) / float64(multiplier)
}

// Int64 获取整数值（舍去小数）
func (d Decimal) Int64() int64 {
	multiplier := pow10(d.precision)
	return d.value / multiplier
}

// String 转换为字符串
func (d Decimal) String() string {
	negative := d.value < 0
	value := d.value
	if negative {
		value = -value
	}

	multiplier := pow10(d.precision)
	intPart := value / multiplier
	decPart := value % multiplier

	if d.precision == 0 {
		if negative {
			return fmt.Sprintf("-%d", intPart)
		}
		return fmt.Sprintf("%d", intPart)
	}

	if negative {
		return fmt.Sprintf("-%d.%0*d", intPart, d.precision, decPart)
	}
	return fmt.Sprintf("%d.%0*d", intPart, d.precision, decPart)
}

// Abs 绝对值
func (d Decimal) Abs() Decimal {
	if d.value < 0 {
		return New(-d.value, d.precision)
	}
	return d
}

// Neg 负数
func (d Decimal) Neg() Decimal {
	return New(-d.value, d.precision)
}

// IsZero 是否为零
func (d Decimal) IsZero() bool {
	return d.value == 0
}

// IsNegative 是否为负数
func (d Decimal) IsNegative() bool {
	return d.value < 0
}

// IsPositive 是否为正数
func (d Decimal) IsPositive() bool {
	return d.value > 0}

// Cmp 比较两个 Decimal
// 返回: -1 (d < other), 0 (d == other), 1 (d > other)
func (d Decimal) Cmp(other Decimal) int {
	// 统一精度
	d1, d2 := alignPrecision(d, other)
	if d1.value < d2.value {
		return -1
	}
	if d1.value > d2.value {
		return 1
	}
	return 0
}

// Eq 等于
func (d Decimal) Eq(other Decimal) bool {
	return d.Cmp(other) == 0
}

// Lt 小于
func (d Decimal) Lt(other Decimal) bool {
	return d.Cmp(other) < 0
}

// Lte 小于等于
func (d Decimal) Lte(other Decimal) bool {
	return d.Cmp(other) <= 0
}

// Gt 大于
func (d Decimal) Gt(other Decimal) bool {
	return d.Cmp(other) > 0
}

// Gte 大于等于
func (d Decimal) Gte(other Decimal) bool {
	return d.Cmp(other) >= 0
}

// Add 加法
func (d Decimal) Add(other Decimal) Decimal {
	d1, d2 := alignPrecision(d, other)
	return New(d1.value+d2.value, d1.precision)
}

// Sub 减法
func (d Decimal) Sub(other Decimal) Decimal {
	d1, d2 := alignPrecision(d, other)
	return New(d1.value-d2.value, d1.precision)
}

// Mul 乘法
func (d Decimal) Mul(other Decimal) Decimal {
	// 直接相乘，然后调整精度
	result := d.value * other.value
	precision := d.precision + other.precision
	
	// 调整到目标精度
	if precision > DefaultPrecision {
		divisor := pow10(precision - DefaultPrecision)
		result = (result + divisor/2) / divisor // 四舍五入
		precision = DefaultPrecision
	} else if precision < DefaultPrecision {
		result *= pow10(DefaultPrecision - precision)
		precision = DefaultPrecision
	}
	
	return New(result, precision)
}

// MulInt 乘以整数
func (d Decimal) MulInt(i int64) Decimal {
	return New(d.value*i, d.precision)
}

// Div 除法
func (d Decimal) Div(other Decimal) Decimal {
	if other.IsZero() {
		panic("division by zero")
	}
	
	// 使用更高精度计算，避免精度损失
	// 将被除数放大，然后除以除数
	// 结果再调整到目标精度
	
	// 放大被除数
	放大倍数 := int64(100000000) // 10^8
	被放大 := d.value * 放大倍数
	
	// 执行除法
	结果 := 被放大 / other.value
	
	// 调整精度
	// d 的精度 + 8 - other 的精度 - 目标精度
	实际精度 := d.precision + 8 - other.precision
	if 实际精度 > DefaultPrecision {
		divisor := pow10(实际精度 - DefaultPrecision)
		结果 = (结果 + divisor/2) / divisor
	} else if 实际精度 < DefaultPrecision {
		结果 *= pow10(DefaultPrecision - 实际精度)
	}
	
	return New(结果, DefaultPrecision)
}

// DivInt 除以整数
func (d Decimal) DivInt(i int64) Decimal {
	if i == 0 {
		panic("division by zero")
	}
	return New(d.value/i, d.precision)
}

// Round 四舍五入到指定小数位
func (d Decimal) Round(places int) Decimal {
	if places >= d.precision {
		return d
	}
	
	diff := d.precision - places
	divisor := pow10(diff)
	remainder := d.value % divisor
	if remainder < 0 {
		remainder = -remainder
	}
	
	value := d.value / divisor
	if d.value < 0 {
		if remainder >= divisor/2 {
			value--
		}
	} else {
		if remainder >= divisor/2 {
			value++
		}
	}
	
	// 保持原始精度，但值已四舍五入
	multiplier := pow10(d.precision - places)
	return New(value*multiplier, d.precision)
}

// Floor 向下取整
func (d Decimal) Floor() Decimal {
	multiplier := pow10(d.precision)
	value := (d.value / multiplier) * multiplier
	return New(value, d.precision)
}

// Ceil 向上取整
func (d Decimal) Ceil() Decimal {
	multiplier := pow10(d.precision)
	if d.value%multiplier == 0 {
		return d
	}
	value := ((d.value / multiplier) + 1) * multiplier
	return New(value, d.precision)
}

// Truncate 截断到指定小数位
func (d Decimal) Truncate(places int) Decimal {
	if places >= d.precision {
		return d
	}
	
	diff := d.precision - places
	divisor := pow10(diff)
	value := (d.value / divisor) * divisor
	
	return New(value, d.precision)
}

// alignPrecision 对齐两个 Decimal 的精度
func alignPrecision(d1, d2 Decimal) (Decimal, Decimal) {
	if d1.precision == d2.precision {
		return d1, d2
	}
	
	maxPrecision := d1.precision
	if d2.precision > maxPrecision {
		maxPrecision = d2.precision
	}
	
	v1 := d1.value
	v2 := d2.value
	
	if d1.precision < maxPrecision {
		v1 *= pow10(maxPrecision - d1.precision)
	}
	if d2.precision < maxPrecision {
		v2 *= pow10(maxPrecision - d2.precision)
	}
	
	return New(v1, maxPrecision), New(v2, maxPrecision)
}

// pow10 计算10的n次方
func pow10(n int) int64 {
	result := int64(1)
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}

// Zero 零值
func Zero() Decimal {
	return New(0, DefaultPrecision)
}

// One 一值
func One() Decimal {
	return New(pow10(DefaultPrecision), DefaultPrecision)
}

// Max 返回较大的 Decimal
func Max(a, b Decimal) Decimal {
	if a.Gte(b) {
		return a
	}
	return b
}

// Min 返回较小的 Decimal
func Min(a, b Decimal) Decimal {
	if a.Lte(b) {
		return a
	}
	return b
}

// Sum 计算 Decimal 切片的和
func Sum(decimals ...Decimal) Decimal {
	result := Zero()
	for _, d := range decimals {
		result = result.Add(d)
	}
	return result
}

// Avg 计算 Decimal 切片的平均值
func Avg(decimals ...Decimal) Decimal {
	if len(decimals) == 0 {
		return Zero()
	}
	sum := Sum(decimals...)
	return sum.DivInt(int64(len(decimals)))
}