// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apis

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	apiproxy "github.com/srinandan/apigeecli/cmd/apis/apiproxydef"
	proxies "github.com/srinandan/apigeecli/cmd/apis/proxies"
	target "github.com/srinandan/apigeecli/cmd/apis/targetendpoint"
)

type pathDetailDef struct {
	OperationID string
	Description string
}

var doc *openapi3.Swagger

func LoadDocumentFromFile(filePath string) (string, []byte, error) {
	var err error
	var jsonContent []byte

	doc, err = openapi3.NewSwaggerLoader().LoadSwaggerFromFile(filePath)
	if err != nil {
		return "", nil, err
	}

	if err = doc.Validate(openapi3.NewSwaggerLoader().Context); err != nil {
		return "", nil, err
	}

	if jsonContent, err = doc.MarshalJSON(); err != nil {
		return "", nil, err
	}

	if isFileYaml(filePath) {
		yamlContent, err := yaml.JSONToYAML(jsonContent)
		return filepath.Base(filePath), yamlContent, err
	} else {
		return filepath.Base(filePath), jsonContent, err
	}
}

func LoadDocumentFromURI(uri string) (string, []byte, error) {
	var err error
	var jsonContent []byte

	u, err := url.Parse(uri)
	if err != nil {
		return "", nil, err
	}

	doc, err = openapi3.NewSwaggerLoader().LoadSwaggerFromURI(u)
	if err != nil {
		return "", nil, err
	}

	if err = doc.Validate(openapi3.NewSwaggerLoader().Context); err != nil {
		return "", nil, err
	}

	if jsonContent, err = doc.MarshalJSON(); err != nil {
		return "", nil, err
	}

	if isFileYaml(uri) {
		yamlContent, err := yaml.JSONToYAML(jsonContent)
		return path.Base(u.Path), yamlContent, err
	} else {
		return path.Base(u.Path), jsonContent, err
	}
}

func isFileYaml(name string) bool {
	if strings.Contains(name, ".yaml") || strings.Contains(name, ".yml") {
		return true
	}
	return false
}

func GenerateAPIProxyDefFromOAS(name string, oasDocName string) (err error) {

	if doc == nil {
		return fmt.Errorf("Open API document not loaded")
	}

	apiproxy.SetDisplayName(name)
	if doc.Info != nil {
		if doc.Info.Description != "" {
			apiproxy.SetDescription(doc.Info.Description)
		}
	}

	apiproxy.SetCreatedAt()
	apiproxy.SetLastModifiedAt()
	apiproxy.SetConfigurationVersion()
	apiproxy.AddTargetEndpoint("default")
	apiproxy.AddProxyEndpoint("default")
	apiproxy.AddResource(oasDocName)
	apiproxy.AddPolicy("Validate-" + name + "-Schema")

	u, err := GetEndpoint(doc)
	if err != nil {
		return err
	}

	apiproxy.SetBasePath(u.Path)

	target.NewTargetEndpoint(u.Scheme + "://" + u.Hostname())

	proxies.NewProxyEndpoint(u.Path)
	proxies.AddStepToPreFlowRequest("OpenAPI-Spec-Validation-1")

	GenerateFlows(doc.Paths)

	return nil
}

func GetEndpoint(doc *openapi3.Swagger) (u *url.URL, err error) {
	if doc.Servers == nil {
		return nil, fmt.Errorf("at least one server must be present")
	}

	return url.Parse(doc.Servers[0].URL)
}

