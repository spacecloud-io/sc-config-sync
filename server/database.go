package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spaceuptech/helpers"

	"github.com/spaceuptech/sc-config-sync/admin"
	"github.com/spaceuptech/sc-config-sync/model"
	"github.com/spaceuptech/sc-config-sync/utils"
)

// HandleDatabaseSync syncs database config
func HandleDatabaseSync(a *admin.Module) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Sync function invoked")
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
		defer cancel()

		req := new(model.DBSyncRequest)
		_ = json.NewDecoder(r.Body).Decode(req)

		defer utils.CloseTheCloser(r.Body)

		if req.ProjectID == "" {
			_ = helpers.Response.SendErrorResponse(ctx, w, http.StatusBadRequest, "project id not found in request body")
			return
		}

		token, err := a.CreateToken(ctx, map[string]interface{}{"id": "config-sync", "role": "admin"})
		if err != nil {
			_ = helpers.Response.SendErrorResponse(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		if err := utils.SyncDatabase(ctx, token, req); err != nil {
			_ = helpers.Response.SendErrorResponse(ctx, w, http.StatusInternalServerError, err.Error())
			return
		}

		_ = helpers.Response.SendOkayResponse(ctx, http.StatusOK, w)
	}
}
