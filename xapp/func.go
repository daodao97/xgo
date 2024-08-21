package xapp

func IsProd() bool {
	return Args.AppEnv == "prod"
}

func IsPre() bool {
	return Args.AppEnv == "pre"
}

func IsTest() bool {
	return Args.AppEnv == "test"
}

func IsLocal() bool {
	return Args.AppEnv == "local"
}
