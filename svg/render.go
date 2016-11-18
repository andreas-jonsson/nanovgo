// Copyright (C) 2016 Andreas T Jonsson

package svg

import "github.com/mode13/nanovgo"

func Render(ctx *nanovgo.Context, s *Svg) error {
	for _, g := range s.Groups {
		if err := renderGroup(ctx, &g); err != nil {
			return err
		}
	}
	return nil
}

func renderGroup(ctx *nanovgo.Context, g *Group) error {
	ctx.Save()

	for _, s := range g.Shapes {
		switch t := s.(type) {
		case *Path:
			if err := renderPath(ctx, t); err != nil {
				return err
			}
		case *Group:
			if err := renderGroup(ctx, t); err != nil {
				return err
			}
		default:
			// return fmt.Errorf("unknown shape: %T", s)
		}
	}

	ctx.Restore()
	return nil
}

func renderPath(ctx *nanovgo.Context, p *Path) error {
	var (
		cmdSinceMove, numDrawCmd int
	)

	ctx.BeginPath()

	ctx.SetStrokeWidth(1)
	ctx.SetStrokeColor(nanovgo.RGBA(255, 0, 0, 255))

	for i, seg := range p.Segments {

		//fmt.Printf("%T: %v\n", seg, seg)

		switch t := seg.(type) {
		case MoveTo:
			if cmdSinceMove > 0 || i == 0 {
				ctx.MoveTo(t.X, t.Y)
			} else {
				ctx.LineTo(t.X, t.Y)
				numDrawCmd++
			}
			cmdSinceMove = -1
		case LineTo:
			ctx.LineTo(t.X, t.Y)
			numDrawCmd++
		case CurveTo:
			ctx.BezierTo(t.C1x, t.C1y, t.C2x, t.C2y, t.X, t.Y)
			numDrawCmd++
		case ArcTo:
			//TODO This is likely to be wrong. /aj
			ctx.ArcTo(t.Rx, t.Ry, t.X, t.Y, t.Rx)
			numDrawCmd++
		case ClosePath:
			if numDrawCmd > 0 {
				ctx.ClosePath()
			}
		}

		cmdSinceMove++
	}

	if numDrawCmd > 0 {
		ctx.Stroke()
		//ctx.Fill()
	} else {
		ctx.Reset()
	}
	return nil
}
