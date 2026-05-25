/******************************************************************************/
/* editor_camera.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_controls

import (
	"math"

	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	rotScale                = 0.005
	zoomScale3DScroll       = float32(0.05)
	zoomScale2DScroll       = float32(1.0)
	zoomScale3D             = zoomScale3DScroll * 0.1
	zoomScale2D             = zoomScale2DScroll * 0.1
	flySpeedScrollIncrement = 0.1
	flySpeedModifierMin     = 0.1
	flySpeedModifierMax     = 10
)

type EditorCameraMode = int

const (
	EditorCameraModeNone = EditorCameraMode(iota)
	EditorCameraMode3d
	EditorCameraMode2d
	EditorCameraModeTop
	EditorCameraModeFront
	EditorCameraModeSide
	EditorCameraModeLeft
	EditorCameraModeRight
)

var cameraModeStrings = []string{"None", "3D", "2D", "Top", "Front", "Side", "Left", "Right"}

type EditorCameraViewport struct {
	Left    float32
	Top     float32
	Width   float32
	Height  float32
	Enabled bool
}

type EditorCamera struct {
	OnModeChange     events.EventWithArg[EditorCameraMode]
	Settings         *editor_settings.EditorCameraSettings
	camera           cameras.Camera
	viewport         EditorCameraViewport
	lastMousePos     matrix.Vec2
	flyStartMousePos matrix.Vec2
	mouseDown        matrix.Vec2
	lastHit          matrix.Vec3
	yawScale         matrix.Float
	dragging         bool
	mode             EditorCameraMode
	resizeId         events.Id
	flyCamStarted    bool
	flySpeedModifier float32
}

func (e *EditorCamera) Mode() EditorCameraMode { return e.mode }
func (e *EditorCamera) ModeString() string {
	if e.mode < 0 || e.mode >= len(cameraModeStrings) {
		return cameraModeStrings[EditorCameraModeNone]
	}
	return cameraModeStrings[e.mode]
}
func (e *EditorCamera) Camera() cameras.Camera { return e.camera }

func (e *EditorCamera) UseAsPrimary(host *engine.Host) {
	if host != nil && e.camera != nil {
		host.Cameras.Primary.ChangeCamera(e.camera)
	}
}

func (e *EditorCamera) SetViewportBounds(left, top, width, height float32) {
	if width <= 0 || height <= 0 {
		return
	}
	e.viewport = EditorCameraViewport{
		Left:    left,
		Top:     top,
		Width:   width,
		Height:  height,
		Enabled: true,
	}
}

func (e *EditorCamera) ClearViewportBounds() {
	e.viewport.Enabled = false
}

func (e *EditorCamera) LookAtPoint() matrix.Vec3 {
	defer tracing.NewRegion("EditorCamera.LookAtPoint").End()
	return e.camera.LookAt()
}

func (e *EditorCamera) viewportSize(host *engine.Host) (float32, float32) {
	if e.viewport.Enabled {
		return e.viewport.Width, e.viewport.Height
	}
	return float32(host.Window.Width()), float32(host.Window.Height())
}

func (e *EditorCamera) viewportCenter(host *engine.Host) (int, int) {
	if e.viewport.Enabled {
		return int(e.viewport.Left + e.viewport.Width*0.5),
			int(e.viewport.Top + e.viewport.Height*0.5)
	}
	return host.Window.Width() / 2, host.Window.Height() / 2
}

func (e *EditorCamera) screenInViewport(pos matrix.Vec2) bool {
	if !e.viewport.Enabled {
		return true
	}
	return pos.X() >= e.viewport.Left &&
		pos.X() <= e.viewport.Left+e.viewport.Width &&
		pos.Y() >= e.viewport.Top &&
		pos.Y() <= e.viewport.Top+e.viewport.Height
}

func (e *EditorCamera) mouseInViewport(host *engine.Host) bool {
	return e.screenInViewport(host.Window.Mouse.ScreenPosition())
}

func (e *EditorCamera) localScreenPosition(pos matrix.Vec2) matrix.Vec2 {
	if !e.viewport.Enabled {
		return pos
	}
	return matrix.NewVec2(pos.X()-e.viewport.Left, pos.Y()-e.viewport.Top)
}

func (e *EditorCamera) localPositionFromScreen(host *engine.Host, pos matrix.Vec2) matrix.Vec2 {
	if !e.viewport.Enabled {
		return matrix.NewVec2(pos.X(), float32(host.Window.Height())-pos.Y())
	}
	return matrix.NewVec2(pos.X()-e.viewport.Left, e.viewport.Height-(pos.Y()-e.viewport.Top))
}

func (e *EditorCamera) mousePosition(host *engine.Host) matrix.Vec2 {
	if !e.viewport.Enabled {
		return host.Window.Mouse.Position()
	}
	return e.localPositionFromScreen(host, host.Window.Mouse.ScreenPosition())
}

func (e *EditorCamera) mouseScreenPosition(host *engine.Host) matrix.Vec2 {
	return e.localScreenPosition(host.Window.Mouse.ScreenPosition())
}

func (e *EditorCamera) mousePositionForRay(mouse *hid.Mouse) matrix.Vec2 {
	if !e.viewport.Enabled {
		return mouse.Position()
	}
	return matrix.NewVec2(mouse.ScreenPosition().X()-e.viewport.Left,
		e.viewport.Height-(mouse.ScreenPosition().Y()-e.viewport.Top))
}

func (e *EditorCamera) mouseScreenPositionForRay(mouse *hid.Mouse) matrix.Vec2 {
	return e.localScreenPosition(mouse.ScreenPosition())
}

func (e *EditorCamera) SetMode(mode EditorCameraMode, host *engine.Host) {
	e.setMode(mode, host, true)
}

func (e *EditorCamera) SetModeForRenderView(mode EditorCameraMode, host *engine.Host) {
	e.setMode(mode, host, false)
}

func (e *EditorCamera) setMode(mode EditorCameraMode, host *engine.Host, bindPrimary bool) {
	defer tracing.NewRegion("EditorCamera.SetMode").End()
	if e.mode == mode && e.camera != nil {
		if bindPrimary {
			e.UseAsPrimary(host)
		}
		return
	}
	e.flySpeedModifier = 1
	e.mode = mode
	switch e.mode {
	case EditorCameraMode3d:
		w, h := e.viewportSize(host)
		cam := cameras.NewStandardCamera(w, h, w, h, matrix.Vec3Backward())
		tc := cameras.ToTurntable(cam)
		tc.SetYawPitchZoom(0, -25, 16)
		tc.SetLookAt(matrix.Vec3Zero())
		tc.SetZoom(15)
		e.camera = tc
	case EditorCameraMode2d:
		prev := e.camera
		if prev == nil && host != nil {
			prev = host.Cameras.Primary.Camera
		}
		cw := prev.Width()
		ch := prev.Height()
		vw, vh := e.viewportSize(host)
		ratio := cw / ch
		w := (cw / cw) * ratio * 10
		h := (ch / cw) * ratio * 10
		oc := cameras.NewStandardCameraOrthographic(w, h, vw, vh, matrix.NewVec3(0, 0, 100))
		e.camera = oc
		if host != nil && host.Window != nil {
			host.Window.OnResize.Remove(e.resizeId)
			e.resizeId = host.Window.OnResize.Add(e.OnWindowResize)
		}
	case EditorCameraModeTop, EditorCameraModeFront, EditorCameraModeSide,
		EditorCameraModeLeft, EditorCameraModeRight:
		vw, vh := e.viewportSize(host)
		e.camera = newFixedOrthographicStageCamera(e.mode, vw, vh)
	}
	if bindPrimary {
		e.UseAsPrimary(host)
	}
	e.OnModeChange.Execute(e.mode)
}

func (e *EditorCamera) OnWindowResize() {
	defer tracing.NewRegion("EditorCamera.OnWindowResize").End()
	klib.NotYetImplemented(309)
}

func (e *EditorCamera) Update(host *engine.Host, delta float64) (changed bool) {
	defer tracing.NewRegion("EditorCamera.Update").End()
	switch e.mode {
	case EditorCameraMode3d:
		win := host.Window
		m := &win.Mouse
		kb := &win.Keyboard
		if !kb.HasAlt() && e.mouseInViewport(host) && m.Pressed(hid.MouseButtonRight) {
			lockX, lockY := e.viewportCenter(host)
			e.flyStartMousePos = m.ScreenPosition()
			win.HideCursor()
			win.LockCursor(lockX, lockY)
			e.lastMousePos = e.localPositionFromScreen(host,
				matrix.NewVec2(float32(lockX), float32(lockY)))
			e.flyCamStarted = true
			return true
		} else if e.flyCamStarted && !kb.HasAlt() && m.Released(hid.MouseButtonRight) {
			e.flyCamStarted = false
			win.UnlockCursor()
			win.SetCursorPosition(int(e.flyStartMousePos.X()), int(e.flyStartMousePos.Y()))
			win.Mouse.SetPosition(e.flyStartMousePos.X(), e.flyStartMousePos.Y(),
				float32(win.Width()), float32(win.Height()))
			win.ShowCursor()
			return false
		} else if e.flyCamStarted && !kb.HasAlt() && m.Held(hid.MouseButtonRight) {
			e.update3dFly(host, delta)
			return true
		} else {
			return e.update3d(host, delta)
		}
	case EditorCameraMode2d:
		return e.update2d(host, delta)
	case EditorCameraModeTop, EditorCameraModeFront, EditorCameraModeSide,
		EditorCameraModeLeft, EditorCameraModeRight:
		return e.updateFixedOrthographic(host, delta)
	case EditorCameraModeNone:
		fallthrough
	default:
		return false
	}
}

func (e *EditorCamera) RayCast(mouse *hid.Mouse) graviton.Ray {
	defer tracing.NewRegion("EditorCamera.RayCast").End()
	if e.mode == EditorCameraMode2d {
		return e.camera.RayCast(e.mouseScreenPositionForRay(mouse))
	} else {
		return e.camera.RayCast(e.mousePositionForRay(mouse))
	}
}

func (e *EditorCamera) Focus(bounds graviton.AABB) {
	defer tracing.NewRegion("EditorCamera.Focus").End()
	z := bounds.Extent.Length()
	if z <= 0.01 {
		z = 5
	} else {
		z *= 5
	}
	if e.camera.IsOrthographic() {
		c := e.camera.(*cameras.StandardCamera)
		p := c.Position()
		p.SetX(bounds.Center.X())
		p.SetY(bounds.Center.Y())
		c.SetPositionAndLookAt(p, bounds.Center.Negative())
		r := c.Width() / c.Height()
		if c.Width() > c.Height() {
			c.Resize(z*r, z)
		} else {
			c.Resize(z, z*r)
		}
	} else {
		c := e.camera.(*cameras.TurntableCamera)
		c.SetLookAt(bounds.Center)
		c.SetZoom(z)
	}
}

func (e *EditorCamera) pan3d(tc *cameras.TurntableCamera, mp matrix.Vec2) {
	defer tracing.NewRegion("EditorCamera.pan3d").End()
	if hitPoint, ok := tc.ForwardPlaneHit(mp, tc.LookAt()); ok {
		if matrix.Vec3Approx(e.lastHit, matrix.Vec3Zero()) {
			e.lastHit = hitPoint
		}
		delta := e.lastHit.Subtract(hitPoint)
		if delta.Equals(matrix.Vec3Zero()) {
			return
		}
		tc.SetLookAt(tc.LookAt().Add(delta))
		e.lastHit, _ = tc.ForwardPlaneHit(mp, tc.LookAt())
	}
}

func (e *EditorCamera) pan2d(oc *cameras.StandardCamera, mp matrix.Vec2, host *engine.Host) {
	defer tracing.NewRegion("EditorCamera.pan2d").End()
	hitPoint := matrix.NewVec3(mp.X(), mp.Y(), 0)
	if matrix.Vec3Approx(e.lastHit, matrix.Vec3Zero()) {
		e.lastHit = hitPoint
	}
	vw, vh := e.viewportSize(host)
	cw := oc.Width() / vw
	ch := oc.Height() / vh
	delta := e.lastHit.Subtract(hitPoint).Multiply(matrix.NewVec3(cw, ch, 0))
	oc.SetPositionAndLookAt(oc.Position().Add(delta), oc.LookAt().Add(delta))
	e.lastHit = hitPoint.Add(delta)
}

func (e *EditorCamera) update3dFly(host *engine.Host, deltaTime float64) (changed bool) {
	defer tracing.NewRegion("EditorCamera.update3dFly").End()
	xSensitivity := e.Settings.FlyXSensitivity
	ySensitivity := e.Settings.FlyYSensitivity
	tc := e.camera.(*cameras.TurntableCamera)
	mouse := &host.Window.Mouse
	kb := &host.Window.Keyboard
	if mouse.Moved() {
		mp := e.mousePosition(host)
		md := e.lastMousePos.Subtract(mp)
		tc.FlyRotate(md.X()*xSensitivity, -md.Y()*ySensitivity)
	}
	cp := e.camera.Position()
	cl := e.camera.LookAt()
	var delta matrix.Vec3
	if mouse.Scrolled() {
		v := e.flySpeedModifier
		if mouse.ScrollY > 0 {
			v += flySpeedScrollIncrement
		} else {
			v -= flySpeedScrollIncrement
		}
		e.flySpeedModifier = klib.Clamp(v, flySpeedModifierMin, flySpeedModifierMax)
	}
	flySpeed := e.Settings.FlySpeed * e.flySpeedModifier
	if kb.KeyHeld(hid.KeyboardKeyW) {
		delta = e.camera.Forward().Scale(matrix.Float(deltaTime) * flySpeed)
		changed = true
	} else if kb.KeyHeld(hid.KeyboardKeyS) {
		delta = e.camera.Forward().Negative().Scale(matrix.Float(deltaTime) * flySpeed)
		changed = true
	}
	if kb.KeyHeld(hid.KeyboardKeyA) {
		delta.AddAssign(e.camera.Right().Negative().Scale(matrix.Float(deltaTime) * flySpeed))
		changed = true
	} else if kb.KeyHeld(hid.KeyboardKeyD) {
		delta.AddAssign(e.camera.Right().Scale(matrix.Float(deltaTime) * flySpeed))
		changed = true
	}
	if kb.KeyHeld(hid.KeyboardKeyQ) {
		delta.AddAssign(e.camera.Up().Negative().Scale(matrix.Float(deltaTime) * flySpeed))
		changed = true
	} else if kb.KeyHeld(hid.KeyboardKeyE) {
		delta.AddAssign(e.camera.Up().Scale(matrix.Float(deltaTime) * flySpeed))
		changed = true
	}
	if changed {
		e.camera.SetPositionAndLookAt(cp.Add(delta), cl.Add(delta))
	}
	return changed
}

func (e *EditorCamera) update3d(host *engine.Host, _ float64) (changed bool) {
	defer tracing.NewRegion("EditorCamera.update3d").End()
	tc := e.camera.(*cameras.TurntableCamera)
	mouse := &host.Window.Mouse
	kb := &host.Window.Keyboard
	mp := e.mousePosition(host)
	mouseInside := e.mouseInViewport(host)
	if mouseInside && (kb.HasAlt() || kb.KeyHeld(hid.KeyboardKeySpace)) {
		changed = true
	}
	if mouseInside && (mouse.Pressed(hid.MouseButtonLeft) || mouse.Pressed(hid.MouseButtonMiddle) ||
		(mouse.Pressed(hid.MouseButtonRight) && kb.HasAlt())) {
		e.dragging = true
		e.mouseDown = mp
		rg := int(math.Abs(float64(int(matrix.Rad2Deg(tc.Pitch())) % 360)))
		if rg < 90 || rg > 270 {
			e.yawScale = rotScale
		} else {
			e.yawScale = -rotScale
		}
		if mouse.Pressed(hid.MouseButtonMiddle) {
			changed = true
		}
	} else if e.dragging && mouse.Held(hid.MouseButtonLeft) {
		if kb.HasAlt() {
			x := (e.lastMousePos.Y() - mp.Y()) * -rotScale
			y := (e.lastMousePos.X() - mp.X()) * e.yawScale
			tc.Orbit(matrix.Vec3{x, y, 0.0})
			changed = true
		} else if kb.KeyHeld(hid.KeyboardKeySpace) {
			e.pan3d(tc, mp)
			changed = true
		}
	} else if e.dragging && mouse.Held(hid.MouseButtonMiddle) {
		e.pan3d(tc, mp)
		changed = true
	} else if e.dragging && mouse.Held(hid.MouseButtonRight) && kb.HasAlt() {
		dragDeltaY := e.lastMousePos.Y() - mp.Y()
		dragDeltaX := mp.X() - e.lastMousePos.X()
		dragDelta := dragDeltaY + dragDeltaX
		zoom := tc.Zoom()
		scale := zoomScale3D
		if zoom < 1.0 {
			scale *= zoom / 1.0
		}
		tc.Dolly(dragDelta * scale)
		changed = true
	} else if mouse.Released(hid.MouseButtonLeft) || mouse.Released(hid.MouseButtonMiddle) ||
		mouse.Released(hid.MouseButtonRight) {
		e.lastHit = matrix.Vec3Zero()
		if mouse.Released(hid.MouseButtonMiddle) {
			changed = true
		}
		e.dragging = false
	}
	if mouseInside && mouse.Scrolled() {
		zoom := tc.Zoom()
		scale := -zoomScale3DScroll
		if zoom < 1.0 {
			scale *= zoom / 1.0
		}
		zoomFloor := klib.ClampAbs(mouse.Scroll().Y(), e.Settings.ZoomSpeed)
		tc.Dolly(zoomFloor * scale)
		changed = true
	}
	e.lastMousePos = mp
	return changed
}

func (e *EditorCamera) update2d(host *engine.Host, _ float64) (changed bool) {
	defer tracing.NewRegion("EditorCamera.update2d").End()
	oc := e.camera.(*cameras.StandardCamera)
	mouse := &host.Window.Mouse
	kb := &host.Window.Keyboard
	mp := e.mousePosition(host)
	mouseInside := e.mouseInViewport(host)
	if mouseInside && (mouse.Pressed(hid.MouseButtonMiddle) ||
		(mouse.Pressed(hid.MouseButtonRight) && kb.HasAlt())) {
		e.dragging = true
		e.mouseDown = mp
		if mouse.Pressed(hid.MouseButtonMiddle) {
			changed = true
		}
	} else if e.dragging && mouse.Held(hid.MouseButtonMiddle) {
		e.pan2d(oc, mp, host)
		changed = true
	} else if e.dragging && mouse.Held(hid.MouseButtonRight) && kb.HasAlt() {
		cam := host.PrimaryCamera()
		cw := cam.Width()
		ch := cam.Height()
		w := oc.Width()
		h := oc.Height()
		r := cw / ch
		dragDeltaY := e.lastMousePos.Y() - mp.Y()
		dragDeltaX := mp.X() - e.lastMousePos.X()
		dragDelta := dragDeltaY + dragDeltaX
		w += (cw / cw) * r * -zoomScale2D * dragDelta
		h += (ch / cw) * r * -zoomScale2D * dragDelta
		if w > matrix.FloatSmallestNonzero && h > matrix.FloatSmallestNonzero {
			oc.Resize(w, h)
			changed = true
		}
	} else if mouse.Released(hid.MouseButtonMiddle) ||
		mouse.Released(hid.MouseButtonRight) {
		e.lastHit = matrix.Vec3Zero()
		if mouse.Released(hid.MouseButtonMiddle) {
			changed = true
		}
		e.dragging = false
	} else if mouseInside && kb.KeyHeld(hid.KeyboardKeySpace) {
		e.pan2d(oc, mp, host)
		changed = true
	}
	if mouseInside && mouse.Scrolled() {
		cam := host.PrimaryCamera()
		cw := cam.Width()
		ch := cam.Height()
		w := oc.Width()
		h := oc.Height()
		r := cw / ch
		zoomFloor := klib.ClampAbs(mouse.Scroll().Y(), e.Settings.ZoomSpeed)
		// Compute mouse world position before zoom (using consistent mapping from pan2d)
		winW, winH := e.viewportSize(host)
		sx := w / winW
		sy := h / winH
		centerX := oc.LookAt().X()
		centerY := oc.LookAt().Y()
		mouseWorldX := centerX + (mp.X()-winW*0.5)*sx
		mouseWorldY := centerY + (mp.Y()-winH*0.5)*sy
		// Compute new size (additive zoom, preserving aspect via r)
		dw := (cw / cw) * r * -zoomScale2DScroll * zoomFloor
		dh := (ch / cw) * r * -zoomScale2DScroll * zoomFloor
		newW := w + dw
		newH := h + dh
		if newW > matrix.FloatSmallestNonzero && newH > matrix.FloatSmallestNonzero {
			// Adjust center so mouse world point stays under cursor after resize
			newSx := newW / winW
			newSy := newH / winH
			newCenterX := mouseWorldX - (mp.X()-winW*0.5)*newSx
			newCenterY := mouseWorldY - (mp.Y()-winH*0.5)*newSy
			delta := matrix.NewVec3(newCenterX-centerX, newCenterY-centerY, 0)
			pos := oc.Position()
			look := oc.LookAt()
			oc.SetPositionAndLookAt(pos.Add(delta), look.Add(delta))
			oc.Resize(newW, newH)
			changed = true
		}
	}
	e.lastMousePos = mp
	return changed
}

func (e *EditorCamera) updateFixedOrthographic(host *engine.Host, _ float64) (changed bool) {
	defer tracing.NewRegion("EditorCamera.updateFixedOrthographic").End()
	oc := e.camera.(*cameras.StandardCamera)
	mouse := &host.Window.Mouse
	kb := &host.Window.Keyboard
	mp := e.mousePosition(host)
	mouseInside := e.mouseInViewport(host)
	if mouseInside && (mouse.Pressed(hid.MouseButtonMiddle) ||
		(mouse.Pressed(hid.MouseButtonRight) && kb.HasAlt())) {
		e.dragging = true
		e.mouseDown = mp
		if mouse.Pressed(hid.MouseButtonMiddle) {
			changed = true
		}
	} else if e.dragging && mouse.Held(hid.MouseButtonMiddle) {
		e.panFixedOrthographic(oc, mp, host)
		changed = true
	} else if mouse.Released(hid.MouseButtonMiddle) ||
		mouse.Released(hid.MouseButtonRight) {
		e.lastHit = matrix.Vec3Zero()
		if mouse.Released(hid.MouseButtonMiddle) {
			changed = true
		}
		e.dragging = false
	} else if mouseInside && kb.KeyHeld(hid.KeyboardKeySpace) {
		e.panFixedOrthographic(oc, mp, host)
		changed = true
	}
	if mouseInside && mouse.Scrolled() {
		zoomFloor := klib.ClampAbs(mouse.Scroll().Y(), e.zoomSpeed())
		r := oc.Width() / oc.Height()
		newW := oc.Width() + r*-zoomScale2DScroll*zoomFloor
		newH := oc.Height() + -zoomScale2DScroll*zoomFloor
		if newW > matrix.FloatSmallestNonzero && newH > matrix.FloatSmallestNonzero {
			oc.Resize(newW, newH)
			changed = true
		}
	}
	e.lastMousePos = mp
	return changed
}

func (e *EditorCamera) panFixedOrthographic(oc *cameras.StandardCamera, mp matrix.Vec2, host *engine.Host) {
	defer tracing.NewRegion("EditorCamera.panFixedOrthographic").End()
	if matrix.Vec3Approx(e.lastHit, matrix.Vec3Zero()) {
		e.lastHit = mp.AsVec3()
	}
	vw, vh := e.viewportSize(host)
	dx := (e.lastHit.X() - mp.X()) * oc.Width() / vw
	dy := (e.lastHit.Y() - mp.Y()) * oc.Height() / vh
	delta := oc.Right().Scale(dx).Add(oc.Up().Scale(dy))
	oc.SetPositionAndLookAt(oc.Position().Add(delta), oc.LookAt().Add(delta))
	e.lastHit = mp.AsVec3()
}

func (e *EditorCamera) zoomSpeed() float32 {
	if e.Settings == nil || e.Settings.ZoomSpeed <= 0 {
		return 120
	}
	return e.Settings.ZoomSpeed
}

func newFixedOrthographicStageCamera(mode EditorCameraMode, viewWidth, viewHeight float32) cameras.Camera {
	if viewWidth <= 0 {
		viewWidth = 1
	}
	if viewHeight <= 0 {
		viewHeight = 1
	}
	const size float32 = 20
	width := size * (viewWidth / viewHeight)
	height := size
	const distance matrix.Float = 50
	position := matrix.NewVec3(0, 0, distance)
	up := matrix.Vec3Up()
	switch mode {
	case EditorCameraModeTop:
		position = matrix.NewVec3(0, distance, 0)
		up = matrix.Vec3Forward()
	case EditorCameraModeFront:
		position = matrix.NewVec3(0, 0, distance)
		up = matrix.Vec3Up()
	case EditorCameraModeSide:
		position = matrix.NewVec3(distance, 0, 0)
		up = matrix.Vec3Up()
	case EditorCameraModeLeft:
		position = matrix.NewVec3(-distance, 0, 0)
		up = matrix.Vec3Up()
	case EditorCameraModeRight:
		position = matrix.NewVec3(distance, 0, 0)
		up = matrix.Vec3Up()
	}
	cam := cameras.NewStandardCameraOrthographic(width, height, viewWidth, viewHeight, position)
	cam.SetLookAtWithUp(matrix.Vec3Zero(), up)
	return cam
}
