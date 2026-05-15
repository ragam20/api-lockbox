package internal

func IsProduction(env string) bool {
	return env == "prod" || env == "production"
}
