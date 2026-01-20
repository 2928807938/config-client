package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Server   ServerConfig   `yaml:"server"`
	Log      LogConfig      `yaml:"log"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `yaml:"driver"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Database        string `yaml:"database"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // 秒
	LogLevel        string `yaml:"log_level"`
	AutoMigrate     bool   `yaml:"auto_migrate"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
	DialTimeout  int    `yaml:"dial_timeout"`  // 秒
	ReadTimeout  int    `yaml:"read_timeout"`  // 秒
	WriteTimeout int    `yaml:"write_timeout"` // 秒
	PoolTimeout  int    `yaml:"pool_timeout"`  // 秒
	IdleTimeout  int    `yaml:"idle_timeout"`  // 秒
	MaxConnAge   int    `yaml:"max_conn_age"`  // 秒
}

// GetAddr 获取Redis地址
func (r *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// GetDialTimeout 获取拨号超时
func (r *RedisConfig) GetDialTimeout() time.Duration {
	return time.Duration(r.DialTimeout) * time.Second
}

// GetReadTimeout 获取读超时
func (r *RedisConfig) GetReadTimeout() time.Duration {
	return time.Duration(r.ReadTimeout) * time.Second
}

// GetWriteTimeout 获取写超时
func (r *RedisConfig) GetWriteTimeout() time.Duration {
	return time.Duration(r.WriteTimeout) * time.Second
}

// GetPoolTimeout 获取连接池超时
func (r *RedisConfig) GetPoolTimeout() time.Duration {
	return time.Duration(r.PoolTimeout) * time.Second
}

// GetIdleTimeout 获取空闲超时
func (r *RedisConfig) GetIdleTimeout() time.Duration {
	return time.Duration(r.IdleTimeout) * time.Second
}

// GetMaxConnAge 获取连接最大存活时间
func (r *RedisConfig) GetMaxConnAge() time.Duration {
	if r.MaxConnAge == 0 {
		return 0
	}
	return time.Duration(r.MaxConnAge) * time.Second
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// GetDSN 获取数据库DSN连接字符串
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		d.Host, d.Port, d.Username, d.Password, d.Database,
	)
}

// GetConnMaxLifetime 获取连接最大生命周期
func (d *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(d.ConnMaxLifetime) * time.Second
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	setDefaults(&config)

	return &config, nil
}

// setDefaults 设置默认值
func setDefaults(config *Config) {
	// 数据库默认值
	if config.Database.Driver == "" {
		config.Database.Driver = "postgres"
	}
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 100
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 10
	}
	if config.Database.ConnMaxLifetime == 0 {
		config.Database.ConnMaxLifetime = 3600
	}
	if config.Database.LogLevel == "" {
		config.Database.LogLevel = "info"
	}

	// Redis默认值
	if config.Redis.Host == "" {
		config.Redis.Host = "localhost"
	}
	if config.Redis.Port == 0 {
		config.Redis.Port = 6379
	}
	if config.Redis.PoolSize == 0 {
		config.Redis.PoolSize = 100
	}
	if config.Redis.MinIdleConns == 0 {
		config.Redis.MinIdleConns = 10
	}
	if config.Redis.DialTimeout == 0 {
		config.Redis.DialTimeout = 5
	}
	if config.Redis.ReadTimeout == 0 {
		config.Redis.ReadTimeout = 3
	}
	if config.Redis.WriteTimeout == 0 {
		config.Redis.WriteTimeout = 3
	}
	if config.Redis.PoolTimeout == 0 {
		config.Redis.PoolTimeout = 4
	}
	if config.Redis.IdleTimeout == 0 {
		config.Redis.IdleTimeout = 300
	}

	// 服务器默认值
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Mode == "" {
		config.Server.Mode = "debug"
	}

	// 日志默认值
	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	if config.Log.Format == "" {
		config.Log.Format = "text"
	}
	if config.Log.Output == "" {
		config.Log.Output = "stdout"
	}
}
