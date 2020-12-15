package utils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spaceuptech/helpers"

	"github.com/spaceuptech/sc-config-sync/model"
)

// SyncDatabase syncs the config of database
func SyncDatabase(ctx context.Context, dbAlias, token string, req *model.DBSyncRequest) error {
	specArr := make(model.SpecObjects, 0)
	specObjs, err := GetDbConfig(token, req.ProjectID, "db-configs", map[string]string{"dbAlias": req.From})
	if err != nil {
		return err
	}
	for _, specObject := range specObjs {
		specObject.Meta["dbAlias"] = dbAlias
		specObject.Meta["id"] = fmt.Sprintf("%s-config", dbAlias)

		// override the db config provided in the request
		for key, value := range req.To {
			specObject.Spec.(map[string]interface{})[key] = value
		}
		specArr = append(specArr, specObject)
	}
	helpers.Logger.LogInfo(helpers.GetRequestID(ctx), fmt.Sprintf("Successfully fetched database config having db alias (%s))", req.From), nil)

	specObjs, err = GetDbSchema(token, req.ProjectID, "db-schemas", map[string]string{"dbAlias": req.From})
	if err != nil {
		return err
	}
	for _, specObject := range specObjs {
		specObject.Meta["dbAlias"] = dbAlias
		specArr = append(specArr, specObject)
	}
	helpers.Logger.LogInfo(helpers.GetRequestID(ctx), fmt.Sprintf("Successfully fetched database schemas, total count (%d)))", len(specObjs)), nil)

	specObjs, err = GetDbRule(token, req.ProjectID, "db-rules", map[string]string{"dbAlias": req.From})
	if err != nil {
		return err
	}
	for _, specObject := range specObjs {
		specObject.Meta["dbAlias"] = dbAlias
		specArr = append(specArr, specObject)
	}
	helpers.Logger.LogInfo(helpers.GetRequestID(ctx), fmt.Sprintf("Successfully fetched database rules, total count (%d)))", len(specObjs)), nil)

	specObjs, err = GetDbPreparedQuery(token, req.ProjectID, "db-prepared-query", map[string]string{"dbAlias": req.From})
	if err != nil {
		return err
	}
	for _, specObject := range specObjs {
		specObject.Meta["dbAlias"] = dbAlias
		specObject.Meta["db"] = dbAlias
		specArr = append(specArr, specObject)
	}
	helpers.Logger.LogInfo(helpers.GetRequestID(ctx), fmt.Sprintf("Successfully fetched database prepared query, total count (%d)))", len(specObjs)), nil)

	if err := MakeHTTPRequest(token, http.MethodPost, "/v1/config/batch-apply", map[string]string{}, map[string]interface{}{"specs": specArr}, new(map[string]interface{})); err != nil {
		return err
	}
	return nil
}

// GetDbRule gets database rule
func GetDbRule(token, project, commandName string, params map[string]string) ([]*model.SpecObject, error) {
	url := fmt.Sprintf("/v1/config/projects/%s/database/collections/rules", project)
	// Get the spec from the server
	payload := new(model.Response)
	if err := MakeHTTPRequest(token, http.MethodGet, url, params, new(map[string]interface{}), payload); err != nil {
		return nil, err
	}

	var objs []*model.SpecObject
	for _, item := range payload.Result {
		obj := item.(map[string]interface{})
		col := obj["col"].(string)
		dbAlias := obj["dbAlias"].(string)
		if col == "event_logs" || col == "invocation_logs" {
			continue
		}
		meta := map[string]string{"project": project, "col": col, "dbAlias": dbAlias}

		delete(obj, "col")
		delete(obj, "dbAlias")

		// Generating the object
		s, err := CreateSpecObject("/v1/config/projects/{project}/database/{dbAlias}/collections/{col}/rules", commandName, meta, obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, s)
	}
	return objs, nil
}

// GetDbConfig gets database config
func GetDbConfig(token, project, commandName string, params map[string]string) ([]*model.SpecObject, error) {
	url := fmt.Sprintf("/v1/config/projects/%s/database/config", project)
	// Get the spec from the server
	payload := new(model.Response)
	if err := MakeHTTPRequest(token, http.MethodGet, url, params, new(map[string]interface{}), payload); err != nil {
		return nil, err
	}

	var objs []*model.SpecObject
	for _, item := range payload.Result {
		obj := item.(map[string]interface{})
		dbAlias := obj["dbAlias"].(string)
		configID := fmt.Sprintf("%s-config", dbAlias)
		meta := map[string]string{"project": project, "dbAlias": dbAlias, "id": configID}

		// Delete the unwanted keys from spec
		delete(obj, "id")
		delete(obj, "dbAlias")

		// Generating the object
		s, err := CreateSpecObject("/v1/config/projects/{project}/database/{dbAlias}/config/{id}", commandName, meta, obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, s)

	}
	return objs, nil
}

// GetDbSchema gets database schema
func GetDbSchema(token, project, commandName string, params map[string]string) ([]*model.SpecObject, error) {
	url := fmt.Sprintf("/v1/config/projects/%s/database/collections/schema/mutate", project)

	// Get the spec from the server
	payload := new(model.Response)
	if err := MakeHTTPRequest(token, http.MethodGet, url, params, new(map[string]interface{}), payload); err != nil {
		return nil, err
	}

	var objs []*model.SpecObject
	for _, item := range payload.Result {
		obj := item.(map[string]interface{})
		col := obj["col"].(string)
		dbAlias := obj["dbAlias"].(string)
		if col == "event_logs" || col == "invocation_logs" || col == "default" {
			continue
		}
		meta := map[string]string{"project": project, "col": col, "dbAlias": dbAlias}

		// Generating the object
		s, err := CreateSpecObject("/v1/config/projects/{project}/database/{dbAlias}/collections/{col}/schema/mutate", commandName, meta, map[string]interface{}{"schema": obj["schema"]})
		if err != nil {
			return nil, err
		}
		objs = append(objs, s)
	}
	return objs, nil
}

// GetDbPreparedQuery gets database prepared query
func GetDbPreparedQuery(token, project, commandName string, params map[string]string) ([]*model.SpecObject, error) {
	url := fmt.Sprintf("/v1/config/projects/%s/database/prepared-queries", project)

	payload := new(model.Response)
	if err := MakeHTTPRequest(token, http.MethodGet, url, params, new(map[string]interface{}), payload); err != nil {
		return nil, err
	}

	var objs []*model.SpecObject
	for _, item := range payload.Result {
		obj := item.(map[string]interface{})
		meta := map[string]string{"project": project, "db": obj["dbAlias"].(string), "id": obj["id"].(string)}
		delete(obj, "dbAlias")
		delete(obj, "id")
		s, err := CreateSpecObject("/v1/config/projects/{project}/database/{db}/prepared-queries/{id}", commandName, meta, obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, s)
	}
	return objs, nil
}
