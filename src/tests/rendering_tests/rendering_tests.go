package tests

import (
	"fmt"
	"kaiju/assets"
	"kaiju/engine"
	"kaiju/host_container"
	"kaiju/klib"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/rendering"
	"kaiju/rendering/loaders"
	"kaiju/systems/console"
	"kaiju/ui"
	"strings"
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
		rendering.FontRegular, 0)
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
	events := map[string]func(*document.DocElement){
		"playGame":     func(*document.DocElement) { println("Clicked playGame") },
		"showSettings": func(*document.DocElement) { println("Clicked showSettings") },
		"showRules":    func(*document.DocElement) { println("Clicked showRules") },
	}
	testHTML, _ := host.AssetDatabase().ReadText("ui/tests/test.html")
	testCSS, _ := host.AssetDatabase().ReadText("ui/tests/test.css")
	markup.DocumentFromHTMLString(host, testHTML, testCSS, nil, events)
}

func testHTMLBinding(host *engine.Host) {
	demoData := struct {
		EntityNames []string
	}{
		EntityNames: []string{"Entity 1", "\tEntity 2", "\t\tEntity 3"},
	}
	testHTML, _ := host.AssetDatabase().ReadText("ui/tests/binding.html")
	markup.DocumentFromHTMLString(host, testHTML, "", demoData, nil)
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

func drawBasicMesh(host *engine.Host, res loaders.Result) {
	sd := TestBasicShaderData{rendering.NewShaderDataBase(), matrix.ColorWhite()}
	m := res[0]
	m.Textures = []string{assets.TextureSquare}
	textures := []*rendering.Texture{}
	for _, t := range m.Textures {
		tex, _ := host.TextureCache().Texture(t, rendering.TextureFilterLinear)
		textures = append(textures, tex)
	}
	mesh := rendering.NewMesh(m.Name, m.Verts, m.Indexes)
	host.MeshCache().AddMesh(mesh)
	host.Drawings.AddDrawing(rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Shader:     host.ShaderCache().ShaderFromDefinition(assets.ShaderDefinitionBasic),
		Mesh:       mesh,
		Textures:   textures,
		ShaderData: &sd,
	})
}

func testMonkeyOBJ(host *engine.Host) {
	const monkeyObj = "meshes/monkey.obj"
	host.Camera.SetPosition(matrix.Vec3{0, 0, 3})
	monkeyData := klib.MustReturn(host.AssetDatabase().ReadText(monkeyObj))
	res := loaders.OBJ(monkeyObj, monkeyData)
	if !res.IsValid() || len(res) != 1 {
		panic("Expected 1 mesh")
	}
	drawBasicMesh(host, res)
}

func testMonkeyGLTF(host *engine.Host) {
	const monkeyGLTF = "meshes/monkey.gltf"
	host.Camera.SetPosition(matrix.Vec3{0, 0, 3})
	res := klib.MustReturn(loaders.GLTF(host.Window.Renderer, monkeyGLTF, host.AssetDatabase()))
	if !res.IsValid() || len(res) != 1 {
		panic("Expected 1 mesh")
	}
	drawBasicMesh(host, res)
}

func testMonkeyGLB(host *engine.Host) {
	const monkeyGLTF = "meshes/monkey.glb"
	host.Camera.SetPosition(matrix.Vec3{0, 0, 3})
	res := klib.MustReturn(loaders.GLTF(host.Window.Renderer, monkeyGLTF, host.AssetDatabase()))
	if !res.IsValid() || len(res) != 1 {
		panic("Expected 1 mesh")
	}
	drawBasicMesh(host, res)
}

func SetupConsole(host *engine.Host) {
	console.For(host).AddCommand("test", func(_ *engine.Host, t string) string {
		var testFunc func(*engine.Host) = nil
		switch strings.ToLower(t) {
		case "drawing":
			testFunc = testDrawing
		case "two drawings":
			testFunc = testTwoDrawings
		case "font":
			testFunc = testFont
		case "oit":
			testFunc = testOIT
		case "panel":
			testFunc = testPanel
		case "label":
			testFunc = testLabel
		case "button":
			testFunc = testButton
		case "html":
			testFunc = testHTML
		case "layout simple":
			testFunc = testLayoutSimple
		case "layout":
			testFunc = testLayout
		case "html binding":
			testFunc = testHTMLBinding
		case "obj":
			testFunc = testMonkeyOBJ
		case "gltf":
			testFunc = testMonkeyGLTF
		case "glb":
			testFunc = testMonkeyGLB
		}
		if testFunc != nil {
			c := host_container.New("Test " + t)
			go c.Run(engine.DefaultWindowWidth, engine.DefaultWindowHeight)
			<-c.PrepLock
			c.Host.Camera.SetPosition(matrix.Vec3{0, 0, 2})
			testFunc(c.Host)
		}
		return "Running test"
	})
}
