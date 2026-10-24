package config

import "testing"

func valid() Config {
	return Config{AppEnv: "dev", APIMode: "mock", JWTSecret: PlaceholderJWTSecret, MockPayment: true}
}

func TestDevDefaultsValid(t *testing.T) {
	if err := valid().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPlaceholderSecretRejectedOutsideDev(t *testing.T) {
	c := valid()
	c.AppEnv = "beta"
	if c.Validate() == nil {
		t.Fatal("expected error for placeholder secret in beta")
	}
}

func TestProductionRules(t *testing.T) {
	c := Config{AppEnv: "production", APIMode: "db", JWTSecret: "0123456789abcdef0123456789abcdef", MockPayment: true, MySQLDSN: "u:p@tcp(h)/db"}
	if c.Validate() == nil {
		t.Fatal("expected errors: mock payment + no TLS")
	}
	c.MockPayment = false
	c.MySQLDSN += "?tls=true"
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestShortSecret(t *testing.T) {
	c := valid()
	c.JWTSecret = "short"
	if c.Validate() == nil {
		t.Fatal("expected error")
	}
}
