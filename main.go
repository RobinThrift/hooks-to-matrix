package main

import (
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"os"
)

func action(c *cli.Context) error {
	configPath := c.String("config-path")

	if len(configPath) == 0 {
		return errors.New("Undefined config path")
	}

	config, err := LoadConfig(configPath)

	if err != nil {
		return err
	}

	ghConfigsMap := config.GetStringMap("github")
	ghConfigs := make([]*GitHubRepoConfig, 0, len(ghConfigsMap))

	for k, v := range ghConfigsMap {
		ghc, err := NewGitHubRepoConfig(k, v.(map[string]interface{}))

		if err != nil {
			return err
		}

		ghConfigs = append(ghConfigs, ghc)
	}

	fmt.Println("starting server...")
	StartServer(config.GetString("port"), ghConfigs)

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "hooks-to-matrix"
	app.Usage = "Starts a web hok server that will post to matrix"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config-path",
			Value: "./config.toml",
			Usage: "set the config file path",
		},
	}

	app.Action = action

	app.Run(os.Args)
}
