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

func contains(account string, accounts []string) bool {
	for _, a := range accounts {
		if account == a {
			return true
		}
	}
	return false
}

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

	var accounts []string

	// var apiKey string

	for _, context := range contexts.Contexts {
		accounts = append(accounts, context.Name)
		// if context.Active == true {
		// 	apiKey = context.Key
		// }
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
				Subcommands: []*cli.Command{
					{
						Name:  "use",
						Usage: "set active account",
						Action: func(cCtx *cli.Context) error {
							account := cCtx.Args().First()
							if c := contains(account, accounts); c != true {
								fmt.Printf("Account %v not found \n\nExisting accounts: %s", account, accounts)
								return nil
							}
							for index, context := range contexts.Contexts {
								if context.Name == account {
									contexts.Contexts[index].Active = true
								} else {
									contexts.Contexts[index].Active = false
								}
							}

							newData, err := json.MarshalIndent(contexts, "", "    ")
							if err != nil {
								fmt.Println("Error marshaling JSON:", err)
								return nil
							}

							err = os.WriteFile("contexts.json", newData, os.ModePerm)
							if err != nil {
								fmt.Println("Error writing JSON file:", err)
								return nil
							}

							fmt.Println("Context updated successfully.")
							return nil
						},
					},
					{
						Name:  "add",
						Usage: "add account",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Usage:    "Name for new account",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "key",
								Usage:    "API Key for new account",
								Required: true,
							},
						},
						Action: func(cCtx *cli.Context) error {
							newAccount := Context{}

							name := cCtx.String("name")
							if c := contains(name, accounts); c == true {
								fmt.Printf("Account with name %s already exists", name)
								return nil
							}

							if len(name) > 0 {
								newAccount.Name = name
							} else {
								fmt.Println("Account name length mus be greater than zero")
								return nil
							}

							newAccount.Key = cCtx.String("key")
							newAccount.Active = false

							contexts.Contexts = append(contexts.Contexts, newAccount)
							newData, err := json.MarshalIndent(contexts, "", "    ")

							if err != nil {
								fmt.Println("Error marshaling JSON:", err)
								return nil
							}

							err = os.WriteFile("contexts.json", newData, os.ModePerm)
							if err != nil {
								fmt.Println("Error writing JSON file:", err)
								return nil
							}

							fmt.Println("Context added successfully.")
							return nil

						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
