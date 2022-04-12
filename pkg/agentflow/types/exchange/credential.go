package exchange

type Credentials struct {
	BasicAuth []BasicAuthCredential
	Redis     []RedisCredential
	Github    []GithubCredential
	MySQL     []MySQLCredential
}

type BasicAuthCredential struct {
	Name     string
	URL      string
	Username string
	Password string
}

type GithubCredential struct {
	Name     string
	APIToken string
}

type RedisCredential struct {
	Name string
	Auth string
}

type MySQLCredential struct {
	Name     string
	Username string
	Password string
}
