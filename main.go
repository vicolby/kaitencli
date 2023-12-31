package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"

	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

type Contexts struct {
	Contexts []Context `json:"contexts"`
}

type Context struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
	Board  string `json:"board"`
}

type Space struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

type Board struct {
	Id          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Columns     []Column `json:"columns"`
	Lanes       []Lane   `json:"lanes"`
	Cards       []Card   `json:"cards"`
}

type Column struct {
	Id         int         `json:"id"`
	Title      string      `json:"title"`
	Subcolumns []Subcolumn `json:"subcolumns"`
	SortOrder  float32     `json:"sort_order"`
}

type Subcolumn struct {
	Id        int     `json:"id"`
	Title     string  `json:"title"`
	SortOrder float32 `json:"sort_order"`
}

type Lane struct {
	Id        int     `json:"id"`
	Title     string  `json:"title"`
	SortOrder float32 `json:"sort_order"`
}

type Card struct {
	Id        int      `json:"id"`
	Title     string   `json:"title"`
	Owner     Owner    `json:"owner"`
	Archived  bool     `json:"archived"`
	Lane_id   int      `json:"lane_id"`
	Column_id int      `json:"column_id"`
	Members   []Member `json:"members"`
}

type Member struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Type     int    `json:"type"`
}

type Owner struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
}

var contexts Contexts
var spaces []Space
var boards []Board
var board Board

func contains(account string, accounts []string) bool {
	for _, a := range accounts {
		if account == a {
			return true
		}
	}
	return false
}

func call(url string, method string, bearer string) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("Got error %s", err.Error())
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", bearer))

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Got error %s", err.Error())
	}
	return response, nil
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

	var apiKey string

	for _, context := range contexts.Contexts {
		accounts = append(accounts, context.Name)
		if context.Active == true {
			apiKey = context.Key
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)

	app := &cli.App{
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			{
				Name:  "contexts",
				Usage: "show accounts",
				Action: func(cCtx *cli.Context) error {
					for _, context := range contexts.Contexts {
						var accRow []string
						accRow = append(accRow, context.Name)
						accRow = append(accRow, strconv.FormatBool(context.Active))
						table.Append(accRow)
					}
					table.SetHeader([]string{"Name", "Active"})
					table.Render()
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
					{
						Name:  "set",
						Usage: "set active board id",
						Action: func(cCtx *cli.Context) error {
							board := cCtx.Args().First()
							for index, context := range contexts.Contexts {
								if context.Active == true {
									contexts.Contexts[index].Board = board
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

							fmt.Println("Board set successfully.")
							return nil
						},
					},
				},
			},
			{
				Name:  "get",
				Usage: "get info about entity",
				Action: func(cCtx *cli.Context) error {
					if cCtx.NArg() <= 0 {
						return fmt.Errorf("Command needs at least one argument\nUse -h flag to show available options")
					}
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:  "spaces",
						Usage: "get list of spaces",
						Action: func(cCtx *cli.Context) error {
							resp, err := call("https://rubbles-stories.kaiten.ru/api/latest/spaces", "GET", apiKey)
							if err != nil {
								return fmt.Errorf("Could not get list of spaces")
							}
							defer resp.Body.Close()

							err = json.NewDecoder(resp.Body).Decode(&spaces)
							if err != nil {
								log.Fatal(err)
							}

							for _, space := range spaces {
								var spaceRow []string
								spaceRow = append(spaceRow, strconv.Itoa(space.Id))
								spaceRow = append(spaceRow, space.Title)
								table.Append(spaceRow)
							}
							table.SetHeader([]string{"Id", "Title"})
							table.Render()
							return nil
						},
					},
					{
						Name: "boards",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "space",
								Usage:    "space id",
								Required: true,
							},
						},
						Usage: "get list of boards",
						Action: func(cCtx *cli.Context) error {
							resp, err := call(fmt.Sprintf("https://rubbles-stories.kaiten.ru/api/latest/spaces/%v/boards", cCtx.String("space")), "GET", apiKey)
							if err != nil {
								return fmt.Errorf("Could not get list of boards")
							}
							defer resp.Body.Close()

							err = json.NewDecoder(resp.Body).Decode(&boards)
							if err != nil {
								log.Fatal(err)
							}

							for _, board := range boards {
								var boardRow []string
								boardRow = append(boardRow, strconv.Itoa(board.Id))
								boardRow = append(boardRow, board.Title)
								boardRow = append(boardRow, board.Description)
								table.Append(boardRow)
							}
							table.SetHeader([]string{"Id", "Title", "Description"})
							table.Render()
							return nil
						},
					},
					{
						Name: "cards",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "board",
								Usage:    "board id",
								Required: false,
							},
						},
						Usage: "get list of board cards",
						Action: func(cCtx *cli.Context) error {
							var board_id string
							for _, context := range contexts.Contexts {
								if context.Active == true && context.Board != "" && cCtx.String("board") == "" {
									board_id = context.Board
								}
							}
							resp, err := call(fmt.Sprintf("https://rubbles-stories.kaiten.ru/api/latest/boards/%v", board_id), "GET", apiKey)
							if err != nil {
								return fmt.Errorf("Could not get list of boards for board_id %v", board_id)
							}
							defer resp.Body.Close()

							err = json.NewDecoder(resp.Body).Decode(&board)
							if err != nil {
								log.Fatal(err)
							}

							columns := make(map[int]string)
							lanes := make(map[int]string)

							for _, col := range board.Columns {
								columns[col.Id] = col.Title
								for _, subc := range col.Subcolumns {
									columns[subc.Id] = subc.Title
								}
							}

							for _, lane := range board.Lanes {
								lanes[lane.Id] = lane.Title
							}

							sort.Slice(board.Cards, func(i, j int) bool {
								return board.Cards[i].Lane_id < board.Cards[j].Lane_id
							})

							for _, card := range board.Cards {
								var responsible string
								for _, member := range card.Members {
									if member.Type == 2 {
										responsible = member.Username
									} else {
										responsible = ""
									}
								}
								var Row []string
								Row = append(Row, lanes[card.Lane_id])
								Row = append(Row, card.Title)
								Row = append(Row, card.Owner.Username)
								Row = append(Row, responsible)
								Row = append(Row, strconv.Itoa(card.Id))
								Row = append(Row, columns[card.Column_id])
								table.Append(Row)
							}

							table.SetHeader([]string{"Lane", "Title", "Owner", "Responsible", "Id", "Column"})
							table.Render()

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
