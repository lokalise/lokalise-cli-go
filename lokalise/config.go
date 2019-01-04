package lokalise

import (
  "fmt"
	"os/user"

	"github.com/BurntSushi/toml"
)

type Config struct {
  Token   string
  Project string
}

func LoadConfig(configFile string) Config {
  var conf Config

  if configFile == "" {
    usr, err := user.Current()
    if err != nil {
        fmt.Println( err )
    }
    configFile = usr.HomeDir + "/.lokalise/lokalise.cfg"
  }

  if _, err := toml.DecodeFile(configFile, &conf); err != nil {
    // fallback to old config location
    if _, errOld := toml.DecodeFile("/etc/lokalise.cfg", &conf); errOld != nil {
      // do nothing if no config
    }
  }

  return conf
}
