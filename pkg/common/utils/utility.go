package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/turtacn/dataseap/pkg/common/constants"
)

// GenerateUUID 生成一个新的UUID字符串
// GenerateUUID generates a new UUID string.
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateRandomString 生成指定长度的随机字符串 (十六进制编码)
// GenerateRandomString generates a random string of specified length (hex encoded).
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}
	numBytes := (length + 1) / 2 // Each byte becomes two hex characters
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	hexStr := hex.EncodeToString(b)
	return hexStr[:length], nil // Trim if length is odd
}

// ToJSONString 将任意对象转换为JSON字符串
// ToJSONString converts any object to a JSON string.
func ToJSONString(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ToJSONStringPretty 将任意对象转换为格式化（美化）的JSON字符串
// ToJSONStringPretty converts any object to a formatted (pretty) JSON string.
func ToJSONStringPretty(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ") // 使用两个空格缩进 Use two spaces for indentation
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FromJSONString 将JSON字符串解析到给定的对象
// FromJSONString parses a JSON string into the given object.
func FromJSONString(jsonStr string, v interface{}) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// StringPtr 返回一个指向字符串值的指针
// StringPtr returns a pointer to a string value.
func StringPtr(s string) *string {
	return &s
}

// IntPtr 返回一个指向整数值的指针
// IntPtr returns a pointer to an integer value.
func IntPtr(i int) *int {
	return &i
}

// BoolPtr 返回一个指向布尔值的指针
// BoolPtr returns a pointer to a boolean value.
func BoolPtr(b bool) *bool {
	return &b
}

// TimeToString 将 time.Time 转换为标准格式字符串
// TimeToString converts time.Time to a standard format string.
func TimeToString(t time.Time) string {
	return t.Format(constants.DefaultTimeFormat)
}

// StringToTime 将标准格式字符串转换为 time.Time
// StringToTime converts a standard format string to time.Time.
func StringToTime(s string) (time.Time, error) {
	return time.Parse(constants.DefaultTimeFormat, s)
}

// IsValidEmail 验证邮箱格式是否有效
// IsValidEmail validates if the email format is valid.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// SanitizeString 清理字符串，移除潜在的有害字符或脚本（基础版）
// SanitizeString sanitizes a string, removing potentially harmful characters or scripts (basic version).
// 注意：这是一个非常基础的清理，对于安全关键的上下文，请使用更专业的库。
// Note: This is a very basic sanitizer. For security-critical contexts, use more specialized libraries.
func SanitizeString(s string) string {
	// 移除 HTML 标签 (简单示例)
	// Remove HTML tags (simple example)
	re := regexp.MustCompile(`<[^>]*>`)
	sanitized := re.ReplaceAllString(s, "")
	// 替换可能导致问题的字符
	// Replace characters that might cause issues
	sanitized = strings.ReplaceAll(sanitized, "'", "''") // SQL injection basic prevention for strings
	// 可以添加更多规则 Can add more rules
	return sanitized
}

// ContainsString 检查字符串切片是否包含特定字符串
// ContainsString checks if a slice of strings contains a specific string.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// UniqueStrings 从字符串切片中移除重复项
// UniqueStrings removes duplicate items from a slice of strings.
func UniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// GetEnvOrDefault 获取环境变量，如果未设置则返回默认值
// GetEnvOrDefault retrieves an environment variable, or returns a default value if not set.
func GetEnvOrDefault(key, defaultValue string) string {
	// This would typically use os.Getenv, but for broader utility without os import here:
	// value := os.Getenv(key)
	// if value == "" {
	// 	return defaultValue
	// }
	// return value
	// Placeholder as os interaction is better in config or main
	// For now, just return default for this utility function as an example.
	// In a real scenario, this function might not be needed if config handling (like viper) handles this.
	panic("GetEnvOrDefault should use os.Getenv, typically handled by a config package like Viper.")
	// return defaultValue
}
