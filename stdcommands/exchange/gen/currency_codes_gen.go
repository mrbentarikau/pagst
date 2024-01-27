//go:build ignore
// +build ignore

// Generates currency codes file for API
// and prevents additional requests during command exec
// go get golang.org/x/tools/internal/gocommand@v0.1.11 etc...
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/mrbentarikau/pagst/stdcommands/util"
)

var cTime = time.Now().UTC()

const templateSource = `// GENERATED using gen/currency_codes_gen.go
{{- printf "\n// %s" (cTime.Format "02-01-2006 15:04:05") }}
// Symbols from https://api.frankfurter.app/currencies

package exchange

var Currencies = map[string]string{
	{{- range $k, $v := .}}
		"{{$k}}": "{{$v}}",
	{{- end}}
}
	
`

var (
	parsedTemplate = template.Must(template.New("").Funcs(template.FuncMap{"cTime": currentTime}).Parse(templateSource))
	flagOut        string
)

func init() {
	flag.StringVar(&flagOut, "o", "../currency_codes.go", "Output file")
	flag.Parse()
}

func CheckErr(errMsg string, err error) {
	if err != nil {
		fmt.Println(errMsg+":", err)
		os.Exit(1)
	}
}

func main() {
	body, err := util.RequestFromAPI("https://api.frankfurter.app/currencies")
	if err != nil {
		return
	}

	check := &Currencies{}
	readerToDecoder := bytes.NewReader(body)
	err = json.NewDecoder(readerToDecoder).Decode(&check)
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

type Currencies map[string]string

func currentTime() time.Time {
	return time.Now().UTC()
}
