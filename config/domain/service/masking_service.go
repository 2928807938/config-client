package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

// MaskingService 配置脱敏服务
// 负责处理敏感配置的识别、加密、解密和脱敏展示
type MaskingService struct {
	encryptionKey []byte // AES-256 加密密钥 (32字节)
	enabled       bool   // 是否启用脱敏功能
}

// NewMaskingService 创建脱敏服务实例
// encryptionKey: AES-256 加密密钥，必须是32字节长度
func NewMaskingService(encryptionKey string, enabled bool) *MaskingService {
	// 如果密钥长度不足32字节，填充到32字节
	key := []byte(encryptionKey)
	if len(key) < 32 {
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	} else if len(key) > 32 {
		key = key[:32]
	}

	return &MaskingService{
		encryptionKey: key,
		enabled:       enabled,
	}
}

// ==================== 敏感配置识别 ====================

// 敏感关键字列表（不区分大小写）
var sensitiveKeywords = []string{
	"password", "passwd", "pwd",
	"secret", "token", "apikey", "api_key",
	"private", "private_key", "privatekey",
	"auth", "authorization",
	"credential", "credentials",
	"security", "secure",
	"salt", "encryption_key",
}

// IsSensitiveKey 判断配置键是否为敏感配置
func (s *MaskingService) IsSensitiveKey(key string) bool {
	if !s.enabled {
		return false
	}

	lowerKey := strings.ToLower(key)

	// 检查是否包含敏感关键字
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lowerKey, keyword) {
			return true
		}
	}

	return false
}

// ==================== 脱敏展示 ====================

// MaskValue 对配置值进行脱敏展示
// 规则:
// - 短字符串(<=8): 全部显示为 ****
// - 中等长度(9-20): 保留前2后2, 中间为 ****
// - 长字符串(>20): 保留前4后4, 中间为 ****
func (s *MaskingService) MaskValue(value string) string {
	if !s.enabled {
		return value
	}

	length := len(value)

	if length == 0 {
		return ""
	}

	// 短字符串：全部脱敏
	if length <= 8 {
		return "****"
	}

	// 中等长度：保留前2后2
	if length <= 20 {
		return value[:2] + "****" + value[length-2:]
	}

	// 长字符串：保留前4后4
	return value[:4] + "****" + value[length-4:]
}

// ShouldMask 判断是否需要脱敏（根据配置键和值类型）
func (s *MaskingService) ShouldMask(key string, valueType string) bool {
	if !s.enabled {
		return false
	}

	// 加密类型的配置一定需要脱敏
	if valueType == "encrypted" {
		return true
	}

	// 检查配置键是否为敏感
	return s.IsSensitiveKey(key)
}

// ==================== 加密解密 ====================

// EncryptValue 加密配置值（使用AES-256-GCM）
func (s *MaskingService) EncryptValue(plaintext string) (string, error) {
	if !s.enabled {
		return plaintext, nil
	}

	if plaintext == "" {
		return "", nil
	}

	// 创建AES cipher
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密数据（nonce + ciphertext）
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Base64编码返回
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptValue 解密配置值
func (s *MaskingService) DecryptValue(ciphertext string) (string, error) {
	if !s.enabled {
		return ciphertext, nil
	}

	if ciphertext == "" {
		return "", nil
	}

	// Base64解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// 创建AES cipher
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	// 使用GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 获取nonce大小
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文数据长度不足")
	}

	// 分离nonce和密文
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// ==================== 工具方法 ====================

// IsEnabled 判断脱敏服务是否启用
func (s *MaskingService) IsEnabled() bool {
	return s.enabled
}

// GetSensitiveKeywords 获取敏感关键字列表（用于配置或调试）
func (s *MaskingService) GetSensitiveKeywords() []string {
	return sensitiveKeywords
}

// AddSensitiveKeyword 动态添加敏感关键字
func (s *MaskingService) AddSensitiveKeyword(keyword string) {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return
	}

	// 检查是否已存在
	for _, k := range sensitiveKeywords {
		if k == keyword {
			return
		}
	}

	sensitiveKeywords = append(sensitiveKeywords, keyword)
}
