package initialize

import (
	"clicli/util/random"
	"github.com/spf13/viper"
)

func ConfigFiles() {
	viper.SetConfigName("application")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		panic("配置文件读取失败")
	}

	if viper.GetString("security.jwt_secret") == "" {
		viper.Set("server.jwt_secret", random.GenerateNumberCode(16))
	}

	viper.WriteConfig()
}
