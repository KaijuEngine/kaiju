/******************************************************************************/
/* editor_action_webapi_test.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/webapi"
)

func TestActionWebAPIListSearchAndRun(t *testing.T) {
	ed := &Editor{}
	ed.history.Initialize(8)
	handler := actionWebAPI{}

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, webapi.VersionPrefix+"/actions", nil)
	handler.ServeEditorWebAPI(ed, res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", res.Code, http.StatusOK)
	}
	var defs []editor_action.Definition
	if err := json.NewDecoder(res.Body).Decode(&defs); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(defs) == 0 {
		t.Fatalf("list returned no definitions")
	}

	res = httptest.NewRecorder()
	req = jsonRequest(http.MethodPost, webapi.VersionPrefix+"/actions/search", map[string]string{"query": "undo"})
	handler.ServeEditorWebAPI(ed, res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("search status = %d, want %d", res.Code, http.StatusOK)
	}
	var entries []editor_action.Entry
	if err := json.NewDecoder(res.Body).Decode(&entries); err != nil {
		t.Fatalf("decode search: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("search returned no entries")
	}

	res = httptest.NewRecorder()
	req = jsonRequest(http.MethodPost, webapi.VersionPrefix+"/actions/run",
		editor_action.Request{ID: ActionEditorUndo})
	handler.ServeEditorWebAPI(ed, res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("run status = %d, want %d", res.Code, http.StatusOK)
	}
	var result editor_action.Result
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatalf("decode run: %v", err)
	}
	if !result.OK {
		t.Fatalf("run result = %#v, want OK", result)
	}
}

func jsonRequest(method, path string, value any) *http.Request {
	data, _ := json.Marshal(value)
	return httptest.NewRequest(method, path, bytes.NewReader(data))
}
