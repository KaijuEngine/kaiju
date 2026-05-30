/******************************************************************************/
/* editor_action_webapi.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/webapi"
)

type actionWebAPI struct{}

type actionSearchRequest struct {
	Query string `json:"query"`
}

type actionAPIRequest struct {
	ID            editor_action.ActionID `json:"id"`
	Params        json.RawMessage        `json:"params,omitempty"`
	Args          json.RawMessage        `json:"args,omitempty"`
	CorrelationID string                 `json:"correlationId,omitempty"`
}

func init() {
	webapi.MustRegister[*Editor](actionWebAPI{})
}

func (actionWebAPI) Routes() []webapi.Route {
	return []webapi.Route{
		{
			Method:      http.MethodGet,
			Path:        "/actions",
			Description: "Lists registered editor actions.",
			Example:     `curl -H "Authorization: Bearer <api-key>" http://127.0.0.1:1337/v1/actions`,
		},
		{
			Method:      http.MethodPost,
			Path:        "/actions/search",
			Description: "Searches runnable editor actions.",
			Example:     `curl -H "Authorization: Bearer <api-key>" -d "{\"query\":\"cube\"}" http://127.0.0.1:1337/v1/actions/search`,
		},
		{
			Method:      http.MethodPost,
			Path:        "/actions/can-run",
			Description: "Checks whether an editor action can run.",
			Example:     `curl -H "Authorization: Bearer <api-key>" -d "{\"id\":\"stage.spawnPrimitive\",\"params\":{\"primitive\":\"cube\"}}" http://127.0.0.1:1337/v1/actions/can-run`,
		},
		{
			Method:      http.MethodPost,
			Path:        "/actions/run",
			Description: "Runs an editor action.",
			Example:     `curl -H "Authorization: Bearer <api-key>" -d "{\"id\":\"stage.spawnPrimitive\",\"params\":{\"primitive\":\"cube\"}}" http://127.0.0.1:1337/v1/actions/run`,
		},
	}
}

func (actionWebAPI) ServeEditorWebAPI(ed *Editor, w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == webapi.VersionPrefix+"/actions":
		writeEditorActionJSON(w, http.StatusOK, ed.Actions().Definitions())
	case r.Method == http.MethodPost && r.URL.Path == webapi.VersionPrefix+"/actions/search":
		var req actionSearchRequest
		if !decodeActionAPIJSON(w, r, &req) {
			return
		}
		writeEditorActionJSON(w, http.StatusOK, ed.Actions().SearchOnMainThread(req.Query))
	case r.Method == http.MethodPost && r.URL.Path == webapi.VersionPrefix+"/actions/can-run":
		req, ok := decodeActionRequest(w, r)
		if !ok {
			return
		}
		req.Source = editor_action.SourceREST
		writeEditorActionJSON(w, http.StatusOK, ed.Actions().CanRunOnMainThread(req))
	case r.Method == http.MethodPost && r.URL.Path == webapi.VersionPrefix+"/actions/run":
		req, ok := decodeActionRequest(w, r)
		if !ok {
			return
		}
		req.Source = editor_action.SourceREST
		writeEditorActionJSON(w, http.StatusOK, ed.Actions().RunOnMainThread(req))
	default:
		http.NotFound(w, r)
	}
}

func decodeActionRequest(w http.ResponseWriter, r *http.Request) (editor_action.Request, bool) {
	var apiReq actionAPIRequest
	if !decodeActionAPIJSON(w, r, &apiReq) {
		return editor_action.Request{}, false
	}
	if apiReq.ID == "" {
		writeEditorActionJSON(w, http.StatusBadRequest, editor_action.Failure("id is required"))
		return editor_action.Request{}, false
	}
	params := apiReq.Params
	if len(params) == 0 {
		params = apiReq.Args
	}
	req := editor_action.Request{
		ID:            apiReq.ID,
		Params:        params,
		CorrelationID: apiReq.CorrelationID,
	}
	return req, true
}

func decodeActionAPIJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		writeEditorActionJSON(w, http.StatusBadRequest, editor_action.Failure(err.Error()))
		return false
	}
	return true
}

func writeEditorActionJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to encode editor action API response", "error", err)
	}
}
