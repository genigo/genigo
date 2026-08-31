package generator

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/genigo/genigo/internal/config"
	"github.com/genigo/genigo/internal/repo"
	"github.com/go-sql-driver/mysql"
)

// Golden file tests: run the generator against the docker fixtures
// (docker-compose.yml) and compare the emitted models with
// testdata/golden/<driver>/*.generated.go
//
// Env: TEST_MYSQL_DSN, TEST_POSTGRES_DSN  (skip when unset)
//      GOLDEN_UPDATE=1 rewrites the golden files
//
//	mysql:    root:genigo@tcp(127.0.0.1:3307)/genigo_test
//	postgres: postgres://genigo:genigo@127.0.0.1:5433/genigo_test?sslmode=disable

func TestGenerateGolden(t *testing.T) {
	cases := []struct {
		driver string
		envDSN string
	}{
		{"mysql", "TEST_MYSQL_DSN"},
		{"postgres", "TEST_POSTGRES_DSN"},
	}

	for _, tc := range cases {
		t.Run(tc.driver, func(t *testing.T) {
			dsn := os.Getenv(tc.envDSN)
			if dsn == "" {
				t.Skipf("%s is not set", tc.envDSN)
			}

			conf, err := confFromDSN(tc.driver, dsn)
			if err != nil {
				t.Fatalf("parse dsn: %+v", err)
			}
			conf.Dir = t.TempDir()
			conf.Pkg = "models"
			conf.Tags = []string{"db", "json"}
			conf.Replace = true

			config.Conf = *conf
			if err := repo.Connect(); err != nil {
				t.Fatalf("connect: %+v", err)
			}
			defer func() { repo.DB.Close(); repo.DB = nil }()

			if err := Generate(); err != nil {
				t.Fatalf("generate: %+v", err)
			}

			goldenDir := filepath.Join("..", "..", "testdata", "golden", tc.driver)
			if os.Getenv("GOLDEN_UPDATE") == "1" {
				if err := os.MkdirAll(goldenDir, 0o755); err != nil {
					t.Fatal(err)
				}
			}

			goldens, err := filepath.Glob(filepath.Join(goldenDir, "*.generated.go"))
			if err != nil {
				t.Fatal(err)
			}
			if len(goldens) == 0 {
				t.Fatalf("no golden files under %s (run with GOLDEN_UPDATE=1)", goldenDir)
			}

			for _, golden := range goldens {
				name := filepath.Base(golden)
				want, err := os.ReadFile(golden)
				if err != nil {
					t.Fatal(err)
				}
				got, err := os.ReadFile(filepath.Join(conf.Dir, name))
				if err != nil {
					t.Fatalf("generator did not emit %s: %+v", name, err)
				}
				if string(got) != string(want) {
					if os.Getenv("GOLDEN_UPDATE") == "1" {
						if err := os.WriteFile(golden, got, 0o644); err != nil {
							t.Fatal(err)
						}
						continue
					}
					t.Errorf("%s: generated output differs from golden file (run GOLDEN_UPDATE=1 to refresh):\n--- golden ---\n%s\n--- generated ---\n%s", name, want, got)
				}
			}
		})
	}
}

func confFromDSN(driver, dsn string) (*config.Config, error) {
	conf := &config.Config{}
	conf.DB.Driver = driver

	switch driver {
	case "postgres":
		u, err := url.Parse(dsn)
		if err != nil {
			return nil, err
		}
		host := u.Hostname()
		port := 5432
		conf.DB.Schema = u.Path
		if len(conf.DB.Schema) > 0 && conf.DB.Schema[0] == '/' {
			conf.DB.Schema = conf.DB.Schema[1:]
		}
		if u.User != nil {
			conf.DB.User = u.User.Username()
			conf.DB.Password, _ = u.User.Password()
		}
		if p := u.Port(); p != "" {
			port = atoi(p)
		}
		conf.DB.Host = host
		conf.DB.Port = port
	default:
		cfg, err := mysql.ParseDSN(dsn)
		if err != nil {
			return nil, err
		}
		conf.DB.User = cfg.User
		conf.DB.Password = cfg.Passwd
		conf.DB.Schema = cfg.DBName
		conf.DB.Host, conf.DB.Port = parseMysqlAddr(cfg.Addr)
	}
	return conf, nil
}

func parseMysqlAddr(addr string) (string, int) {
	// addr comes as host:port from mysql.ParseDSN
	host, port, ok := splitLastColon(addr)
	if !ok {
		return addr, 3306
	}
	return host, atoi(port)
}

func atoi(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return n
		}
		n = n*10 + int(s[i]-'0')
	}
	return n
}

func splitLastColon(s string) (string, string, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ':' {
			return s[:i], s[i+1:], true
		}
	}
	return s, "", false
}
