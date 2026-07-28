/******************************************************************************/
/* steam_input.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package steam

/*
#include <stdlib.h>
#include "steam_wrapper.h"
*/
import "C"

import (
	"path/filepath"
	"sync"
	"unsafe"

	"kaijuengine.com/matrix"
)

const (
	InputMaxControllers  = 16
	InputMaxOrigins      = 8
	InputMaxActiveLayers = 16
)

type InputHandle uint64
type InputActionSetHandle uint64
type InputDigitalActionHandle uint64
type InputAnalogActionHandle uint64
type InputActionOrigin int32
type InputSourceMode int32
type InputType int32
type InputGlyphSize int32
type InputGlyphStyle uint32
type InputHapticLocation int32

const (
	InputHandleInvalid        InputHandle = 0
	InputHandleAllControllers InputHandle = ^InputHandle(0)

	InputActionSetHandleInvalid     InputActionSetHandle     = 0
	InputDigitalActionHandleInvalid InputDigitalActionHandle = 0
	InputAnalogActionHandleInvalid  InputAnalogActionHandle  = 0
	InputActionOriginInvalid        InputActionOrigin        = -1
	InputGamepadIndexInvalid                                 = -1
)

const (
	InputSourceModeNone InputSourceMode = iota
	InputSourceModeDPad
	InputSourceModeButtons
	InputSourceModeFourButtons
	InputSourceModeAbsoluteMouse
	InputSourceModeRelativeMouse
	InputSourceModeJoystickMove
	InputSourceModeJoystickMouse
	InputSourceModeJoystickCamera
	InputSourceModeScrollWheel
	InputSourceModeTrigger
	InputSourceModeTouchMenu
	InputSourceModeMouseJoystick
	InputSourceModeMouseRegion
	InputSourceModeRadialMenu
	InputSourceModeSingleButton
	InputSourceModeSwitches
)

const (
	InputTypeUnknown InputType = iota
	InputTypeSteamController
	InputTypeXbox360Controller
	InputTypeXboxOneController
	InputTypeGenericGamepad
	InputTypePS4Controller
	InputTypeAppleMFiController
	InputTypeAndroidController
	InputTypeSwitchJoyConPair
	InputTypeSwitchJoyConSingle
	InputTypeSwitchProController
	InputTypeMobileTouch
	InputTypePS3Controller
	InputTypePS5Controller
	InputTypeSteamDeckController
	InputTypeCount
)

const (
	InputGlyphSizeSmall InputGlyphSize = iota
	InputGlyphSizeMedium
	InputGlyphSizeLarge
)

const (
	InputGlyphStyleKnockout         InputGlyphStyle = 0
	InputGlyphStyleLight            InputGlyphStyle = 0x1
	InputGlyphStyleDark             InputGlyphStyle = 0x2
	InputGlyphStyleNeutralColorABXY InputGlyphStyle = 0x10
	InputGlyphStyleSolidABXY        InputGlyphStyle = 0x20
)

const (
	InputHapticLocationLeft  InputHapticLocation = 1
	InputHapticLocationRight InputHapticLocation = 2
	InputHapticLocationBoth  InputHapticLocation = 3
)

type DigitalActionData struct {
	State    bool
	Active   bool
	Pressed  bool
	Held     bool
	Released bool
}

type AnalogActionData struct {
	Mode   InputSourceMode
	X      matrix.Float
	Y      matrix.Float
	Active bool
}

type InputMotionData struct {
	Rotation        matrix.Quaternion
	Acceleration    matrix.Vec3
	AngularVelocity matrix.Vec3
}

type digitalActionKey struct {
	input  InputHandle
	action InputDigitalActionHandle
}

type analogActionKey struct {
	input  InputHandle
	action InputAnalogActionHandle
}

