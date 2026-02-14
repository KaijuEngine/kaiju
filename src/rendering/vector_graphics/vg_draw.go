package vector_graphics

import "kaiju/matrix"

// DrawingCommandType represents the type of drawing command
type DrawingCommandType uint8

const (
	DrawingCommandMoveTo DrawingCommandType = iota
	DrawingCommandLineTo
	DrawingCommandBezierCurveTo
	DrawingCommandQuadraticCurveTo
	DrawingCommandClosePath
)

const (
	// Kappa is the optimal constant for approximating a circle with cubic bezier curves
	// kappa = 4/3 * tan(pi/8) ≈ 0.5522847498307936
	Kappa matrix.Float = 0.5522847498307936
)

// DrawingContext represents the state for drawing operations
type DrawingContext struct {
	CurrentPoint    matrix.Vec2
	SubpathStart    matrix.Vec2
	HasCurrentPoint bool
	Commands        []DrawingCommand
}

// DrawingCommand represents a drawing operation
type DrawingCommand struct {
	Type   DrawingCommandType
	Points []matrix.Vec2
}

// NewDrawingContext creates a new drawing context
func NewDrawingContext() DrawingContext {
	return DrawingContext{}
}

// DrawLine draws a line from the current point to a new point
func (dc *DrawingContext) DrawLine(to matrix.Vec2) {
	if !dc.HasCurrentPoint {
		dc.CurrentPoint = to
		dc.HasCurrentPoint = true
		return
	}
	// Add line segment to current path
	dc.Commands = append(dc.Commands, DrawingCommand{
		Type:   DrawingCommandLineTo,
		Points: []matrix.Vec2{to},
	})
	dc.CurrentPoint = to
}

// DrawBezierCurve draws a cubic bezier curve from current point
func (dc *DrawingContext) DrawBezierCurve(control1, control2, end matrix.Vec2) {
	if !dc.HasCurrentPoint {
		dc.CurrentPoint = end
		dc.HasCurrentPoint = true
		return
	}
	// Add bezier curve to current path
	dc.Commands = append(dc.Commands, DrawingCommand{
		Type:   DrawingCommandBezierCurveTo,
		Points: []matrix.Vec2{control1, control2, end},
	})
	dc.CurrentPoint = end
}

// DrawQuadraticCurve draws a quadratic bezier curve from current point
func (dc *DrawingContext) DrawQuadraticCurve(control, end matrix.Vec2) {
	if !dc.HasCurrentPoint {
		dc.CurrentPoint = end
		dc.HasCurrentPoint = true
		return
	}
	// Add quadratic curve to current path
	dc.Commands = append(dc.Commands, DrawingCommand{
		Type:   DrawingCommandQuadraticCurveTo,
		Points: []matrix.Vec2{control, end},
	})
	dc.CurrentPoint = end
}

// MoveTo moves the pen to a new position without drawing
func (dc *DrawingContext) MoveTo(point matrix.Vec2) {
	dc.CurrentPoint = point
	dc.SubpathStart = point
	dc.HasCurrentPoint = true
	dc.Commands = append(dc.Commands, DrawingCommand{
		Type:   DrawingCommandMoveTo,
		Points: []matrix.Vec2{point},
	})
}

// ClosePath closes the current subpath by drawing a line back to the start
func (dc *DrawingContext) ClosePath() {
	if dc.HasCurrentPoint && dc.SubpathStart != dc.CurrentPoint {
		// Add closing line segment
		dc.Commands = append(dc.Commands, DrawingCommand{
			Type:   DrawingCommandClosePath,
			Points: []matrix.Vec2{},
		})
		dc.CurrentPoint = dc.SubpathStart
	}
}

// DrawRect draws a rectangle
func (dc *DrawingContext) DrawRect(x, y, width, height matrix.Float) {
	dc.MoveTo(matrix.NewVec2(x, y))
	dc.DrawLine(matrix.NewVec2(x+width, y))
	dc.DrawLine(matrix.NewVec2(x+width, y+height))
	dc.DrawLine(matrix.NewVec2(x, y+height))
	dc.ClosePath()
}

