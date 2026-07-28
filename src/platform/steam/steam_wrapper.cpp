#if BUILD_STEAM_API
#include "steam_wrapper.h"
#include "../../../publishing/steam_sdk/public/steam/steam_api_flat.h"

static ISteamInput* sSteamInput = nullptr;
static bool sSteamInputInitialized = false;

class SteamGameCallbacks {
private:
	STEAM_CALLBACK(SteamGameCallbacks, OnGameOverlayActivated, GameOverlayActivated_t);
	STEAM_CALLBACK(SteamGameCallbacks, OnUserStatsReceived, UserStatsReceived_t);
	STEAM_CALLBACK(SteamGameCallbacks, OnUserStatsStored, UserStatsStored_t);
};

void SteamGameCallbacks::OnGameOverlayActivated(GameOverlayActivated_t* pCallback) {
	goOnGameOverlayActivated((bool)pCallback->m_bActive);
}

void SteamGameCallbacks::OnUserStatsReceived(UserStatsReceived_t* pCallback) {
	goOnUserStatsReceived(pCallback->m_nGameID, pCallback->m_eResult);
}

void SteamGameCallbacks::OnUserStatsStored(UserStatsStored_t* pCallback) {
	goOnUserStatsStored();
}

static SteamGameCallbacks* sSteamGameCallbacks = nullptr;
static inline void register_callbacks() { sSteamGameCallbacks = new SteamGameCallbacks(); }
static inline void unregister_callbacks() { delete sSteamGameCallbacks; }

