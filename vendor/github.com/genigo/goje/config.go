package goje

import (
	"fmt"
	"net/url"
	"time"
)

/*
Goje database config schema
# yaml example
driver: mysql
host: 127.0.0.1
port: 3306
user: root
password:
schema: mydbname
*/
type DBConfig struct {
	Driver   string            `json:"driver" yaml:"driver"`
	Host     string            `json:"host" yaml:"host"`
	Port     int               `json:"port" yaml:"port"`
	User     string            `json:"user" yaml:"user"`
	Password string            `json:"password" yaml:"password"`
	Schema   string            `json:"schema" yaml:"schema"`
	Flags    map[string]string `json:"flags" yaml:"flags"`

	MaxIdleTime     time.Duration `json:"MaxIdleTime" yaml:"MaxIdleTime"`
	MaxOpenConns    int           `json:"MaxOpenConns" yaml:"MaxOpenConns"`
	MaxIdleConns    int           `json:"MaxIdleConns" yaml:"MaxIdleConns"`
	ConnMaxLifetime time.Duration `json:"ConnMaxLifetime" yaml:"ConnMaxLifetime"`
}

func (db DBConfig) String() string {
	// Create the base URL structure
	u := &url.URL{
		Scheme: "",
		User:   url.UserPassword(db.User, db.Password),
		Host:   fmt.Sprintf("tcp(%s:%d)", db.Host, db.Port),
		Path:   "/" + db.Schema,
	}

	// Create query parameters
	q := url.Values{}
	q.Set("parseTime", "True")

	// Add all custom flags from the db.Flags map
	for k, v := range db.Flags {
		q.Set(k, v)
	}

	// Set the query string on the URL
	u.RawQuery = q.Encode()

	return u.String()[2:]
}
