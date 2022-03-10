package converter

type Configurator interface {
	RedactSecret(redaction string)
	ApplyDefaults() error
}
