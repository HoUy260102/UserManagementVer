package configs

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Server struct {
	Port int `yaml:"port"`
}
type Database struct {
	URI  string `yaml:"uri"`
	Name string `yaml:"name"`
}
type Jwt struct {
	SecretKey                      string `yaml:"secret_key"`
	Issuer                         string `yaml:"issuer"`
	JwtAccessTokenExpirationTime   int    `yaml:"jwt_access_token_expiration_time"`
	JwtRefreshTokenExpirationTime  int    `yaml:"jwt_refresh_token_expiration_time"`
	JwtAprrovedTokenExpirationTime int    `yaml:"jwt_aprroved_token_expiration_time"`
}

type Email struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type Config struct {
	Server   Server   `yaml:"server"`
	Database Database `yaml:"database"`
	Jwt      Jwt      `yaml:"jwt"`
	Email    Email    `yaml:"email"`
}

var AppConfig *Config

func LoadFileConfig() {
	_ = godotenv.Load() // không cần truyền đường dẫn

	data, err := os.ReadFile("configs/config.yaml")
	if err != nil {
		log.Fatal("Lỗi đọc file", err)
	}
	//Thay biến môi trường trong file YAML bằng giá trị thực tế.
	expandedYaml := os.ExpandEnv(string(data))
	var cfg Config

	//Đọc YAML đã thay thế, ánh xạ vào struct Go.
	if err := yaml.Unmarshal([]byte(expandedYaml), &cfg); err != nil {
		fmt.Errorf("error parsing YAML: %w", err)
		return
	}

	AppConfig = &cfg
	fmt.Println("Config loaded successfully")
}
