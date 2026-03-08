package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Agent  AgentConfig
	LLM    LLMConfig
	Memory MemoryConfig
}

type AgentConfig struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
	MaxSteps    int    `mapstructure:"max_steps"`
}

// LLMConfig 完整LLM配置
type LLMConfig struct {
	// 基础连接
	Backend string `mapstructure:"backend"`  // ollama, openai, claude, azure, deepseek等
	BaseURL string `mapstructure:"base_url"` // API端点
	APIKey  string `mapstructure:"api_key"`  // 认证密钥（支持环境变量）

	// 模型选择
	Model          string `mapstructure:"model"`           // 模型ID
	ModelVersion   string `mapstructure:"model_version"`   // 特定版本（Azure需要）
	OrganizationID string `mapstructure:"organization_id"` // OpenAI组织ID（可选）

	// 生成参数
	Temperature      float64 `mapstructure:"temperature"`       // 0.0-2.0，默认0.7
	MaxTokens        int     `mapstructure:"max_tokens"`        // 最大生成token数
	TopP             float64 `mapstructure:"top_p"`             // 核采样，默认1.0
	FrequencyPenalty float64 `mapstructure:"frequency_penalty"` // 重复惩罚，默认0
	PresencePenalty  float64 `mapstructure:"presence_penalty"`  // 主题新鲜度，默认0

	// 高级设置
	Timeout        int  `mapstructure:"timeout_seconds"` // 请求超时，默认60
	RetryAttempts  int  `mapstructure:"retry_attempts"`  // 失败重试次数，默认3
	Stream         bool `mapstructure:"stream"`          // 默认启用流式，默认false
	EnableThinking bool `mapstructure:"enable_thinking"` // 是否显示思维链（Claude/QwQ）
}

// IsCloudBackend 判断是否为云端后端
func (l *LLMConfig) IsCloudBackend() bool {
	cloudBackends := map[string]bool{
		"openai":   true,
		"claude":   true,
		"azure":    true,
		"deepseek": true,
	}
	return cloudBackends[l.Backend]
}

type MemoryConfig struct {
	Type string `mapstructure:"type"` // memory类型，如"redis", "in-memory", "file"

	Maxsize        int    `mapstructure:"max_size"`
	Parsistent     bool   `mapstructure:"persistent"`      // 是否持久化存储
	ParsistentPath string `mapstructure:"persistent_path"` // 持久化存储路径（仅当persistent为true时有效）
	VectorDBURL    string `mapstructure:"vector_db_url"`   // 向量数据库URL（如Pinecone、Weaviate等）
	EmbeddingModel string `mapstructure:"embedding_model"` // 用于生成向量的模型ID
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/mika-agent/")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file error: %w", err)
		}
	}

	v.SetEnvPrefix("MIKA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.LLM.Model == "" {
		return fmt.Errorf("llm.model is required")
	}

	// 云端API检查Key
	if cfg.LLM.IsCloudBackend() && cfg.LLM.APIKey == "" {
		return fmt.Errorf("llm.api_key is required for backend: %s", cfg.LLM.Backend)
	}

	if cfg.LLM.Temperature < 0 || cfg.LLM.Temperature > 2 {
		return fmt.Errorf("llm.temperature must be between 0 and 2")
	}

	return nil
}
