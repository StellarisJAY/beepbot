package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrInvalidKey        = errors.New("invalid encryption key: must be 32 bytes")
	ErrInvalidCiphertext = errors.New("invalid ciphertext: too short")
	ErrDecryptionFailed  = errors.New("decryption failed: authentication failed")
)

// Encryptor 加密器
type Encryptor struct {
	key []byte
}

// NewEncryptor 创建加密器
// key 必须是 32 字节的 AES-256 密钥
func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	return &Encryptor{key: key}, nil
}

// NewEncryptorFromBase64 从 Base64 编码的密钥创建加密器
func NewEncryptorFromBase64(encodedKey string) (*Encryptor, error) {
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, err
	}
	return NewEncryptor(key)
}

// Encrypt 加密明文，返回 Base64 编码的密文
// 格式：base64(nonce + ciphertext + tag)
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 Base64 编码的密文
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// GenerateKey 生成新的 32 字节随机密钥
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateKeyBase64 生成 Base64 编码的密钥
func GenerateKeyBase64() (string, error) {
	key, err := GenerateKey()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// GetOrCreateEncryptionKey 获取或创建加密密钥
// 优先级：环境变量 > 配置文件 > 自动生成
func GetOrCreateEncryptionKey(configKey, keyFilePath string) ([]byte, error) {
	// 1. 尝试从环境变量获取
	if envKey := os.Getenv("BEEPBOT_ENCRYPTION_KEY"); envKey != "" {
		return base64.StdEncoding.DecodeString(envKey)
	}

	// 2. 尝试从配置获取
	if configKey != "" {
		return base64.StdEncoding.DecodeString(configKey)
	}

	// 3. 尝试从密钥文件读取
	if keyData, err := os.ReadFile(keyFilePath); err == nil {
		return base64.StdEncoding.DecodeString(string(keyData))
	}

	// 4. 生成新密钥并保存
	key, err := GenerateKey()
	if err != nil {
		return nil, err
	}

	// 确保目录存在
	dir := filepath.Dir(keyFilePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)
	if err := os.WriteFile(keyFilePath, []byte(encodedKey), 0600); err != nil {
		return nil, err
	}

	return key, nil
}

// MaskAPIKey 脱敏显示 API Key
func MaskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
