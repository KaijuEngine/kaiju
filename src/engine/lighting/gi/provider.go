/******************************************************************************/
/* provider.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gi

import (
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const (
	ProviderNull       = "null"
	ProviderBakedProbe = "baked-probes"
	ProviderDDGI       = "ddgi"
)

type ViewID uint64
type InstanceID uint64
type MaterialID uint64
type LightID uint64

type Capabilities struct {
	VulkanMajor            uint32
	VulkanMinor            uint32
	DeviceLocalMemoryMB    uint64
	DedicatedComputeQueue  bool
	Synchronization2       bool
	TimelineSemaphore      bool
	DescriptorIndexing     bool
	BufferDeviceAddress    bool
	AccelerationStructure  bool
	DeferredHostOperations bool
	RayQuery               bool
}

func (c Capabilities) SupportsDynamicDDGI() bool {
	versionSupported := c.VulkanMajor > 1 || c.VulkanMajor == 1 && c.VulkanMinor >= 2
	return versionSupported && c.BufferDeviceAddress && c.AccelerationStructure &&
		c.DeferredHostOperations && c.RayQuery
}

type ContributionFlags uint8

const (
	ContributionOff ContributionFlags = iota
	ContributionStatic
	ContributionRigid
	ContributionReceivesOnly
)

type SceneInstance struct {
	ID            InstanceID
	Material      MaterialID
	Model         matrix.Mat4
	PreviousModel matrix.Mat4
	Bounds        graviton.AABB
	Contribution  ContributionFlags
}

type SceneLight struct {
	ID     LightID
	Bounds graviton.AABB
	Dirty  bool
}

type SceneDelta struct {
	AddedInstances   []SceneInstance
	UpdatedInstances []SceneInstance
	RemovedInstances []InstanceID
	UpdatedLights    []SceneLight
	EnvironmentDirty bool
}

type InvalidationReason uint8

const (
	InvalidationUnknown InvalidationReason = iota
	InvalidationGeometry
	InvalidationLight
	InvalidationEnvironment
	InvalidationScenario
)

type Invalidation struct {
	Bounds graviton.AABB
	Reason InvalidationReason
}

type FrameInputs struct {
	View             ViewID
	DirectLighting   rendering.FrameGraphResource
	Depth            rendering.FrameGraphResource
	NormalRoughness  rendering.FrameGraphResource
	AlbedoMetallic   rendering.FrameGraphResource
	Motion           rendering.FrameGraphResource
	HistoryReset     bool
	RuntimeSeconds   float32
	DeltaTimeSeconds float32
}

type Outputs struct {
	DiffuseIrradiance rendering.FrameGraphResource
	SpecularRadiance  rendering.FrameGraphResource
	Valid             bool
}

type ProbeFieldBinding struct {
	Irradiance rendering.FrameGraphResource
	Distance   rendering.FrameGraphResource
	Metadata   rendering.FrameGraphResource
	Valid      bool
}

type Stats struct {
	Provider           string
	GPUTimeMS          float32
	TraceTimeMS        float32
	ResolveTimeMS      float32
	AccelerationTimeMS float32
	MemoryUsedMB       uint32
	ActiveProbes       uint32
	InactiveProbes     uint32
	UpdatedProbes      uint32
	RaysTraced         uint64
	ResidentBricks     uint32
	Converged          bool
	FallbackReason     string
}

type DebugView struct {
	ID    string
	Label string
}

type ProviderContext struct {
	Capabilities Capabilities
	Assets       AssetReader
}

type Provider interface {
	ID() string
	Supports(Capabilities) bool
	Initialize(ProviderContext) error
	Configure(Settings) error
	SyncScene(SceneDelta) error
	AddUpdatePasses(*rendering.FrameGraph, FrameInputs) error
	AddResolvePasses(*rendering.FrameGraph, FrameInputs) (Outputs, error)
	ProbeField(ViewID) ProbeFieldBinding
	ShaderData(matrix.Vec3, float32) rendering.GlobalIlluminationForRender
	Invalidate(Invalidation)
	ResetHistory(ViewID)
	SetScenario(string) error
	Stats() Stats
	DebugViews() []DebugView
	Shutdown()
}

type ProviderFactory func() Provider