// DrawCircle draws a circle using bezier curves
func (dc *DrawingContext) DrawCircle(center matrix.Vec2, radius matrix.Float) {
	kappa := Kappa * radius // Optimal kappa for circle approximation
	dc.MoveTo(matrix.NewVec2(center.X(), center.Y()-radius))
	dc.DrawBezierCurve(
		matrix.NewVec2(center.X()+kappa*radius, center.Y()-radius),
		matrix.NewVec2(center.X()+radius, center.Y()-kappa*radius),
		matrix.NewVec2(center.X()+radius, center.Y()),
	)
	dc.DrawBezierCurve(
		matrix.NewVec2(center.X()+radius, center.Y()+kappa*radius),
		matrix.NewVec2(center.X()+kappa*radius, center.Y()+radius),
		matrix.NewVec2(center.X(), center.Y()+radius),
	)
	dc.DrawBezierCurve(
		matrix.NewVec2(center.X()-kappa*radius, center.Y()+radius),
		matrix.NewVec2(center.X()-radius, center.Y()+kappa*radius),
		matrix.NewVec2(center.X()-radius, center.Y()),
	)
	dc.DrawBezierCurve(
		matrix.NewVec2(center.X()-radius, center.Y()-kappa*radius),
		matrix.NewVec2(center.X()-kappa*radius, center.Y()-radius),
		matrix.NewVec2(center.X(), center.Y()-radius),
	)
	dc.ClosePath()
}

// DrawEllipse draws an ellipse using bezier curves
func (dc *DrawingContext) DrawEllipse(center matrix.Vec2, radiusX, radiusY matrix.Float) {
	kappaX := Kappa * radiusX
	kappaY := Kappa * radiusY
	dc.MoveTo(matrix.NewVec2(center.X(), center.Y()-radiusY))
	dc.DrawBezierCurve(
		matrix.NewVec2(center.X()+kappaX, center.Y()-radiusY),
		matrix.NewVec2(center.X()+radiusX, center.Y()-kappaY),
		matrix.NewVec2(center.X()+radiusX, center.Y()),
	)
	dc.DrawBezierCurve(
		matrix.NewVec2(center.X()+radiusX, center.Y()+kappaY),
		matrix.NewVec2(center.X()+kappaX, center.Y()+radiusY),
		matrix.NewVec2(center.X(), center.Y()+radiusY),
	)
	dc.DrawBezierCurve(
		matrix.NewVec2(center.X()-kappaX, center.Y()+radiusY),
		matrix.NewVec2(center.X()-radiusX, center.Y()+kappaY),
		matrix.NewVec2(center.X()-radiusX, center.Y()),
	)
	dc.DrawBezierCurve(
		matrix.NewVec2(center.X()-radiusX, center.Y()-kappaY),
		matrix.NewVec2(center.X()-kappaX, center.Y()-radiusY),
		matrix.NewVec2(center.X(), center.Y()-radiusY),
	)
	dc.ClosePath()
}

// DrawArc draws an arc from the current point
// This is a simplified implementation that approximates arcs with bezier curves
func (dc *DrawingContext) DrawArc(center matrix.Vec2, radius, startAngle, endAngle matrix.Float, clockwise bool) {
	// For a full implementation, we would use trigonometry to calculate arc points
	// and approximate them with bezier curves
	// For now, this serves as a placeholder for the editor API

	// The actual implementation would:
	// 1. Calculate the angle difference
	// 2. Split into segments (max 90 degrees per segment for good approximation)
	// 3. Calculate control points using the kappa method
	// 4. Draw each segment as a bezier curve
	_ = center
	_ = radius
	_ = startAngle
	_ = endAngle
	_ = clockwise
}

// DrawPath draws a complete path
func (dc *DrawingContext) DrawPath(path Path) {
	if len(path.Points) == 0 {
		return
	}
	dc.MoveTo(path.Points[0])
	for _, line := range path.Lines {
		dc.DrawLine(line.End)
	}
	for _, quad := range path.Quadratics {
		dc.DrawQuadraticCurve(quad.Control, quad.End)
	}
	for _, curve := range path.Curves {
		dc.DrawBezierCurve(curve.Control1, curve.Control2, curve.End)
	}
	if path.IsClosed {
		dc.ClosePath()
	}
}

// DrawShape draws a complete shape with its path and styling
func (dc *DrawingContext) DrawShape(shape Shape) {
	dc.DrawPath(shape.Path)
	// Apply fill and stroke styling
}

// CreateRectShape creates a rectangle shape
func CreateRectShape(x, y, width, height matrix.Float) Shape {
	var dc DrawingContext
	dc.DrawRect(x, y, width, height)
	return Shape{
		Path:        dc.ToPath(),
		FillColor:   matrix.NewColor(1, 1, 1, 1),
		StrokeColor: matrix.NewColor(0, 0, 0, 1),
		StrokeWidth: 1,
	}
}

// CreateCircleShape creates a circle shape
func CreateCircleShape(center matrix.Vec2, radius matrix.Float) Shape {
	var dc DrawingContext
	dc.DrawCircle(center, radius)
	return Shape{
		Path:        dc.ToPath(),
		FillColor:   matrix.NewColor(1, 1, 1, 1),
		StrokeColor: matrix.NewColor(0, 0, 0, 1),
		StrokeWidth: 1,
	}
}

