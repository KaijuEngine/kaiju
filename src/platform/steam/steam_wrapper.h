#ifndef STEAM_WRAPPER_H
#define STEAM_WRAPPER_H
#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdbool.h>

extern void goOnGameOverlayActivated(bool);
extern void goOnUserStatsReceived(uint64_t, int);
extern void goOnUserStatsStored();

bool c_SteamAPI_Init();
void c_SteamAPI_Shutdown();
bool c_SteamAPI_RestartAppIfNecessary(uint32_t unOwnAppID);
void c_SteamAPI_RunCallbacks();

////////////////////////////////////////////////////////////////////////////////
// Steam Input                                                                //
////////////////////////////////////////////////////////////////////////////////
bool c_SteamInput_IsInitialized();
bool c_SteamInput_SetInputActionManifestFilePath(const char* path);
void c_SteamInput_RunFrame();
int c_SteamInput_GetConnectedControllers(uint64_t* handlesOut, int maxHandles);

uint64_t c_SteamInput_GetActionSetHandle(const char* actionSetName);
void c_SteamInput_ActivateActionSet(uint64_t inputHandle, uint64_t actionSetHandle);
uint64_t c_SteamInput_GetCurrentActionSet(uint64_t inputHandle);
void c_SteamInput_ActivateActionSetLayer(uint64_t inputHandle, uint64_t actionSetLayerHandle);
void c_SteamInput_DeactivateActionSetLayer(uint64_t inputHandle, uint64_t actionSetLayerHandle);
void c_SteamInput_DeactivateAllActionSetLayers(uint64_t inputHandle);
int c_SteamInput_GetActiveActionSetLayers(uint64_t inputHandle, uint64_t* handlesOut, int maxHandles);

uint64_t c_SteamInput_GetDigitalActionHandle(const char* actionName);
bool c_SteamInput_GetDigitalActionData(uint64_t inputHandle, uint64_t actionHandle,
	bool* state, bool* active);
int c_SteamInput_GetDigitalActionOrigins(uint64_t inputHandle, uint64_t actionSetHandle,
	uint64_t actionHandle, int32_t* originsOut, int maxOrigins);
const char* c_SteamInput_GetStringForDigitalActionName(uint64_t actionHandle);

uint64_t c_SteamInput_GetAnalogActionHandle(const char* actionName);
bool c_SteamInput_GetAnalogActionData(uint64_t inputHandle, uint64_t actionHandle,
	int32_t* mode, float* x, float* y, bool* active);
int c_SteamInput_GetAnalogActionOrigins(uint64_t inputHandle, uint64_t actionSetHandle,
	uint64_t actionHandle, int32_t* originsOut, int maxOrigins);
const char* c_SteamInput_GetStringForAnalogActionName(uint64_t actionHandle);
void c_SteamInput_StopAnalogActionMomentum(uint64_t inputHandle, uint64_t actionHandle);

const char* c_SteamInput_GetGlyphPNGForActionOrigin(int32_t origin, int32_t size, uint32_t flags);
const char* c_SteamInput_GetGlyphSVGForActionOrigin(int32_t origin, uint32_t flags);
const char* c_SteamInput_GetStringForActionOrigin(int32_t origin);
bool c_SteamInput_ShowBindingPanel(uint64_t inputHandle);
int32_t c_SteamInput_GetInputTypeForHandle(uint64_t inputHandle);
uint64_t c_SteamInput_GetControllerForGamepadIndex(int index);
int c_SteamInput_GetGamepadIndexForController(uint64_t inputHandle);

bool c_SteamInput_GetMotionData(uint64_t inputHandle,
	float* rotQuatX, float* rotQuatY, float* rotQuatZ, float* rotQuatW,
	float* posAccelX, float* posAccelY, float* posAccelZ,
	float* rotVelX, float* rotVelY, float* rotVelZ);
void c_SteamInput_TriggerVibration(uint64_t inputHandle, uint16_t leftSpeed, uint16_t rightSpeed);
void c_SteamInput_TriggerVibrationExtended(uint64_t inputHandle,
	uint16_t leftSpeed, uint16_t rightSpeed, uint16_t leftTriggerSpeed, uint16_t rightTriggerSpeed);
void c_SteamInput_TriggerSimpleHapticEvent(uint64_t inputHandle, int32_t location,
	uint8_t intensity, int8_t gainDB, uint8_t otherIntensity, int8_t otherGainDB);

////////////////////////////////////////////////////////////////////////////////
// Steam Friends                                                              //
////////////////////////////////////////////////////////////////////////////////
const char* c_SteamAPI_SteamFriends_GetPersonalName();

////////////////////////////////////////////////////////////////////////////////
// Steam User                                                                 //
////////////////////////////////////////////////////////////////////////////////
bool c_SteamUser_BLoggedOn();

////////////////////////////////////////////////////////////////////////////////
// Steam User Stats                                                           //
////////////////////////////////////////////////////////////////////////////////
bool c_SteamUserStats_RequestCurrentStats();

////////////////////////////////////////////////////////////////////////////////
// Steam Utils                                                                //
////////////////////////////////////////////////////////////////////////////////
int64_t c_SteamUtils_GetAppID();

#ifdef __cplusplus
}
#endif
#endif
