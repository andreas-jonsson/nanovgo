// Copyright (C) 2016 Andreas T Jonsson

package svg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/mode13/nanovgo"
)

type (
	CurveTo struct {
		IsAbsolute               bool
		C1x, C1y, C2x, C2y, X, Y float32
	}

	MoveTo struct {
		IsAbsolute bool
		X, Y       float32
	}

	ArcTo struct {
		IsAbsolute                    bool
		Rx, Ry, XAxisRotate           float32
		LargeArcFlag, SweepFlag, X, Y float32
	}

	LineTo    MoveTo
	ClosePath struct{}
)

type Path struct {
	Id        string
	Style     map[string]interface{}
	Segments  []interface{}
	Transform nanovgo.TransformMatrix
}

func (p *Path) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	p.Style = make(map[string]interface{})
	p.Transform = nanovgo.IdentityMatrix()

	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "id":
			p.Id = attr.Value
		case "style":
			if err := p.parseStyle(attr.Value); err != nil {
				return err
			}
		case "d":
			if err := p.parseSegments(attr.Value); err != nil {
				return err
			}
		case "transform":
			//g.TransformString = attr.Value
			//t, err := parseTransform(g.TransformString)

			//g.Transform = &t
		}
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}

		if _, ok := token.(xml.EndElement); ok {
			return nil
		}
	}
}

func (p *Path) parseStyle(value string) error {
	params := strings.Split(value, ";")
	for _, param := range params {
		kv := strings.Split(param, ":")
		if len(kv) != 2 {
			return fmt.Errorf("could not parse style: %s", strings.TrimSpace(param))
		}

		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])

		if intValue, err := strconv.ParseInt(v, 10, 32); err == nil {
			p.Style[k] = int(intValue)
		} else if floatValue, err := strconv.ParseFloat(v, 32); err == nil {
			p.Style[k] = float32(floatValue)
		} else {
			p.Style[k] = v
		}
	}

	return nil
}

func cleanSegmentString(s string) string {
	var buf bytes.Buffer
	for i, c := range s {
		if i > 0 && c == '-' && unicode.IsDigit(rune(s[i-1])) {
			fmt.Fprint(&buf, " ")
		}
		fmt.Fprintf(&buf, "%c", c)
	}
	return buf.String()
}

func (p *Path) parseSegments(value string) error {
	const controlCharacters = "aAcChHmMlLvVz"
	fmt.Println(value)
	value = cleanSegmentString(value)

	splitCommands := func(c rune) bool {
		for _, r := range controlCharacters {
			if r == c {
				return true
			}
		}
		return false
	}

	commandData := strings.FieldsFunc(value, splitCommands)
	commands := []rune{}

	for _, c := range value {
		for _, r := range controlCharacters {
			if r == c {
				commands = append(commands, c)
				break
			}
		}
	}

	toFloat := func(s string) float32 {
		f, err := strconv.ParseFloat(s, 32)
		if err != nil {
			log.Panicln(err)
		}
		return float32(f)
	}

	getArgs := func() []string {
		return strings.FieldsFunc(strings.TrimSpace(commandData[0]), func(c rune) bool {
			return unicode.IsSpace(c) || c == ','
		})
	}

	for _, cmd := range commands {
		isAbsolute := unicode.IsUpper(cmd)
		switch cmd {
		case 'a', 'A':
			args := getArgs()
			commandData = commandData[1:]

			for i := 0; i < len(args); i += 7 {
				p.Segments = append(p.Segments, ArcTo{
					isAbsolute,
					toFloat(args[i]),
					toFloat(args[i+1]),
					toFloat(args[i+2]),
					toFloat(args[i+3]),
					toFloat(args[i+4]),
					toFloat(args[i+5]),
					toFloat(args[i+6]),
				})
			}
		case 'm', 'M':
			args := getArgs()
			commandData = commandData[1:]

			for i := 0; i < len(args); i += 2 {
				p.Segments = append(p.Segments, MoveTo{isAbsolute, toFloat(args[i]), toFloat(args[i+1])})
			}
		case 'l', 'L', 'h', 'H', 'v', 'V':
			//TODO Add real support for H and V.

			args := getArgs()
			commandData = commandData[1:]

			for i := 0; i < len(args); i += 2 {
				p.Segments = append(p.Segments, LineTo{isAbsolute, toFloat(args[i]), toFloat(args[i+1])})
			}
		case 'c', 'C':
			args := getArgs()
			commandData = commandData[1:]

			if len(args)%6 == 0 {
				//Cubic Bezier
				for i := 0; i < len(args); i += 6 {
					p.Segments = append(p.Segments, CurveTo{
						isAbsolute,
						toFloat(args[i]),
						toFloat(args[i+1]),
						toFloat(args[i+2]),
						toFloat(args[i+3]),
						toFloat(args[i+4]),
						toFloat(args[i+5]),
					})
				}
			} else {
				//Quadratic Bezier
				//TODO Implement this. /aj
				log.Println("Quadratic bezier curves are not implemented yet.")
			}
		case 'z':
			p.Segments = append(p.Segments, ClosePath{})
		}
	}

	if len(commandData) > 0 {
		return fmt.Errorf("did not consume all data. (%d items left)\n%q", len(commandData), commandData)
	}

	return nil
}
