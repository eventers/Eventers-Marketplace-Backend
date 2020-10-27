package config

import (
	"github.com/spf13/viper"
)

const (
	DBURL = "database.mysql"

	FirebaseProjectID             = "firebase.project_id"
	FirebaseServiceAccountKeyPath = "firebase.service_account_key_path"

	FromAddress            = "algorand.from_address"
	FromSecurityParaphrase = "algorand.from_security_paraphrase"
	ApiAddress             = "algorand.api_address"
	ApiKey                 = "algorand.api_key"
	AmountFactor           = "algorand.amount_factor"
	MinFee                 = "algorand.min_fee"
	SeedAlgo               = "algorand.seed_algo"

	VaultAddress   = "vault.address"
	VaultToken     = "vault.token"
	VaultUnSealKey = "vault.unseal_key"
	UserPath       = "vault.user_path"
	TempPath       = "vault.temp_path"

	Port               = "server.port"
	JWTOfflineInterval = "server.jwt_offline_interval"
	Secret             = "server.secret"

	RedisAddress  = "redis.address"
	RedisPassword = "redis.password"
	RedisDB       = "redis.db"

	TwilioAccountSID = "twilio.account_sid"
	TwilioAuthToken  = "twilio.auth_token"
	TwilioURL        = "twilio.url"
	TwilioFrom       = "twilio.from"
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault(Port, "9000")
	viper.SetDefault(JWTOfflineInterval, 120)
}