extern "C" {
	bool c_SteamAPI_Init() {
		if (SteamAPI_Init()) {
			register_callbacks();
			sSteamInput = SteamAPI_SteamInput();
			if (sSteamInput != nullptr) {
				sSteamInputInitialized = SteamAPI_ISteamInput_Init(sSteamInput, false);
			}
			return true;
		}
		return false;
	}
	void c_SteamAPI_Shutdown() {
		if (sSteamInputInitialized) {
			SteamAPI_ISteamInput_Shutdown(sSteamInput);
		}
		sSteamInputInitialized = false;
		sSteamInput = nullptr;
		unregister_callbacks();
		SteamAPI_Shutdown();
	}
	bool c_SteamAPI_RestartAppIfNecessary(uint32_t unOwnAppID) {
		return SteamAPI_RestartAppIfNecessary(unOwnAppID);
	}
	void c_SteamAPI_RunCallbacks() { SteamAPI_RunCallbacks(); }

	////////////////////////////////////////////////////////////////////////////
	// Steam Input                                                            //
	////////////////////////////////////////////////////////////////////////////
	bool c_SteamInput_IsInitialized() {
		return sSteamInputInitialized && sSteamInput != nullptr;
	}

	bool c_SteamInput_SetInputActionManifestFilePath(const char* path) {
		return c_SteamInput_IsInitialized() && path != nullptr &&
			SteamAPI_ISteamInput_SetInputActionManifestFilePath(sSteamInput, path);
	}

	void c_SteamInput_RunFrame() {
		if (c_SteamInput_IsInitialized()) {
			SteamAPI_ISteamInput_RunFrame(sSteamInput, true);
		}
	}

	int c_SteamInput_GetConnectedControllers(uint64_t* handlesOut, int maxHandles) {
		if (!c_SteamInput_IsInitialized() || handlesOut == nullptr || maxHandles <= 0) {
			return 0;
		}
		InputHandle_t handles[STEAM_INPUT_MAX_COUNT] = {};
		const int count = SteamAPI_ISteamInput_GetConnectedControllers(sSteamInput, handles);
		const int copyCount = count < maxHandles ? count : maxHandles;
		for (int i = 0; i < copyCount; i++) {
			handlesOut[i] = handles[i];
		}
		return copyCount;
	}

	uint64_t c_SteamInput_GetActionSetHandle(const char* actionSetName) {
		if (!c_SteamInput_IsInitialized() || actionSetName == nullptr) {
			return 0;
		}
		return SteamAPI_ISteamInput_GetActionSetHandle(sSteamInput, actionSetName);
	}

	void c_SteamInput_ActivateActionSet(uint64_t inputHandle, uint64_t actionSetHandle) {
		if (c_SteamInput_IsInitialized() && inputHandle != 0 && actionSetHandle != 0) {
			SteamAPI_ISteamInput_ActivateActionSet(sSteamInput, inputHandle, actionSetHandle);
		}
	}

	uint64_t c_SteamInput_GetCurrentActionSet(uint64_t inputHandle) {
		if (!c_SteamInput_IsInitialized() || inputHandle == 0) {
			return 0;
		}
		return SteamAPI_ISteamInput_GetCurrentActionSet(sSteamInput, inputHandle);
	}

	void c_SteamInput_ActivateActionSetLayer(uint64_t inputHandle, uint64_t actionSetLayerHandle) {
		if (c_SteamInput_IsInitialized() && inputHandle != 0 && actionSetLayerHandle != 0) {
			SteamAPI_ISteamInput_ActivateActionSetLayer(sSteamInput, inputHandle, actionSetLayerHandle);
		}
	}

	void c_SteamInput_DeactivateActionSetLayer(uint64_t inputHandle, uint64_t actionSetLayerHandle) {
		if (c_SteamInput_IsInitialized() && inputHandle != 0 && actionSetLayerHandle != 0) {
			SteamAPI_ISteamInput_DeactivateActionSetLayer(sSteamInput, inputHandle, actionSetLayerHandle);
		}
	}

	void c_SteamInput_DeactivateAllActionSetLayers(uint64_t inputHandle) {
		if (c_SteamInput_IsInitialized() && inputHandle != 0) {
			SteamAPI_ISteamInput_DeactivateAllActionSetLayers(sSteamInput, inputHandle);
		}
	}

	int c_SteamInput_GetActiveActionSetLayers(uint64_t inputHandle, uint64_t* handlesOut, int maxHandles) {
		if (!c_SteamInput_IsInitialized() || inputHandle == 0 ||
			handlesOut == nullptr || maxHandles <= 0) {
			return 0;
		}
		InputActionSetHandle_t handles[STEAM_INPUT_MAX_ACTIVE_LAYERS] = {};
		const int count = SteamAPI_ISteamInput_GetActiveActionSetLayers(sSteamInput, inputHandle, handles);
		const int copyCount = count < maxHandles ? count : maxHandles;
		for (int i = 0; i < copyCount; i++) {
			handlesOut[i] = handles[i];
		}
		return copyCount;
	}

	uint64_t c_SteamInput_GetDigitalActionHandle(const char* actionName) {
		if (!c_SteamInput_IsInitialized() || actionName == nullptr) {
			return 0;
		}
		return SteamAPI_ISteamInput_GetDigitalActionHandle(sSteamInput, actionName);
	}

	bool c_SteamInput_GetDigitalActionData(uint64_t inputHandle, uint64_t actionHandle,
		bool* state, bool* active) {
		if (!c_SteamInput_IsInitialized() || inputHandle == 0 || actionHandle == 0 ||
			state == nullptr || active == nullptr) {
			return false;
		}
		const InputDigitalActionData_t data = SteamAPI_ISteamInput_GetDigitalActionData(
			sSteamInput, inputHandle, actionHandle);
		*state = data.bState;
		*active = data.bActive;
		return true;
	}

	int c_SteamInput_GetDigitalActionOrigins(uint64_t inputHandle, uint64_t actionSetHandle,
		uint64_t actionHandle, int32_t* originsOut, int maxOrigins) {
		if (!c_SteamInput_IsInitialized() || inputHandle == 0 || actionSetHandle == 0 ||
			actionHandle == 0 || originsOut == nullptr || maxOrigins <= 0) {
			return 0;
		}
		EInputActionOrigin origins[STEAM_INPUT_MAX_ORIGINS] = {};
		const int count = SteamAPI_ISteamInput_GetDigitalActionOrigins(
			sSteamInput, inputHandle, actionSetHandle, actionHandle, origins);
		const int copyCount = count < maxOrigins ? count : maxOrigins;
		for (int i = 0; i < copyCount; i++) {
			originsOut[i] = static_cast<int32_t>(origins[i]);
		}
		return copyCount;
	}

	const char* c_SteamInput_GetStringForDigitalActionName(uint64_t actionHandle) {
		if (!c_SteamInput_IsInitialized() || actionHandle == 0) {
			return "";
		}
		return SteamAPI_ISteamInput_GetStringForDigitalActionName(sSteamInput, actionHandle);
	}

	uint64_t c_SteamInput_GetAnalogActionHandle(const char* actionName) {
		if (!c_SteamInput_IsInitialized() || actionName == nullptr) {
			return 0;
		}
		return SteamAPI_ISteamInput_GetAnalogActionHandle(sSteamInput, actionName);
	}

	bool c_SteamInput_GetAnalogActionData(uint64_t inputHandle, uint64_t actionHandle,
		int32_t* mode, float* x, float* y, bool* active) {
		if (!c_SteamInput_IsInitialized() || inputHandle == 0 || actionHandle == 0 ||
			mode == nullptr || x == nullptr || y == nullptr || active == nullptr) {
			return false;
		}
		const InputAnalogActionData_t data = SteamAPI_ISteamInput_GetAnalogActionData(
			sSteamInput, inputHandle, actionHandle);
		*mode = static_cast<int32_t>(data.eMode);
		*x = data.x;
		*y = data.y;
		*active = data.bActive;
		return true;
	}

	int c_SteamInput_GetAnalogActionOrigins(uint64_t inputHandle, uint64_t actionSetHandle,
		uint64_t actionHandle, int32_t* originsOut, int maxOrigins) {
		if (!c_SteamInput_IsInitialized() || inputHandle == 0 || actionSetHandle == 0 ||
			actionHandle == 0 || originsOut == nullptr || maxOrigins <= 0) {
			return 0;
		}
		EInputActionOrigin origins[STEAM_INPUT_MAX_ORIGINS] = {};
		const int count = SteamAPI_ISteamInput_GetAnalogActionOrigins(
			sSteamInput, inputHandle, actionSetHandle, actionHandle, origins);
		const int copyCount = count < maxOrigins ? count : maxOrigins;
		for (int i = 0; i < copyCount; i++) {
			originsOut[i] = static_cast<int32_t>(origins[i]);
		}
		return copyCount;
	}

	const char* c_SteamInput_GetStringForAnalogActionName(uint64_t actionHandle) {
		if (!c_SteamInput_IsInitialized() || actionHandle == 0) {
			return "";
		}
		return SteamAPI_ISteamInput_GetStringForAnalogActionName(sSteamInput, actionHandle);
	}

	void c_SteamInput_StopAnalogActionMomentum(uint64_t inputHandle, uint64_t actionHandle) {
		if (c_SteamInput_IsInitialized() && inputHandle != 0 && actionHandle != 0) {
			SteamAPI_ISteamInput_StopAnalogActionMomentum(sSteamInput, inputHandle, actionHandle);
		}
	}

	const char* c_SteamInput_GetGlyphPNGForActionOrigin(int32_t origin, int32_t size, uint32_t flags) {
		if (!c_SteamInput_IsInitialized()) {
			return "";
		}
		return SteamAPI_ISteamInput_GetGlyphPNGForActionOrigin(
			sSteamInput, static_cast<EInputActionOrigin>(origin),
			static_cast<ESteamInputGlyphSize>(size), flags);
	}

	const char* c_SteamInput_GetGlyphSVGForActionOrigin(int32_t origin, uint32_t flags) {
		if (!c_SteamInput_IsInitialized()) {
			return "";
		}
		return SteamAPI_ISteamInput_GetGlyphSVGForActionOrigin(
			sSteamInput, static_cast<EInputActionOrigin>(origin), flags);
	}

	const char* c_SteamInput_GetStringForActionOrigin(int32_t origin) {
		if (!c_SteamInput_IsInitialized()) {
			return "";
		}
		return SteamAPI_ISteamInput_GetStringForActionOrigin(
			sSteamInput, static_cast<EInputActionOrigin>(origin));
	}

	bool c_SteamInput_ShowBindingPanel(uint64_t inputHandle) {
		return c_SteamInput_IsInitialized() && inputHandle != 0 &&
			SteamAPI_ISteamInput_ShowBindingPanel(sSteamInput, inputHandle);
	}

	int32_t c_SteamInput_GetInputTypeForHandle(uint64_t inputHandle) {
		if (!c_SteamInput_IsInitialized() || inputHandle == 0) {
			return static_cast<int32_t>(k_ESteamInputType_Unknown);
		}
		return static_cast<int32_t>(
			SteamAPI_ISteamInput_GetInputTypeForHandle(sSteamInput, inputHandle));
	}

	uint64_t c_SteamInput_GetControllerForGamepadIndex(int index) {
		if (!c_SteamInput_IsInitialized()) {
			return 0;
		}
		return SteamAPI_ISteamInput_GetControllerForGamepadIndex(sSteamInput, index);
	}

	int c_SteamInput_GetGamepadIndexForController(uint64_t inputHandle) {
		if (!c_SteamInput_IsInitialized() || inputHandle == 0) {
			return -1;
		}
		return SteamAPI_ISteamInput_GetGamepadIndexForController(sSteamInput, inputHandle);
	}

	bool c_SteamInput_GetMotionData(uint64_t inputHandle,
		float* rotQuatX, float* rotQuatY, float* rotQuatZ, float* rotQuatW,
		float* posAccelX, float* posAccelY, float* posAccelZ,
		float* rotVelX, float* rotVelY, float* rotVelZ) {
		if (!c_SteamInput_IsInitialized() || inputHandle == 0 ||
			rotQuatX == nullptr || rotQuatY == nullptr || rotQuatZ == nullptr || rotQuatW == nullptr ||
			posAccelX == nullptr || posAccelY == nullptr || posAccelZ == nullptr ||
			rotVelX == nullptr || rotVelY == nullptr || rotVelZ == nullptr) {
			return false;
		}
		const InputMotionData_t data = SteamAPI_ISteamInput_GetMotionData(sSteamInput, inputHandle);
		*rotQuatX = data.rotQuatX;
		*rotQuatY = data.rotQuatY;
		*rotQuatZ = data.rotQuatZ;
		*rotQuatW = data.rotQuatW;
		*posAccelX = data.posAccelX;
		*posAccelY = data.posAccelY;
		*posAccelZ = data.posAccelZ;
		*rotVelX = data.rotVelX;
		*rotVelY = data.rotVelY;
		*rotVelZ = data.rotVelZ;
		return true;
	}

	void c_SteamInput_TriggerVibration(uint64_t inputHandle, uint16_t leftSpeed, uint16_t rightSpeed) {
		if (c_SteamInput_IsInitialized() && inputHandle != 0) {
			SteamAPI_ISteamInput_TriggerVibration(sSteamInput, inputHandle, leftSpeed, rightSpeed);
		}
	}

	void c_SteamInput_TriggerVibrationExtended(uint64_t inputHandle,
		uint16_t leftSpeed, uint16_t rightSpeed, uint16_t leftTriggerSpeed, uint16_t rightTriggerSpeed) {
		if (c_SteamInput_IsInitialized() && inputHandle != 0) {
			SteamAPI_ISteamInput_TriggerVibrationExtended(sSteamInput, inputHandle,
				leftSpeed, rightSpeed, leftTriggerSpeed, rightTriggerSpeed);
		}
	}

	void c_SteamInput_TriggerSimpleHapticEvent(uint64_t inputHandle, int32_t location,
		uint8_t intensity, int8_t gainDB, uint8_t otherIntensity, int8_t otherGainDB) {
		if (c_SteamInput_IsInitialized() && inputHandle != 0) {
			SteamAPI_ISteamInput_TriggerSimpleHapticEvent(sSteamInput, inputHandle,
				static_cast<EControllerHapticLocation>(location), intensity,
				static_cast<char>(gainDB), otherIntensity, static_cast<char>(otherGainDB));
		}
	}

	////////////////////////////////////////////////////////////////////////////
	// Steam Friends                                                          //
	////////////////////////////////////////////////////////////////////////////
	const char* c_SteamAPI_SteamFriends_GetPersonalName() {
		return SteamFriends()->GetPersonaName();
	}

	////////////////////////////////////////////////////////////////////////////
	// Steam User                                                             //
	////////////////////////////////////////////////////////////////////////////
	bool c_SteamUser_BLoggedOn() {
		return SteamUser() != nullptr && SteamUser()->BLoggedOn();
	}

	////////////////////////////////////////////////////////////////////////////
	// Steam User Stats                                                       //
	////////////////////////////////////////////////////////////////////////////
	bool c_SteamUserStats_RequestCurrentStats() {
		return SteamUserStats()->RequestCurrentStats();
	}

	////////////////////////////////////////////////////////////////////////////
	// Steam Utils                                                            //
	////////////////////////////////////////////////////////////////////////////
	int64_t c_SteamUtils_GetAppID() { return SteamUtils()->GetAppID(); }
}

#endif
