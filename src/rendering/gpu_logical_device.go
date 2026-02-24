package rendering

import (
	"kaiju/build"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"sort"
	"unsafe"
)

type GPULogicalDevice struct {
	GPUHandle
	graphicsQueue   unsafe.Pointer
	computeQueue    unsafe.Pointer
	presentQueue    unsafe.Pointer
	SwapChain       GPUSwapChain
	bufferTrash     bufferDestroyer
	dbg             *memoryDebugger
	renderPassCache map[string]*RenderPass
}

type GPUImageCreateRequest struct {
	Flags       GPUImageCreateFlags
	ImageType   GPUImageType
	Format      GPUFormat
	Extent      matrix.Vec3i
	MipLevels   uint32
	ArrayLayers uint32
	Samples     GPUSampleCountFlags
	Tiling      GPUImageTiling
	Usage       GPUImageUsageFlags
}

func (g *GPULogicalDevice) Setup(inst *GPUApplicationInstance, physicalDevice *GPUPhysicalDevice) error {
	defer tracing.NewRegion("GPULogicalDevice.Setup").End()
	g.renderPassCache = make(map[string]*RenderPass)
	return g.setupImpl(inst, physicalDevice)
}

func (g *GPULogicalDevice) WaitIdle() {
	defer tracing.NewRegion("GPULogicalDevice.WaitIdle").End()
	g.waitIdleImpl()
}

func (g *GPULogicalDevice) WaitForFences(fences []GPUFence) {
	defer tracing.NewRegion("GPULogicalDevice.WaitForFences").End()
	g.waitForFencesImpl(fences)
}

func (g *GPULogicalDevice) SetupDebug(inst *GPUApplicationInstance) {
	if build.Debug {
		defer tracing.NewRegion("GPULogicalDevice.SetupDebug").End()
		g.dbg = inst.dbg
		g.bufferTrash = newBufferDestroyer(g, g.dbg)
	}
}

func (g *GPULogicalDevice) ImageMemoryRequirements(image GPUImage) GPUMemoryRequirements {
	defer tracing.NewRegion("GPULogicalDevice.ImageMemoryRequirements").End()
	return g.imageMemoryRequirementsImpl(image)
}

func (g *GPULogicalDevice) CreateImageView(id *TextureId, aspectFlags GPUImageAspectFlags, viewType GPUImageViewType) error {
	defer tracing.NewRegion("GPULogicalDevice.CreateImageView").End()
	return g.createImageViewImpl(id, aspectFlags, viewType)
}

func (g *GPULogicalDevice) FreeTexture(texId *TextureId) {
	defer tracing.NewRegion("GPULogicalDevice.FreeTexture").End()
	g.freeTextureImpl(texId)
}

func (g *GPULogicalDevice) RemakeSwapChain(inst *GPUApplicationInstance) error {
	passes := make([]*RenderPass, 0, len(g.renderPassCache))
	for _, v := range g.renderPassCache {
		passes = append(passes, v)
	}
	// We need to sort the passes because some passes require resources from
	// others and need to be re-constructed afterwords
	sort.Slice(passes, func(i, j int) bool {
		return passes[i].construction.Sort < passes[j].construction.Sort
	})
	for i := range len(passes) {
		if err := passes[i].Recontstruct(inst); err != nil {
			return err
		}
	}
	return nil
}
