package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

type Contexts struct {
	Contexts []Context `json:"contexts"`
}

type Context struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

var contexts Contexts

func main() {

	jsonFile, err := os.Open("contexts.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
	}

	if err := json.Unmarshal(byteValue, &contexts); err != nil {
		log.Fatal(err)
	}

	app := &cli.App{
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			{
				Name:  "contexts",
				Usage: "show accounts",
				Action: func(cCtx *cli.Context) error {
					for _, context := range contexts.Contexts {
						fmt.Printf("Name: %s\nActive: %v\n\n", context.Name, context.Active)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
