// Copyright (C) 2016 Andreas T Jonsson

package svg

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"

	"github.com/mode13/nanovgo"
)

type Svg struct {
	Title     string  `xml:"title"`
	Groups    []Group `xml:"g"`
	Transform nanovgo.TransformMatrix
}

type Group struct {
	Id          string
	Stroke      string
	StrokeWidth int
	Fill        string
	FillRule    string
	Shapes      []interface{}
	Transform   nanovgo.TransformMatrix
}

func (g *Group) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	g.Transform = nanovgo.IdentityMatrix()

	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "id":
			g.Id = attr.Value
		case "stroke":
			g.Stroke = attr.Value
		case "stroke-width":
			if intValue, err := strconv.ParseInt(attr.Value, 10, 32); err != nil {
				return err
			} else {
				g.StrokeWidth = int(intValue)
			}
		case "fill":
			g.Fill = attr.Value
		case "fill-rule":
			g.FillRule = attr.Value
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

		switch tok := token.(type) {
		case xml.StartElement:
			var shape interface{}

			switch tok.Name.Local {
			case "g":
				shape = &Group{}
			case "path":
				shape = &Path{}
			default:
				return fmt.Errorf("unknown shape: %s", tok.Name.Local)
			}

			if err = decoder.DecodeElement(shape, &tok); err != nil {
				return fmt.Errorf("error decoding element of group: %v", err)
			} else {
				g.Shapes = append(g.Shapes, shape)
			}
		case xml.EndElement:
			return nil
		}
	}
}

func ParseSvg(reader io.Reader, scale float32) (*Svg, error) {
	var s Svg
	s.Transform = nanovgo.IdentityMatrix()
	if scale > 0 {
		nanovgo.ScaleMatrix(scale, scale)
	} else if scale < 0 {
		scale = 1.0 / -scale
		nanovgo.ScaleMatrix(scale, scale)

	}

	err := xml.NewDecoder(reader).Decode(&s)
	if err != nil {
		return &s, err
	}
	return &s, nil
}
