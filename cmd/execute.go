package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/genigo/genigo/internal/config"
	"github.com/genigo/genigo/internal/generator"
	"github.com/genigo/genigo/internal/repo"
	"github.com/genigo/genigo/internal/version"
)

func Execute() {
	conf := flag.String("c", "genigo.yaml", "Set config path")
	flag.Parse()

	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == "init" {
		driver := initDriver()
		initConfig(*conf, driver)
		fmt.Println("Config file created successfully!")
		fmt.Printf("Edit [%s] to set database connection and other options\n", *conf)
		return
	}

	if len(os.Args) > 1 && strings.Contains(os.Args[1], "ersion") {
		fmt.Printf("genigo Ver(%s)\n Developed by Mahmoud Eskandari.\n", version.Ver)
		return
	}

	err := config.ReadConfig(*conf)
	if err != nil {
		log.Fatalf("Error when reading config: %+v \n **** Run: `genigo init` to create a default config file ****\n", err)
	}

	err = repo.Connect()
	if err != nil {
		log.Fatalf("Error on connecting to database: %+v\n", err)
	}

	err = generator.Generate()
	if err != nil {
		log.Fatalf("Error on generate files: %+v\n", err)
	}

}

// initDriver reads the optional --driver flag of `genigo init`
// (`genigo init --driver postgres`), default is mysql
func initDriver() string {
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		arg := strings.TrimLeft(args[i], "-")
		if strings.HasPrefix(arg, "driver=") {
			return strings.TrimPrefix(arg, "driver=")
		}
		if arg == "driver" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return "mysql"
}