// InputSystem owns the native Steam Input interface and caches registered
// action state once per Steam callback frame.
type InputSystem struct {
	mu             sync.RWMutex
	pollMu         sync.Mutex
	initialized    bool
	controllers    []InputHandle
	digitalActions map[InputDigitalActionHandle]string
	analogActions  map[InputAnalogActionHandle]string
	digitalData    map[digitalActionKey]DigitalActionData
	analogData     map[analogActionKey]AnalogActionData
}

func newInputSystem() *InputSystem {
	return &InputSystem{
		digitalActions: make(map[InputDigitalActionHandle]string),
		analogActions:  make(map[InputAnalogActionHandle]string),
		digitalData:    make(map[digitalActionKey]DigitalActionData),
		analogData:     make(map[analogActionKey]AnalogActionData),
	}
}

func (s *InputSystem) initialize() bool {
	nativeInitialized := bool(C.c_SteamInput_IsInitialized())
	s.mu.Lock()
	defer s.mu.Unlock()
	s.initialized = nativeInitialized
	s.controllers = nil
	s.digitalActions = make(map[InputDigitalActionHandle]string)
	s.analogActions = make(map[InputAnalogActionHandle]string)
	s.digitalData = make(map[digitalActionKey]DigitalActionData)
	s.analogData = make(map[analogActionKey]AnalogActionData)
	return nativeInitialized
}

func (s *InputSystem) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.initialized = false
	s.controllers = nil
	s.digitalActions = make(map[InputDigitalActionHandle]string)
	s.analogActions = make(map[InputAnalogActionHandle]string)
	s.digitalData = make(map[digitalActionKey]DigitalActionData)
	s.analogData = make(map[analogActionKey]AnalogActionData)
}

func (s *InputSystem) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.initialized
}

// SetActionManifestFilePath selects a bundled Steam Input action manifest.
// Steam requires an absolute path; relative paths are resolved from the
// process working directory.
func (s *InputSystem) SetActionManifestFilePath(path string) bool {
	if !s.IsInitialized() || path == "" {
		return false
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	cPath := C.CString(absolutePath)
	defer C.free(unsafe.Pointer(cPath))
	return bool(C.c_SteamInput_SetInputActionManifestFilePath(cPath))
}

// RunFrame explicitly refreshes Steam Input and the engine-side action cache.
// The normal engine bootstrap already refreshes the cache through
// SteamAPI_RunCallbacks, so most games do not need to call this.
func (s *InputSystem) RunFrame() {
	if !s.IsInitialized() {
		return
	}
	C.c_SteamInput_RunFrame()
	s.poll()
}

// Controllers returns the connected Steam Input controller handles from the
// most recent frame.
func (s *InputSystem) Controllers() []InputHandle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]InputHandle(nil), s.controllers...)
}

