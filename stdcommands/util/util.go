package util

import (
	"io"
	"net/http"
	"strconv"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
)

func RequestFromAPI(query string, extraHeaders ...map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", common.ConfBotUserAgent.GetString())

	// adding additional headers
	if len(extraHeaders) > 0 {
		for k, v := range extraHeaders[0] {
			req.Header.Set(k, v)
		}
	}

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

	return body, nil
}

func HumanizeThousands(input int64) string {
	var f1, f2 string

	i := int(input)
	if i < 0 {
		i = i * -1
		f2 = "-"
	}
	str := strconv.Itoa(i)

	idx := 0
	for i = len(str) - 1; i >= 0; i-- {
		idx++
		if idx == 4 {
			idx = 1
			f1 = f1 + ","
		}
		f1 = f1 + string(str[i])
	}

	for i = len(f1) - 1; i >= 0; i-- {
		f2 = f2 + string(f1[i])
	}
	return f2
}
