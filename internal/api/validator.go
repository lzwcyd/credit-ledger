package api

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ValidateRequired 校验必填字段
func ValidateRequired(fields map[string]string) error {
	var missing []string
	for name, value := range fields {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少必填字段: %s", strings.Join(missing, ", "))
	}
	return nil
}

// ValidateDateFormat 校验日期格式 YYYY-MM-DD
func ValidateDateFormat(date string) error {
	if date == "" {
		return fmt.Errorf("日期不能为空")
	}
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("日期格式错误: %s，应为 YYYY-MM-DD", date)
	}
	return nil
}

// ValidateLoanNo 校验借据编号格式（字母数字，6-32位）
func ValidateLoanNo(loanNo string) error {
	if loanNo == "" {
		return fmt.Errorf("借据编号不能为空")
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{6,32}$`, loanNo)
	if !matched {
		return fmt.Errorf("借据编号格式错误: %s，应为6-32位字母数字", loanNo)
	}
	return nil
}

// ValidatePositiveFloat 校验正数
func ValidatePositiveFloat(name string, value float64) error {
	if value <= 0 {
		return fmt.Errorf("%s 必须大于0", name)
	}
	return nil
}

// ValidateRange 校验范围
func ValidateRange(name string, value, min, max int) error {
	if value < min || value > max {
		return fmt.Errorf("%s 必须在 %d-%d 之间", name, min, max)
	}
	return nil
}
