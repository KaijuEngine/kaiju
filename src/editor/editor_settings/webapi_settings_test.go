/******************************************************************************/
/* webapi_settings_test.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_settings

import "testing"

func TestNormalizeWebAPIUsesDefaultPortAndGeneratesKey(t *testing.T) {
	settings := Settings{}
	settings.NormalizeWebAPI()
	if settings.WebAPI.Port != 1337 {
		t.Fatalf("port = %d, want 1337", settings.WebAPI.Port)
	}
	if settings.WebAPI.APIKey == "" {
		t.Fatal("expected generated API key")
	}
}

func TestNormalizeWebAPIPreservesValidValues(t *testing.T) {
	settings := Settings{
		WebAPI: WebAPISettings{
			Enabled: true,
			Port:    2020,
			APIKey:  "secret",
		},
	}
	settings.NormalizeWebAPI()
	if !settings.WebAPI.Enabled {
		t.Fatal("enabled should be preserved")
	}
	if settings.WebAPI.Port != 2020 {
		t.Fatalf("port = %d, want 2020", settings.WebAPI.Port)
	}
	if settings.WebAPI.APIKey != "secret" {
		t.Fatalf("APIKey = %q, want secret", settings.WebAPI.APIKey)
	}
}
