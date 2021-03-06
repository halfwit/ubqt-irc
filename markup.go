package main

import (
	"bytes"
	"fmt"

	"github.com/altid/libs/markup"
)

var colorCode = map[string]string{
	markup.White:      "0",
	markup.Black:      "1",
	markup.Blue:       "2",
	markup.Green:      "3",
	markup.Red:        "4",
	markup.Brown:      "5",
	markup.Purple:     "6",
	markup.Orange:     "7",
	markup.Yellow:     "8",
	markup.LightGreen: "9",
	markup.Cyan:       "10",
	markup.LightCyan:  "11",
	markup.LightBlue:  "12",
	markup.Pink:       "13",
	markup.Grey:       "14",
	markup.LightGrey:  "15",
}

func input(l *markup.Lexer) (*msg, error) {
	var m bytes.Buffer
	for {
		i := l.Next()
		switch i.ItemType {
		case markup.EOF:
			d := m.String()
			m := &msg{
				data: d,
				fn:   fself,
			}
			return m, nil
		case markup.ErrorText:
			return nil, fmt.Errorf("error parsing input: %v", i.Data)
		case markup.URLLink, markup.URLText, markup.ImagePath, markup.ImageLink, markup.ImageText:
			continue
		case markup.ColorText, markup.ColorTextBold:
			data, err := getColors(i.Data, l)
			if err != nil {
				return nil, err
			}

			m.WriteString(data)
		case markup.BoldText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case markup.EmphasisText:
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
		case markup.StrongText:
			m.WriteString("")
			m.WriteString("")
			m.Write(i.Data)
			m.WriteString("")
			m.WriteString("")
		default:
			m.Write(i.Data)
		}
	}
}

func getColors(current []byte, l *markup.Lexer) (string, error) {
	var text bytes.Buffer
	var color bytes.Buffer

	text.Write(current)

	for {
		i := l.Next()
		switch i.ItemType {
		case markup.ErrorText:
			return "", fmt.Errorf("%s", i.Data)
		case markup.EOF:
			return color.String(), nil
		case markup.ColorCode:
			code := colorCode[string(i.Data)]
			if n := bytes.IndexByte(i.Data, ','); n >= 0 {
				code = colorCode[string(i.Data[:n])]
				code += ","
				code += colorCode[string(i.Data[n+1:])]
			}
			color.WriteString("")
			color.WriteString(code)
			color.WriteString(text.String())
			color.WriteString("")
			return color.String(), nil
		case markup.ColorTextBold:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case markup.ColorTextEmphasis:
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
		case markup.ColorTextStrong:
			text.WriteString("")
			text.WriteString("")
			text.Write(i.Data)
			text.WriteString("")
			text.WriteString("")
		case markup.ColorText:
			text.Write(i.Data)
		}
	}
}
