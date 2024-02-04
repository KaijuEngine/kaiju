package main

import (
	"fmt"
	"kaiju/assets"
	"kaiju/bootstrap"
	"kaiju/editor/project/ui/hierarchy"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/systems/console"
	"kaiju/ui"
	"kaiju/uimarkup"
	"kaiju/uimarkup/markup"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"
)

const TriangleShaderDataSize = int(unsafe.Sizeof(TestBasicShaderData{}))

type TestBasicShaderData struct {
	rendering.ShaderDataBase
	Color matrix.Color
}

func (t TestBasicShaderData) Size() int {
	const size = int(unsafe.Sizeof(TestBasicShaderData{}) - rendering.ShaderBaseDataStart)
	return size
}

func init() {
	runtime.LockOSThread()
}

func testDrawing(host *engine.Host) {
	shader := host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionBasic)
	mesh := rendering.NewMeshQuad(host.MeshCache())
	droidTex, _ := host.TextureCache().Texture("textures/android.png", rendering.TextureFilterNearest)
	tsd := TestBasicShaderData{rendering.NewShaderDataBase(), matrix.ColorWhite()}
	host.Drawings.AddDrawing(rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Shader:     shader,
		Mesh:       mesh,
		Textures:   []*rendering.Texture{droidTex},
		ShaderData: &tsd,
		Transform:  nil,
	})
}

func testTwoDrawings(host *engine.Host) {
	positions := []matrix.Vec3{
		{-1, 0.0, 0.0},
		{1, 0.0, 0.0},
	}
	colors := []matrix.Color{
		{1.0, 0.0, 0.0, 1.0},
		{0.0, 1.0, 0.0, 1.0},
	}
	rots := []matrix.Float{45, -45}
	for i := 0; i < 2; i++ {
		shader := host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionBasic)
		mesh := rendering.NewMeshQuad(host.MeshCache())
		droidTex, _ := host.TextureCache().Texture("textures/android.png", rendering.TextureFilterNearest)
		tsd := TestBasicShaderData{Color: colors[i]}
		m := matrix.Mat4Identity()
		m.Rotate(matrix.Vec3{0.0, rots[i], 0.0})
		m.Translate(positions[i])
		tsd.SetModel(m)
		host.Drawings.AddDrawing(rendering.Drawing{
			Renderer:   host.Window.Renderer,
			Shader:     shader,
			Mesh:       mesh,
			Textures:   []*rendering.Texture{droidTex},
			ShaderData: &tsd,
			Transform:  nil,
		})
	}
}

func testFont(host *engine.Host) {
	drawings := host.FontCache().RenderMeshes(host, "Hello, World!",
		0, float32(host.Window.Height())*0.5, 0, 64, float32(host.Window.Width()), matrix.ColorBlack(), matrix.ColorCornflowerBlue(),
		rendering.FontJustifyCenter, rendering.FontBaselineCenter,
		matrix.Vec3One(), true, false, []rendering.FontRange{},
		rendering.FontRegular)
	host.Drawings.AddDrawings(drawings)
}

func testOIT(host *engine.Host) {
	positions := []matrix.Vec3{
		{-0.75, 0.0, -0.75},
		{-0.5, 0.0, -0.5},
		{-0.25, 0.0, -0.25},
		{0.0, 0.0, 0.0},
	}
	colors := []matrix.Color{
		{1.0, 0.0, 1.0, 0.5},
		{1.0, 0.0, 0.0, 1.0},
		{0.0, 1.0, 0.0, 0.5},
		{0.0, 0.0, 1.0, 0.5},
	}
	shader := host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionBasic)
	mesh := rendering.NewMeshQuad(host.MeshCache())
	droidTex, _ := host.TextureCache().Texture("textures/square.png", rendering.TextureFilterNearest)
	for i := 0; i < len(positions); i++ {
		tsd := TestBasicShaderData{Color: colors[i]}
		m := matrix.Mat4Identity()
		m.Translate(positions[i])
		tsd.SetModel(m)
		host.Drawings.AddDrawing(rendering.Drawing{
			Renderer:    host.Window.Renderer,
			Shader:      shader,
			Mesh:        mesh,
			Textures:    []*rendering.Texture{droidTex},
			ShaderData:  &tsd,
			Transform:   nil,
			UseBlending: colors[i].A() < 1.0,
		})
		host.NewEntity().SetName(fmt.Sprintf("OIT %d", i))
	}
}

func testPanel(host *engine.Host) {
	tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	p := ui.NewPanel(host, tex, ui.AnchorBottomLeft)
	p.DontFitContent()
	p.Layout().Scale(100, 100)
	p.Layout().SetOffset(10, 10)
}

func testLabel(host *engine.Host) {
	l := ui.NewLabel(host, "Hello, World!", ui.AnchorBottomCenter)
	l.Layout().Scale(100, 50)
}

func testButton(host *engine.Host) {
	tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	btn := ui.NewButton(host, tex, "Click me!", ui.AnchorCenter)
	btn.Layout().Scale(100, 50)
	clickCount := 0
	btn.AddEvent(ui.EventTypeClick, func() {
		clickCount++
		btn.Label().SetText(fmt.Sprintf("Clicked x%d!", clickCount))
	})
}

