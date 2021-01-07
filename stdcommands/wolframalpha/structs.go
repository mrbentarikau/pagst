package wolframalpha

import "encoding/xml"

type WolframAlpha struct {
	Queryresult struct {
		XMLName       xml.Name `xml:"queryresult"`
		Text          string   `xml:",chardata"`
		Success       string   `xml:"success,attr"`
		AttrError     string   `xml:"error,attr"`
		Numpods       string   `xml:"numpods,attr"`
		Datatypes     string   `xml:"datatypes,attr"`
		Timedout      string   `xml:"timedout,attr"`
		Timedoutpods  string   `xml:"timedoutpods,attr"`
		Timing        string   `xml:"timing,attr"`
		Parsetiming   string   `xml:"parsetiming,attr"`
		Parsetimedout string   `xml:"parsetimedout,attr"`
		Recalculate   string   `xml:"recalculate,attr"`
		ID            string   `xml:"id,attr"`
		Host          string   `xml:"host,attr"`
		Server        string   `xml:"server,attr"`
		Related       string   `xml:"related,attr"`
		Version       string   `xml:"version,attr"`
		Pod           []struct {
			Text       string `xml:",chardata"`
			Title      string `xml:"title,attr"`
			Scanner    string `xml:"scanner,attr"`
			ID         string `xml:"id,attr"`
			Position   string `xml:"position,attr"`
			Error      string `xml:"error,attr"`
			Numsubpods string `xml:"numsubpods,attr"`
			Primary    string `xml:"primary,attr"`
			Subpod     struct {
				Text         string `xml:",chardata"`
				Title        string `xml:"title,attr"`
				Plaintext    string `xml:"plaintext"`
				Microsources struct {
					Text        string `xml:",chardata"`
					Microsource string `xml:"microsource"`
				} `xml:"microsources"`
			} `xml:"subpod"`
			Expressiontypes struct {
				Text           string `xml:",chardata"`
				Count          string `xml:"count,attr"`
				Expressiontype struct {
					Text string `xml:",chardata"`
					Name string `xml:"name,attr"`
				} `xml:"expressiontype"`
			} `xml:"expressiontypes"`
			States struct {
				Text  string `xml:",chardata"`
				Count string `xml:"count,attr"`
				State []struct {
					Text  string `xml:",chardata"`
					Name  string `xml:"name,attr"`
					Input string `xml:"input,attr"`
				} `xml:"state"`
			} `xml:"states"`
			Infos struct {
				Text  string `xml:",chardata"`
				Count string `xml:"count,attr"`
				Info  []struct {
					Text  string `xml:",chardata"`
					Units struct {
						Text  string `xml:",chardata"`
						Count string `xml:"count,attr"`
						Unit  []struct {
							Text  string `xml:",chardata"`
							Short string `xml:"short,attr"`
							Long  string `xml:"long,attr"`
						} `xml:"unit"`
					} `xml:"units"`
					Link struct {
						Text     string `xml:",chardata"`
						URL      string `xml:"url,attr"`
						AttrText string `xml:"text,attr"`
					} `xml:"link"`
				} `xml:"info"`
			} `xml:"infos"`
		} `xml:"pod"`
		Assumptions struct {
			Text       string `xml:",chardata"`
			Count      string `xml:"count,attr"`
			Assumption struct {
				Text     string `xml:",chardata"`
				Type     string `xml:"type,attr"`
				Template string `xml:"template,attr"`
				Count    string `xml:"count,attr"`
				Value    []struct {
					Text  string `xml:",chardata"`
					Name  string `xml:"name,attr"`
					Desc  string `xml:"desc,attr"`
					Input string `xml:"input,attr"`
				} `xml:"value"`
			} `xml:"assumption"`
		} `xml:"assumptions"`
		Userinfoused struct {
			Text     string `xml:",chardata"`
			Count    string `xml:"count,attr"`
			Userinfo struct {
				Text string `xml:",chardata"`
				Name string `xml:"name,attr"`
			} `xml:"userinfo"`
		} `xml:"userinfoused"`
		Sources struct {
			Text   string `xml:",chardata"`
			Count  string `xml:"count,attr"`
			Source []struct {
				Text     string `xml:",chardata"`
				URL      string `xml:"url,attr"`
				AttrText string `xml:"text,attr"`
			} `xml:"source"`
		} `xml:"sources"`
		Error struct {
			Text string `xml:",chardata"`
			Code string `xml:"code"`
			Msg  string `xml:"msg"`
		} `xml:"error"`
	}
}
