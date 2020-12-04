package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spaceuptech/helpers"

	"github.com/spaceuptech/sc-config-sync/model"
)

// SyncDatabase syncs the config of database
func SyncDatabase(ctx context.Context, token string, req *model.DBSyncRequest) error {
	fmt.Println("Tone", token)
	specArr := make(model.SpecObjects, 0)
	specObjs, err := GetDbConfig(token, req.ProjectID, "db-configs", map[string]string{"dbAlias": req.From})
	if err != nil {
		return err
	}
	for _, specObject := range specObjs {
		specObject.Meta["dbAlias"] = req.To.DbAlias
		specObject.Meta["id"] = fmt.Sprintf("%s-config", req.To.DbAlias)

		specObject.Spec.(map[string]interface{})["name"] = req.To.DBName
		specObject.Spec.(map[string]interface{})["conn"] = req.To.Conn
		specArr = append(specArr, specObject)
	}
	helpers.Logger.LogInfo(helpers.GetRequestID(ctx), fmt.Sprintf("Successfully fetched database config having db alias (%s))", req.From), nil)

	specObjs, err = GetDbSchema(token, req.ProjectID, "db-schemas", map[string]string{"dbAlias": req.From})
	if err != nil {
		return err
	}
	for _, specObject := range specObjs {
		specObject.Meta["dbAlias"] = req.To.DbAlias
		specArr = append(specArr, specObject)
	}
	helpers.Logger.LogInfo(helpers.GetRequestID(ctx), fmt.Sprintf("Successfully fetched database schemas, total count (%d)))", len(specObjs)), nil)

	specObjs, err = GetDbRule(token, req.ProjectID, "db-rules", map[string]string{"dbAlias": req.From})
	if err != nil {
		return err
	}
	for _, specObject := range specObjs {
		specObject.Meta["dbAlias"] = req.To.DbAlias
		specArr = append(specArr, specObject)
	}
	helpers.Logger.LogInfo(helpers.GetRequestID(ctx), fmt.Sprintf("Successfully fetched database rules, total count (%d)))", len(specObjs)), nil)

	specObjs, err = GetDbPreparedQuery(token, req.ProjectID, "db-prepared-query", map[string]string{"dbAlias": req.From})
	if err != nil {
		return err
	}
	for _, specObject := range specObjs {
		specObject.Meta["dbAlias"] = req.To.DbAlias
		specObject.Meta["db"] = req.To.DbAlias
		specArr = append(specArr, specObject)
	}
	helpers.Logger.LogInfo(helpers.GetRequestID(ctx), fmt.Sprintf("Successfully fetched database prepared query, total count (%d)))", len(specObjs)), nil)

	for _, specObject := range specArr {
		if err := ApplySpec(ctx, token, model.GatewayAddr, specObject); err != nil {
			return err
		}
	}
	return nil
}

// ApplySpec takes a spec object and applies it
func ApplySpec(ctx context.Context, token, hostAddr string, specObj *model.SpecObject) error {
	requestBody, err := json.Marshal(specObj.Spec)
	if err != nil {
		return helpers.Logger.LogError(helpers.GetRequestID(ctx), "error while applying service unable to marshal spec", err, nil)
	}
	url, err := adjustPath(fmt.Sprintf("%s%s", hostAddr, specObj.API), specObj.Meta)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return helpers.Logger.LogError(helpers.GetRequestID(ctx), "error while applying service unable to send http request", err, nil)
	}

	v := map[string]interface{}{}
	_ = json.NewDecoder(resp.Body).Decode(&v)
	CloseTheCloser(req.Body)

	if resp.StatusCode == http.StatusAccepted {
		// Make checker send this status
		helpers.Logger.LogInfo(helpers.GetRequestID(ctx), fmt.Sprintf("Successfully queued %s", specObj.Type), nil)
	} else if resp.StatusCode == http.StatusOK {
		helpers.Logger.LogInfo(helpers.GetRequestID(ctx), fmt.Sprintf("Successfully applied %s", specObj.Type), nil)
	} else {
		_ = helpers.Logger.LogError(helpers.GetRequestID(ctx), fmt.Sprintf("error while applying service got http status code %s", resp.Status), fmt.Errorf("%s", v["error"]), nil)
		return fmt.Errorf("%v", v["error"])
	}
	return nil
}

func adjustPath(path string, meta map[string]string) (string, error) {
	newPath := path
	for {
		pre := strings.IndexRune(newPath, '{')
		if pre < 0 {
			return newPath, nil
		}
		post := strings.IndexRune(newPath, '}')

		key := strings.TrimSuffix(strings.TrimPrefix(newPath[pre:post], "{"), "}")
		value, p := meta[key]
		if !p {
			return "", fmt.Errorf("provided key (%s) does not exist in metadata", key)
		}

		newPath = newPath[:pre] + value + newPath[post+1:]
	}
}

// GetDbRule gets database rule
func GetDbRule(token, project, commandName string, params map[string]string) ([]*model.SpecObject, error) {
	url := fmt.Sprintf("/v1/config/projects/%s/database/collections/rules", project)
	// Get the spec from the server
	payload := new(model.Response)
	if err := MakeHTTPRequest(token, http.MethodGet, url, params, payload); err != nil {
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
	if err := MakeHTTPRequest(token, http.MethodGet, url, params, payload); err != nil {
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
	if err := MakeHTTPRequest(token, http.MethodGet, url, params, payload); err != nil {
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
	if err := MakeHTTPRequest(token, http.MethodGet, url, params, payload); err != nil {
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