func GetHTTPMethod(pathItem *openapi3.PathItem, keyPath string) map[string]pathDetailDef {

	pathMap := make(map[string]pathDetailDef)
	alternateOperationId := strings.ReplaceAll(keyPath, "\\", "_")

	if pathItem.Get != nil {
		getPathDetail := pathDetailDef{}
		if pathItem.Get.OperationID != "" {
			getPathDetail.OperationID = pathItem.Get.OperationID
		} else {
			getPathDetail.OperationID = "get_" + alternateOperationId
		}
		if pathItem.Get.Description != "" {
			getPathDetail.Description = pathItem.Get.Description
		}
		pathMap["get"] = getPathDetail
	}

	if pathItem.Post != nil {
		postPathDetail := pathDetailDef{}
		if pathItem.Post.OperationID != "" {
			postPathDetail.OperationID = pathItem.Post.OperationID
		} else {
			postPathDetail.OperationID = "post_" + alternateOperationId
		}
		if pathItem.Post.Description != "" {
			postPathDetail.Description = pathItem.Post.Description
		}
		pathMap["post"] = postPathDetail
	}

	if pathItem.Put != nil {
		putPathDetail := pathDetailDef{}
		if pathItem.Put.OperationID != "" {
			putPathDetail.OperationID = pathItem.Put.OperationID
		} else {
			putPathDetail.OperationID = "put_" + alternateOperationId
		}
		if pathItem.Put.Description != "" {
			putPathDetail.Description = pathItem.Put.Description
		}
		pathMap["put"] = putPathDetail
	}

	if pathItem.Patch != nil {
		patchPathDetail := pathDetailDef{}
		if pathItem.Patch.OperationID != "" {
			patchPathDetail.OperationID = pathItem.Patch.OperationID
		} else {
			patchPathDetail.OperationID = "patch_" + alternateOperationId
		}
		if pathItem.Patch.Description != "" {
			patchPathDetail.Description = pathItem.Patch.Description
		}
		pathMap["patch"] = patchPathDetail
	}

	if pathItem.Delete != nil {
		deletePathDetail := pathDetailDef{}
		if pathItem.Delete.OperationID != "" {
			deletePathDetail.OperationID = pathItem.Delete.OperationID
		} else {
			deletePathDetail.OperationID = "delete_" + alternateOperationId
		}
		if pathItem.Delete.Description != "" {
			deletePathDetail.Description = pathItem.Delete.Description
		}
		pathMap["delete"] = deletePathDetail
	}

	if pathItem.Options != nil {
		optionsPathDetail := pathDetailDef{}
		if pathItem.Options.OperationID != "" {
			optionsPathDetail.OperationID = pathItem.Options.OperationID
		} else {
			optionsPathDetail.OperationID = "options_" + alternateOperationId
		}
		if pathItem.Options.Description != "" {
			optionsPathDetail.Description = pathItem.Options.Description
		}
		pathMap["options"] = optionsPathDetail
	}

	if pathItem.Trace != nil {
		tracePathDetail := pathDetailDef{}
		if pathItem.Trace.OperationID != "" {
			tracePathDetail.OperationID = pathItem.Trace.OperationID
		} else {
			tracePathDetail.OperationID = "trace_" + alternateOperationId
		}
		if pathItem.Trace.Description != "" {
			tracePathDetail.Description = pathItem.Trace.Description
		}
		pathMap["trace"] = tracePathDetail
	}

	if pathItem.Head != nil {
		headPathDetail := pathDetailDef{}
		if pathItem.Head.OperationID != "" {
			headPathDetail.OperationID = pathItem.Head.OperationID
		} else {
			headPathDetail.OperationID = "head_" + alternateOperationId
		}
		if pathItem.Head.Description != "" {
			headPathDetail.Description = pathItem.Head.Description
		}
		pathMap["head"] = headPathDetail
	}

	return pathMap
}

func GenerateFlows(paths openapi3.Paths) {
	for keyPath := range paths {
		pathMap := GetHTTPMethod(paths[keyPath], keyPath)
		for method, pathDetail := range pathMap {
			proxies.AddFlow(pathDetail.OperationID, replacePathWithWildCard(keyPath), method, pathDetail.Description)
		}
	}
}

func replacePathWithWildCard(keyPath string) string {
	re := regexp.MustCompile(`{(.*?)}`)
	if strings.ContainsAny(keyPath, "{") {
		return re.ReplaceAllLiteralString(keyPath, "*")
	} else {
		return keyPath
	}
}
