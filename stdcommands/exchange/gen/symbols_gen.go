//go:build ignore
// +build ignore

// Generates symbols file for API
// and prevents additional requests during command exec
// go get golang.org/x/tools/internal/gocommand@v0.1.11 etc...
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"text/template"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
)

const templateSource = `// GENERATED using gen/symbols_gen.go

// Symbols from https://api.exchangerate.host/symbols

package exchange

var Symbols = map[string]map[string]string{
	{{- range $k, $v := .Symbols}}
		"{{$k}}": {
			"Description": "{{$v.Description}}",
			"Code": "{{$v.Code}}",
		},
	{{- end}}
}
	
`

var (
	parsedTemplate = template.Must(template.New("").Parse(templateSource))
	flagOut        string
)

func init() {
	flag.StringVar(&flagOut, "o", "../exchange_symbols.go", "Output file")
	flag.Parse()
}

func CheckErr(errMsg string, err error) {
	if err != nil {
		fmt.Println(errMsg+":", err)
		os.Exit(1)
	}
}

func main() {
	check, err := requestAPI("https://api.exchangerate.host/symbols")
	if err != nil {
		return
	}

	file, err := os.Create(flagOut)
	CheckErr("Failed creating output file", err)
	defer file.Close()
	err = parsedTemplate.Execute(file, check)
	CheckErr("Failed executing template", err)
	cmd := exec.Command("go", "fmt")
	err = cmd.Run()
	CheckErr("Failed running gofmt", err)
}

func requestAPI(query string) (*Result, error) {
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", common.ConfBotUserAgent.GetString())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, commands.NewPublicError("HTTP err: ", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := &Result{}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type SymbolInfo struct {
	Description string `json:"description,omitempty"`
	Code        string `json:"code,omitempty"`
}

type Result struct {
	Symbols map[string]SymbolInfo `json:"symbols,omitempty"`
}
