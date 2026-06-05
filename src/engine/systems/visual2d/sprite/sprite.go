/******************************************************************************/
/* sprite.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package sprite

import (
	"log/slog"
	"weak"

	"kaijuengine.com/debug"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

const (
	spriteZAxisScaleFactor = 16.0
)

var (
	sFilter           = rendering.TextureFilterLinear
	sPixelPositioning = false
)

type Sprite struct {
	Entity          engine.Entity
	host            weak.Pointer[engine.Host]
	currentClipName string
	currentClip     SpriteSheetClip
	uvAnimation     []AnimatedUV
	frameDelay      float64
	fps             float32
	frameCount      int
	currentFrame    int
	spriteSheet     SpriteSheet
	drawings        []rendering.Drawing
	clipIdx         int
	updateId        engine.UpdateId
	baseScale       matrix.Vec3
	paused          bool
	autoScaleZ      bool
	enforcedBlended bool
	flipHorizontal  bool
}

func (s *Sprite) IsValid() bool { return len(s.drawings) > 0 }

func (s *Sprite) Init(x, y, width, height float32, host *engine.Host, texture string, color matrix.Color) {
	tex, err := host.TextureCache().Texture(texture, sFilter)
	if err != nil {
		slog.Error("failed to find the requested texture", "texture", texture)
		tex, _ = host.TextureCache().Texture(assets.TextureSquare, sFilter)
	}
	s.InitFromTexture(x, y, width, height, host, tex, color)
	s.Entity.OnDestroy.Add(func() {
		for i := range s.drawings {
			s.drawings[i].ShaderData.Destroy()
		}
	})
}

func (s *Sprite) InitFromTexture(x, y, width, height float32, host *engine.Host, texture *rendering.Texture, color matrix.Color) {
	s.baseInit(x, y, width, height, host)
	s.drawings = klib.SliceSetLen(s.drawings, 1)
	s.drawings[0], _ = s.buildDrawing(host, color, texture)
	host.Drawings.AddDrawing(s.drawings[0])
}

func (s *Sprite) InitUVAnimation(x, y, width, height float32, host *engine.Host, texture *rendering.Texture, color matrix.Color, inUVAnimation []AnimatedUV) {
	s.InitFromTexture(x, y, width, height, host, texture, color)
	s.uvAnimation = inUVAnimation
	s.updateId = host.Updater.AddUpdate(s.update)
	s.frameCount = len(s.uvAnimation)
}

func (s *Sprite) InitFlipBook(x, y, width, height float32, host *engine.Host, textures []*rendering.Texture, inFPS float32) {
	s.baseInit(x, y, width, height, host)
	s.drawings = klib.SliceSetLen(s.drawings, len(textures))
	for i := range len(textures) {
		s.drawings[i], _ = s.buildDrawing(host, matrix.ColorWhite(), textures[i])
		if i > 0 {
			s.drawings[i].ShaderData.Deactivate()
		} else {
			s.drawings[i].ShaderData.Activate()
		}
		host.Drawings.AddDrawing(s.drawings[i])
	}
	s.fps = inFPS
	s.frameCount = len(textures)
	s.updateId = host.UILateUpdater.AddUpdate(s.update)
}

func (s *Sprite) InitSheet(x, y, width, height float32, host *engine.Host, textureKey, sheetDataKey string, inFPS float32, initialClip string) error {
	data, err := host.AssetDatabase().Read(sheetDataKey)
	if err != nil {
		return err
	}
	texture, err := host.TextureCache().Texture(textureKey, sFilter)
	if err != nil {
		return err
	}
	// TODO:  Return an error if the sprite is not new
	s.InitFromTexture(x, y, width, height, host, texture, matrix.ColorWhite())
	s.spriteSheet, err = NewSheetFromBin(data)
	if err != nil {
		return err
	}
	s.fps = inFPS
	if !s.SetSheetClip(initialClip) {
		slog.Error("initial clip was not found", "clip", initialClip)
	}
	s.updateId = host.Updater.AddUpdate(s.update)
	return nil
}

func (s *Sprite) Position() matrix.Vec2 { return s.Entity.Transform.Position().AsVec2() }

func (s *Sprite) ContainsPoint(point matrix.Vec2) bool {
	return s.Entity.Transform.ContainsPoint(point.AsVec3())
}

func (s *Sprite) SetSize(width, height float32) {
	s.Entity.Transform.SetScale(matrix.NewVec3(width, height, 1))
}

func (s *Sprite) Move(x, y float32) {
	if matrix.Approx(x, 0) && matrix.Approx(y, 0) {
		return
	}
	p := s.Entity.Transform.Position()
	if sPixelPositioning {
		x = matrix.Round(x)
		y = matrix.Round(y)
	}
	s.Entity.Transform.SetPosition(matrix.NewVec3(p.X()+x, p.Y()+y, p.Z()))
}

func (s *Sprite) Move3D(x, y, z float32) {
	if matrix.Approx(x, 0) && matrix.Approx(y, 0) && matrix.Approx(z, 0) {
		return
	}
	p := s.Entity.Transform.Position()
	if sPixelPositioning {
		x = matrix.Round(x)
		y = matrix.Round(y)
	}
	s.Entity.Transform.SetPosition(matrix.NewVec3(p.X()+x, p.Y()+y, p.Z()+z))
}

func (s *Sprite) FlipHorizontal() {
	s.flipHorizontal = !s.flipHorizontal
	if s.isSpriteSheet() {
		s.setSheetFrame(s.currentFrame)
	} else {
		for i := range len(s.drawings) {
			sd := s.ShaderData(i)
			sd.UVs[matrix.Vx] += sd.UVs[matrix.Vz]
			sd.UVs[matrix.Vz] *= -1
		}
	}
}

func (s *Sprite) SetPosition(x, y float32) {
	p := s.Entity.Transform.Position()
	if sPixelPositioning {
		x = matrix.Round(x)
		y = matrix.Round(y)
	}
	s.Entity.Transform.SetPosition(matrix.NewVec3(x, y, p.Z()))
}

func (s *Sprite) SetPosition3D(x, y, z float32) {
	if sPixelPositioning {
		x = matrix.Round(x)
		y = matrix.Round(y)
	}
	s.Entity.Transform.SetPosition(matrix.NewVec3(x, y, z))
	if s.autoScaleZ {
		scale := s.baseScale.Add(matrix.Vec3One().Scale(z * spriteZAxisScaleFactor))
		scale.SetZ(1)
		s.Entity.Transform.SetScale(scale)
	}
}

func (s *Sprite) SetPositionZ(z float32) {
	p := s.Entity.Transform.Position()
	s.SetPosition3D(p.X(), p.Y(), z)
}

func (s *Sprite) SetFrame(frame int) {
	if s.isUVAnimated() {
		s.ShaderData(0).UVs = s.uvAnimation[frame].uv
	} else {
		s.drawings[s.currentFrame].ShaderData.Deactivate()
		s.drawings[frame].ShaderData.Activate()
	}
	s.currentFrame = frame
}

func (s *Sprite) SetTexture(texture *rendering.Texture) {
	if s.drawings[0].IsValid() {
		s.recreateDrawing(0, s.drawings[0].Material.HasTransparentSuffix())
		s.drawings[0].Material = s.drawings[0].Material.SelectRoot().CreateInstance([]*rendering.Texture{texture})
		host := s.host.Value()
		debug.EnsureNotNil(host)
		host.Drawings.AddDrawing(s.drawings[0])
	}
}

func (s *Sprite) SetFlipBookAnimation(inFPS float32, textures []*rendering.Texture) {
	count := len(textures)
	for i := range s.drawings {
		s.drawings[i].ShaderData.Destroy()
	}
	s.drawings = klib.SliceSetLen(s.drawings, count)
	host := s.host.Value()
	debug.EnsureNotNil(host)
	for i := range count {
		s.drawings[i], _ = s.buildDrawing(host, matrix.ColorWhite(), textures[i])
	}
	s.frameCount = count
	s.fps = inFPS
	s.currentFrame = 0
	s.resetDelay()
	s.SetFrame(0)
}

func (s *Sprite) SetColor(color matrix.Color) {
	for i := range s.drawings {
		s.drawings[i].ShaderData.(*shader_data_registry.ShaderDataUnlit).Color = color
	}
	s.setBlendingInternal(s.enforcedBlended || color.A() < 1)
}

func (s *Sprite) SetSheetClip(name string) bool {
	if s.currentClipName != name {
		if clip, ok := s.spriteSheet.Clips[name]; !ok {
			return false
		} else {
			s.currentClip = clip
		}
		s.currentClipName = name
		s.setSheetFrame(0)
		s.frameCount = len(s.currentClip.Frames)
	}
	return true
}

func (s *Sprite) Activate() {
	s.Entity.Activate()
	s.currentFrame = 0
	s.SetFrame(s.currentFrame)
}

func (s *Sprite) Deactivate() {
	for i := range s.drawings {
		s.drawings[i].ShaderData.Deactivate()
	}
	s.Entity.Deactivate()
}

func (s *Sprite) SetBlended() {
	s.enforcedBlended = true
	s.setBlendingInternal(true)
}

func (s *Sprite) SetUnBlended() {
	s.enforcedBlended = false
	s.setBlendingInternal(false)
}

func (s *Sprite) SwapMaterial(mat *rendering.Material) {
	host := s.host.Value()
	debug.EnsureNotNil(host)
	for i := range s.drawings {
		s.drawings[i].Material = mat
		s.recreateDrawing(i, mat.HasTransparentSuffix())
		host.Drawings.AddDrawing(s.drawings[i])
	}
}

func (s *Sprite) PlayAnimation()                             { s.paused = false }
func (s *Sprite) StopAnimation()                             { s.paused = true }
func (s *Sprite) SetFrameRate(inFPS float32)                 { s.fps = inFPS }
func SetDefaultTextureFilter(filter rendering.TextureFilter) { sFilter = filter }
func SetPixelPositioning(pixelPositioning bool)              { sPixelPositioning = pixelPositioning }

func (s *Sprite) SetUVs(drawing int, inUVs matrix.Vec4) {
	s.ShaderData(drawing).UVs = inUVs
}

func (s *Sprite) ShaderData(drawing int) *shader_data_registry.ShaderDataUnlit {
	return s.drawings[drawing].ShaderData.(*shader_data_registry.ShaderDataUnlit)
}

func (s *Sprite) isFlipBook() bool    { return len(s.drawings) > 1 }
func (s *Sprite) isSpriteSheet() bool { return len(s.spriteSheet.Clips) > 0 }
func (s *Sprite) isUVAnimated() bool  { return len(s.uvAnimation) > 0 }

func (s *Sprite) recreateDrawing(drawingIndex int, blended bool) {
	d := &s.drawings[drawingIndex]
	sd := s.ShaderData(drawingIndex)
	proxy := *sd
	sd.Destroy()
	sd = &shader_data_registry.ShaderDataUnlit{}
	*sd = proxy
	d.ShaderData = sd
	if d.Material.HasTransparentSuffix() != blended {
		host := s.host.Value()
		debug.EnsureNotNil(host)
		mat, err := host.MaterialCache().Material(assets.MaterialDefinitionUnlitTransparent)
		if err == nil {
			d.Material = mat.CreateInstance(d.Material.Textures)
		} else {
			slog.Error("failed to convert sprite to transparent material", "error", err)
		}
	}
}

func (s *Sprite) resetDelay() {
	if s.isUVAnimated() {
		s.frameDelay = float64(s.uvAnimation[s.currentFrame].seconds)
	} else {
		s.frameDelay = 1.0 / float64(s.fps)
	}
}

func (s *Sprite) setSheetFrame(frame int) {
	s.clipIdx = frame
	f := s.currentClip.Frames[frame]
	texture := s.drawings[0].Material.Textures[0]
	h := float32(f.Rectangle.Height()) / float32(texture.Height)
	for i := range s.drawings {
		x := float32(f.Rectangle.X()) / float32(texture.Width)
		z := float32(f.Rectangle.Width()) / float32(texture.Width)
		if s.flipHorizontal {
			x += z
			z *= -1
		}
		y := 1.0 - h - float32(f.Rectangle.Y())/float32(texture.Height)
		s.ShaderData(i).UVs = matrix.NewVec4(x, y, z, h)
	}
}

func (s *Sprite) update(deltaTime float64) {
	if !s.Entity.IsActive() || s.frameCount <= 1 {
		return
	}
	s.frameDelay -= deltaTime
	if s.frameDelay <= 0.0 {
		nextFrame := s.currentFrame + 1
		if nextFrame == s.frameCount {
			nextFrame = 0
		}
		if s.isUVAnimated() || s.isFlipBook() {
			s.SetFrame(nextFrame)
		} else if s.isSpriteSheet() {
			s.currentFrame = nextFrame
			frame := s.clipIdx + 1
			if frame < s.frameCount {
				s.setSheetFrame(frame)
			} else {
				s.setSheetFrame(0)
			}
		}
		s.resetDelay()
	}
}

func (s *Sprite) baseInit(x, y, width, height float32, host *engine.Host) {
	if sPixelPositioning {
		x = matrix.Round(x)
		y = matrix.Round(y)
	}
	// TODO:  Return an error if the sprite is not new
	*s = Sprite{
		host:      weak.Make(host),
		baseScale: matrix.NewVec3(width, height, 1.0),
	}
	s.Entity.Init(host.WorkGroup())
	s.Entity.Transform.SetPosition(matrix.NewVec3(x, y, 0))
	s.Entity.Transform.SetScale(matrix.NewVec3(width, height, 1))
}

func (s *Sprite) buildDrawing(host *engine.Host, color matrix.Color, texture *rendering.Texture) (rendering.Drawing, error) {
	matDef := assets.MaterialDefinitionUnlit
	if s.enforcedBlended || color.A() < 1 {
		matDef = assets.MaterialDefinitionUnlitTransparent
	}
	mat, err := host.MaterialCache().Material(matDef)
	if err != nil {
		slog.Error("failed to create the sprite material", "material", matDef, "error", err)
		return rendering.Drawing{}, err
	}
	mat = mat.CreateInstance([]*rendering.Texture{texture})
	mesh := rendering.NewMeshQuad(host.MeshCache())
	sd := shader_data_registry.Create(mat.Shader.DrawInstanceDataName()).(*shader_data_registry.ShaderDataUnlit)
	sd.Color = color
	d := rendering.Drawing{
		Material:   mat,
		Mesh:       mesh,
		ShaderData: sd,
		Transform:  &s.Entity.Transform,
		ViewCuller: &host.Cameras.Primary,
	}
	return d, err
}

func (s *Sprite) setBlendingInternal(blended bool) {
	host := s.host.Value()
	debug.EnsureNotNil(host)
	for i := range s.drawings {
		if s.drawings[i].Material.HasTransparentSuffix() != blended {
			s.recreateDrawing(i, blended)
			host.Drawings.AddDrawing(s.drawings[i])
		}
	}
}