// CreateEllipseShape creates an ellipse shape
func CreateEllipseShape(center matrix.Vec2, radiusX, radiusY matrix.Float) Shape {
	var dc DrawingContext
	dc.DrawEllipse(center, radiusX, radiusY)
	return Shape{
		Path:        dc.ToPath(),
		FillColor:   matrix.NewColor(1, 1, 1, 1),
		StrokeColor: matrix.NewColor(0, 0, 0, 1),
		StrokeWidth: 1,
	}
}

// ToPath converts the drawing context to a path
func (dc *DrawingContext) ToPath() Path {
	path := Path{
		Points: make([]matrix.Vec2, 0, len(dc.Commands)),
	}
	var currentPoint matrix.Vec2
	for _, cmd := range dc.Commands {
		switch cmd.Type {
		case DrawingCommandMoveTo:
			currentPoint = cmd.Points[0]
			path.Points = append(path.Points, currentPoint)
		case DrawingCommandLineTo:
			if len(cmd.Points) > 0 {
				path.Lines = append(path.Lines, Line{
					Start: currentPoint,
					End:   cmd.Points[0],
				})
				currentPoint = cmd.Points[0]
				path.Points = append(path.Points, currentPoint)
			}
		case DrawingCommandBezierCurveTo:
			if len(cmd.Points) >= 3 {
				path.Curves = append(path.Curves, BezierCurve{
					Start:    currentPoint,
					Control1: cmd.Points[0],
					Control2: cmd.Points[1],
					End:      cmd.Points[2],
				})
				currentPoint = cmd.Points[2]
				path.Points = append(path.Points, currentPoint)
			}
		case DrawingCommandQuadraticCurveTo:
			if len(cmd.Points) >= 2 {
				path.Quadratics = append(path.Quadratics, QuadraticBezier{
					Start:   currentPoint,
					Control: cmd.Points[0],
					End:     cmd.Points[1],
				})
				currentPoint = cmd.Points[1]
				path.Points = append(path.Points, currentPoint)
			}
		case DrawingCommandClosePath:
			// ClosePath doesn't add a new point, just marks the path as closed
			// The actual closing line is handled by setting IsClosed
		}
	}
	// Check if the path should be closed based on the last command
	if len(dc.Commands) > 0 && dc.Commands[len(dc.Commands)-1].Type == DrawingCommandClosePath {
		path.IsClosed = true
	}
	return path
}

// DrawFilledRect draws a filled rectangle
func (dc *DrawingContext) DrawFilledRect(x, y, width, height matrix.Float, color matrix.Color) {
	dc.DrawRect(x, y, width, height)
	// Apply fill color
}

// DrawStrokedRect draws a stroked rectangle
func (dc *DrawingContext) DrawStrokedRect(x, y, width, height matrix.Float, strokeColor matrix.Color, strokeWidth matrix.Float) {
	dc.DrawRect(x, y, width, height)
	// Apply stroke color and width
}

// CreatePathFromPoints creates a path from a series of points
func CreatePathFromPoints(points []matrix.Vec2, isClosed bool) Path {
	if len(points) < 2 {
		return Path{}
	}
	path := Path{
		Points:   points,
		IsClosed: isClosed,
	}
	// Create line segments between points
	for i := 0; i < len(points)-1; i++ {
		path.Lines = append(path.Lines, Line{
			Start: points[i],
			End:   points[i+1],
		})
	}
	if isClosed && len(points) > 2 {
		path.Lines = append(path.Lines, Line{
			Start: points[len(points)-1],
			End:   points[0],
		})
	}
	return path
}

// CreateBezierPath creates a path with bezier curves
func CreateBezierPath(start matrix.Vec2, curves []BezierCurve, isClosed bool) Path {
	path := Path{
		Points:   []matrix.Vec2{start},
		Curves:   curves,
		IsClosed: isClosed,
	}
	// Collect all points from curves
	currentPoint := start
	path.Points = append(path.Points, currentPoint)
	for _, curve := range curves {
		currentPoint = curve.End
		path.Points = append(path.Points, currentPoint)
	}
	return path
}

// CreateQuadraticPath creates a path with quadratic bezier curves
func CreateQuadraticPath(start matrix.Vec2, curves []QuadraticBezier, isClosed bool) Path {
	path := Path{
		Points:     []matrix.Vec2{start},
		Quadratics: curves,
		IsClosed:   isClosed,
	}
	// Collect all points from curves
	currentPoint := start
	path.Points = append(path.Points, currentPoint)
	for _, curve := range curves {
		currentPoint = curve.End
		path.Points = append(path.Points, currentPoint)
	}
	return path
}
