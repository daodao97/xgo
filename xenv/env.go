package xenv

import "os"

var AppEnv string

func init() {
	AppEnv = os.Getenv("APP_NEV")
	if AppEnv == "" {
		AppEnv = "dev"
	}
}

func IsDev() bool {
	return AppEnv == "dev"
}
