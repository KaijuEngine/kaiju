#include "steam_wrapper.h"
#include "../../../publishing/steam_sdk/public/steam/steam_api_flat.h"

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
			return true;
		}
		return false;
	}
	void c_SteamAPI_Shutdown() { SteamAPI_Shutdown(); unregister_callbacks(); }
	bool c_SteamAPI_RestartAppIfNecessary(uint32_t unOwnAppID) {
		return SteamAPI_RestartAppIfNecessary(unOwnAppID);
	}
	void c_SteamAPI_RunCallbacks() { SteamAPI_RunCallbacks(); }

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