func testHTML(host *engine.Host) {
	events := map[string]func(*markup.DocElement){
		"playGame":     func(*markup.DocElement) { println("Clicked playGame") },
		"showSettings": func(*markup.DocElement) { println("Clicked showSettings") },
		"showRules":    func(*markup.DocElement) { println("Clicked showRules") },
	}
	testHTML, _ := host.AssetDatabase().ReadText("ui/tests/test.html")
	testCSS, _ := host.AssetDatabase().ReadText("ui/tests/test.css")
	uimarkup.DocumentFromHTMLString(host, testHTML, testCSS, nil, events)
}

func testHTMLBinding(host *engine.Host) {
	demoData := struct {
		EntityNames []string
	}{
		EntityNames: []string{"Entity 1", "\tEntity 2", "\t\tEntity 3"},
	}

	testHTML, _ := host.AssetDatabase().ReadText("ui/tests/binding.html")
	uimarkup.DocumentFromHTMLString(host, testHTML, "", demoData, nil)
}

func testLayoutSimple(host *engine.Host) {
	tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)
	panels := []*ui.Panel{
		ui.NewPanel(host, tex, ui.AnchorBottomLeft),
		ui.NewPanel(host, tex, ui.AnchorBottomCenter),
		ui.NewPanel(host, tex, ui.AnchorBottomRight),
		ui.NewPanel(host, tex, ui.AnchorLeft),
		ui.NewPanel(host, tex, ui.AnchorRight),
		ui.NewPanel(host, tex, ui.AnchorCenter),
		ui.NewPanel(host, tex, ui.AnchorTopLeft),
		ui.NewPanel(host, tex, ui.AnchorTopCenter),
		ui.NewPanel(host, tex, ui.AnchorTopRight),
	}
	for _, p := range panels {
		p.DontFitContent()
		p.Layout().Scale(100, 100)
		p.Layout().SetOffset(10, 10)
	}
}

func testLayout(host *engine.Host) {
	tex, _ := host.TextureCache().Texture(assets.TextureSquare, rendering.TextureFilterLinear)

	p1 := ui.NewPanel(host, tex, ui.AnchorTopLeft)
	p1.Entity().SetName("p1")
	//p1.Layout().Scale(300, 100)

	p2 := ui.NewPanel(host, tex, ui.AnchorTopLeft)
	p2.Entity().SetName("p2")
	p2.SetColor(matrix.ColorBlue())
	//p2.Layout().SetPadding(5, 5, 5, 5)
	p2.Layout().SetMargin(5, 5, 5, 5)
	//p2.DontFitContent()
	//p2.Layout().Scale(64, 64)
	//p2.Layout().SetOffset(10, 10)

	p3 := ui.NewPanel(host, tex, ui.AnchorTopLeft)
	p3.Entity().SetName("p3")
	p3.SetColor(matrix.ColorRed())
	p3.Layout().Scale(32, 32)
	p3.Layout().SetOffset(10, 10)
	//p3.Layout().SetMargin(5, 5, 0, 0)

	p1.AddChild(p2)
	p2.AddChild(p3)
}

const (
	pprofCPU  = "cpu.prof"
	pprofHeap = "heap.prof"
)

func addConsole(host *engine.Host) {
	var pprofFile *os.File = nil
	console.For(host).AddCommand("EntityCount", func(string) string {
		return fmt.Sprintf("Entity count: %d", len(host.Entities()))
	})
	hrc := hierarchy.New()
	console.For(host).AddCommand("hrc", func(string) string {
		hrc.Destroy()
		hrc.Create(host)
		return ""
	})
	console.For(host).AddCommand("pprof", func(arg string) string {
		if arg == "start" {
			pprofFile = klib.MustReturn(os.Create(pprofCPU))
			pprof.StartCPUProfile(pprofFile)
			return "CPU profile started"
		} else if arg == "stop" {
			if pprofFile == nil {
				return "CPU profile not yet started"
			}
			pprof.StopCPUProfile()
			pprofFile.Close()
			return "CPU profile written to " + pprofCPU
		} else if arg == "heap" {
			hp := klib.MustReturn(os.Create(pprofHeap))
			pprof.WriteHeapProfile(hp)
			hp.Close()
			return "Heap profile written to " + pprofHeap
		} else {
			return ""
		}
	})
}

func main() {
	lastTime := time.Now()
	host, err := engine.NewHost()
	if err != nil {
		panic(err)
	}
	host.Window.Renderer.Initialize(&host, int32(host.Window.Width()), int32(host.Window.Height()))
	host.FontCache().Init(host.Window.Renderer, host.AssetDatabase(), &host)
	bootstrap.Main(&host)
	host.Camera.SetPosition(matrix.Vec3{0.0, 0.0, 2.0})
	//testDrawing(&host)
	//testTwoDrawings(&host)
	//testFont(&host)
	//testOIT(&host)
	//testPanel(&host)
	//testLabel(&host)
	//testButton(&host)
	//testHTML(&host)
	//[Kaiju Console]\nkl\nj\nj\nj\nj\nj\nj\nj\nj\nj\n\nj
	//testLayoutSimple(&host)
	//testLayout(&host)
	testHTMLBinding(&host)
	//addConsole(&host)
	for !host.Closing {
		since := time.Since(lastTime)
		deltaTime := since.Seconds()
		lastTime = time.Now()
		host.Update(deltaTime)
		host.Render()
	}
}