func (s *InputSystem) ActionSetHandle(name string) InputActionSetHandle {
	if !s.IsInitialized() || name == "" {
		return InputActionSetHandleInvalid
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return InputActionSetHandle(C.c_SteamInput_GetActionSetHandle(cName))
}

func (s *InputSystem) ActivateActionSet(input InputHandle, actionSet InputActionSetHandle) {
	if s.IsInitialized() {
		C.c_SteamInput_ActivateActionSet(C.uint64_t(input), C.uint64_t(actionSet))
	}
}

func (s *InputSystem) CurrentActionSet(input InputHandle) InputActionSetHandle {
	if !s.IsInitialized() {
		return InputActionSetHandleInvalid
	}
	return InputActionSetHandle(C.c_SteamInput_GetCurrentActionSet(C.uint64_t(input)))
}

func (s *InputSystem) ActivateActionSetLayer(input InputHandle, layer InputActionSetHandle) {
	if s.IsInitialized() {
		C.c_SteamInput_ActivateActionSetLayer(C.uint64_t(input), C.uint64_t(layer))
	}
}

func (s *InputSystem) DeactivateActionSetLayer(input InputHandle, layer InputActionSetHandle) {
	if s.IsInitialized() {
		C.c_SteamInput_DeactivateActionSetLayer(C.uint64_t(input), C.uint64_t(layer))
	}
}

func (s *InputSystem) DeactivateAllActionSetLayers(input InputHandle) {
	if s.IsInitialized() {
		C.c_SteamInput_DeactivateAllActionSetLayers(C.uint64_t(input))
	}
}

func (s *InputSystem) ActiveActionSetLayers(input InputHandle) []InputActionSetHandle {
	if !s.IsInitialized() {
		return nil
	}
	var native [InputMaxActiveLayers]C.uint64_t
	count := int(C.c_SteamInput_GetActiveActionSetLayers(
		C.uint64_t(input), &native[0], C.int(len(native))))
	layers := make([]InputActionSetHandle, count)
	for i := range count {
		layers[i] = InputActionSetHandle(native[i])
	}
	return layers
}

// DigitalActionHandle resolves and registers a manifest action for automatic
// per-frame polling.
func (s *InputSystem) DigitalActionHandle(name string) InputDigitalActionHandle {
	if !s.IsInitialized() || name == "" {
		return InputDigitalActionHandleInvalid
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	handle := InputDigitalActionHandle(C.c_SteamInput_GetDigitalActionHandle(cName))
	if handle != InputDigitalActionHandleInvalid {
		s.mu.Lock()
		s.digitalActions[handle] = name
		s.mu.Unlock()
	}
	return handle
}

// DigitalAction returns the cached state produced before game updates in the
// current frame.
func (s *InputSystem) DigitalAction(input InputHandle, action InputDigitalActionHandle) DigitalActionData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.digitalData[digitalActionKey{input: input, action: action}]
}

func (s *InputSystem) DigitalActionOrigins(input InputHandle, actionSet InputActionSetHandle,
	action InputDigitalActionHandle) []InputActionOrigin {
	if !s.IsInitialized() {
		return nil
	}
	var native [InputMaxOrigins]C.int32_t
	count := int(C.c_SteamInput_GetDigitalActionOrigins(
		C.uint64_t(input), C.uint64_t(actionSet), C.uint64_t(action),
		&native[0], C.int(len(native))))
	origins := make([]InputActionOrigin, count)
	for i := range count {
		origins[i] = InputActionOrigin(native[i])
	}
	return origins
}

func (s *InputSystem) DigitalActionName(action InputDigitalActionHandle) string {
	if !s.IsInitialized() {
		return ""
	}
	return cGoString(C.c_SteamInput_GetStringForDigitalActionName(C.uint64_t(action)))
}

// AnalogActionHandle resolves and registers a manifest action for automatic
// per-frame polling.
func (s *InputSystem) AnalogActionHandle(name string) InputAnalogActionHandle {
	if !s.IsInitialized() || name == "" {
		return InputAnalogActionHandleInvalid
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	handle := InputAnalogActionHandle(C.c_SteamInput_GetAnalogActionHandle(cName))
	if handle != InputAnalogActionHandleInvalid {
		s.mu.Lock()
		s.analogActions[handle] = name
		s.mu.Unlock()
	}
	return handle
}

// AnalogAction returns the cached state produced before game updates in the
// current frame.
func (s *InputSystem) AnalogAction(input InputHandle, action InputAnalogActionHandle) AnalogActionData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.analogData[analogActionKey{input: input, action: action}]
}

func (s *InputSystem) AnalogActionOrigins(input InputHandle, actionSet InputActionSetHandle,
	action InputAnalogActionHandle) []InputActionOrigin {
	if !s.IsInitialized() {
		return nil
	}
	var native [InputMaxOrigins]C.int32_t
	count := int(C.c_SteamInput_GetAnalogActionOrigins(
		C.uint64_t(input), C.uint64_t(actionSet), C.uint64_t(action),
		&native[0], C.int(len(native))))
	origins := make([]InputActionOrigin, count)
	for i := range count {
		origins[i] = InputActionOrigin(native[i])
	}
	return origins
}

func (s *InputSystem) AnalogActionName(action InputAnalogActionHandle) string {
	if !s.IsInitialized() {
		return ""
	}
	return cGoString(C.c_SteamInput_GetStringForAnalogActionName(C.uint64_t(action)))
}

func (s *InputSystem) StopAnalogActionMomentum(input InputHandle, action InputAnalogActionHandle) {
	if s.IsInitialized() {
		C.c_SteamInput_StopAnalogActionMomentum(C.uint64_t(input), C.uint64_t(action))
	}
}

func (s *InputSystem) GlyphPNG(origin InputActionOrigin, size InputGlyphSize,
	style InputGlyphStyle) string {
	if !s.IsInitialized() {
		return ""
	}
	return cGoString(C.c_SteamInput_GetGlyphPNGForActionOrigin(
		C.int32_t(origin), C.int32_t(size), C.uint32_t(style)))
}

func (s *InputSystem) GlyphSVG(origin InputActionOrigin, style InputGlyphStyle) string {
	if !s.IsInitialized() {
		return ""
	}
	return cGoString(C.c_SteamInput_GetGlyphSVGForActionOrigin(
		C.int32_t(origin), C.uint32_t(style)))
}

func (s *InputSystem) OriginName(origin InputActionOrigin) string {
	if !s.IsInitialized() {
		return ""
	}
	return cGoString(C.c_SteamInput_GetStringForActionOrigin(C.int32_t(origin)))
}

func (s *InputSystem) ShowBindingPanel(input InputHandle) bool {
	return s.IsInitialized() &&
		bool(C.c_SteamInput_ShowBindingPanel(C.uint64_t(input)))
}

func (s *InputSystem) InputType(input InputHandle) InputType {
	if !s.IsInitialized() {
		return InputTypeUnknown
	}
	return InputType(C.c_SteamInput_GetInputTypeForHandle(C.uint64_t(input)))
}

func (s *InputSystem) ControllerForGamepadIndex(index int) InputHandle {
	if !s.IsInitialized() {
		return InputHandleInvalid
	}
	return InputHandle(C.c_SteamInput_GetControllerForGamepadIndex(C.int(index)))
}

func (s *InputSystem) GamepadIndexForController(input InputHandle) int {
	if !s.IsInitialized() {
		return InputGamepadIndexInvalid
	}
	return int(C.c_SteamInput_GetGamepadIndexForController(C.uint64_t(input)))
}

func (s *InputSystem) MotionData(input InputHandle) (InputMotionData, bool) {
	if !s.IsInitialized() {
		return InputMotionData{}, false
	}
	var qx, qy, qz, qw C.float
	var ax, ay, az C.float
	var vx, vy, vz C.float
	ok := bool(C.c_SteamInput_GetMotionData(
		C.uint64_t(input),
		&qx, &qy, &qz, &qw,
		&ax, &ay, &az,
		&vx, &vy, &vz))
	if !ok {
		return InputMotionData{}, false
	}
	return InputMotionData{
		Rotation: matrix.QuaternionFromXYZW([4]matrix.Float{
			matrix.Float(qx), matrix.Float(qy), matrix.Float(qz), matrix.Float(qw),
		}),
		Acceleration: matrix.NewVec3(
			matrix.Float(ax), matrix.Float(ay), matrix.Float(az)),
		AngularVelocity: matrix.NewVec3(
			matrix.Float(vx), matrix.Float(vy), matrix.Float(vz)),
	}, true
}

func (s *InputSystem) TriggerVibration(input InputHandle, leftSpeed, rightSpeed uint16) {
	if s.IsInitialized() {
		C.c_SteamInput_TriggerVibration(
			C.uint64_t(input), C.uint16_t(leftSpeed), C.uint16_t(rightSpeed))
	}
}

func (s *InputSystem) TriggerVibrationExtended(input InputHandle,
	leftSpeed, rightSpeed, leftTriggerSpeed, rightTriggerSpeed uint16) {
	if s.IsInitialized() {
		C.c_SteamInput_TriggerVibrationExtended(
			C.uint64_t(input),
			C.uint16_t(leftSpeed), C.uint16_t(rightSpeed),
			C.uint16_t(leftTriggerSpeed), C.uint16_t(rightTriggerSpeed))
	}
}

func (s *InputSystem) TriggerSimpleHapticEvent(input InputHandle, location InputHapticLocation,
	intensity uint8, gainDB int8, otherIntensity uint8, otherGainDB int8) {
	if s.IsInitialized() {
		C.c_SteamInput_TriggerSimpleHapticEvent(
			C.uint64_t(input), C.int32_t(location),
			C.uint8_t(intensity), C.int8_t(gainDB),
			C.uint8_t(otherIntensity), C.int8_t(otherGainDB))
	}
}

func (s *InputSystem) poll() {
	if !s.IsInitialized() {
		return
	}
	s.pollMu.Lock()
	defer s.pollMu.Unlock()

	controllers := nativeInputControllers()

	s.mu.RLock()
	digitalActions := make([]InputDigitalActionHandle, 0, len(s.digitalActions))
	for action := range s.digitalActions {
		digitalActions = append(digitalActions, action)
	}
	analogActions := make([]InputAnalogActionHandle, 0, len(s.analogActions))
	for action := range s.analogActions {
		analogActions = append(analogActions, action)
	}
	previousDigital := make(map[digitalActionKey]DigitalActionData, len(s.digitalData))
	for key, data := range s.digitalData {
		previousDigital[key] = data
	}
	s.mu.RUnlock()

	digitalData := make(map[digitalActionKey]DigitalActionData,
		len(controllers)*len(digitalActions))
	analogData := make(map[analogActionKey]AnalogActionData,
		len(controllers)*len(analogActions))

	for _, input := range controllers {
		for _, action := range digitalActions {
			key := digitalActionKey{input: input, action: action}
			state, active, ok := nativeDigitalActionData(input, action)
			if ok {
				digitalData[key] = nextDigitalActionData(previousDigital[key], state, active)
			}
		}
		for _, action := range analogActions {
			key := analogActionKey{input: input, action: action}
			if data, ok := nativeAnalogActionData(input, action); ok {
				analogData[key] = data
			}
		}
	}

	s.mu.Lock()
	s.controllers = controllers
	s.digitalData = digitalData
	s.analogData = analogData
	s.mu.Unlock()
}

func nativeInputControllers() []InputHandle {
	var native [InputMaxControllers]C.uint64_t
	count := int(C.c_SteamInput_GetConnectedControllers(&native[0], C.int(len(native))))
	controllers := make([]InputHandle, count)
	for i := range count {
		controllers[i] = InputHandle(native[i])
	}
	return controllers
}

func nativeDigitalActionData(input InputHandle,
	action InputDigitalActionHandle) (state, active, ok bool) {
	var nativeState, nativeActive C.bool
	ok = bool(C.c_SteamInput_GetDigitalActionData(
		C.uint64_t(input), C.uint64_t(action), &nativeState, &nativeActive))
	return bool(nativeState), bool(nativeActive), ok
}

func nativeAnalogActionData(input InputHandle,
	action InputAnalogActionHandle) (AnalogActionData, bool) {
	var mode C.int32_t
	var x, y C.float
	var active C.bool
	ok := bool(C.c_SteamInput_GetAnalogActionData(
		C.uint64_t(input), C.uint64_t(action), &mode, &x, &y, &active))
	if !ok {
		return AnalogActionData{}, false
	}
	return AnalogActionData{
		Mode:   InputSourceMode(mode),
		X:      matrix.Float(x),
		Y:      matrix.Float(y),
		Active: bool(active),
	}, true
}

func nextDigitalActionData(previous DigitalActionData, state, active bool) DigitalActionData {
	wasDown := previous.State && previous.Active
	isDown := state && active
	return DigitalActionData{
		State:    state,
		Active:   active,
		Pressed:  isDown && !wasDown,
		Held:     isDown && wasDown,
		Released: !isDown && wasDown,
	}
}

func cGoString(value *C.char) string {
	if value == nil {
		return ""
	}
	return C.GoString(value)
}
