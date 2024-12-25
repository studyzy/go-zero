package gogen

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/vars"
)

//go:embed client.tpl
var clientTemplate string

//go:embed client-method.tpl
var clientMethodTemplate string

func genClient(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	// load client method template
	templateText, err := pathx.LoadTemplate(category, clientMethodTemplateFile, clientMethodTemplate)
	if err != nil {
		return err
	}
	t := template.New("groupTemplate")
	// Add all functions from strings package to the template
	funcMap := template.FuncMap{
		"Contains":  strings.Contains,
		"Count":     strings.Count,
		"HasPrefix": strings.HasPrefix,
		"HasSuffix": strings.HasSuffix,
		"Index":     strings.Index,
		"Join":      strings.Join,
		"Repeat":    strings.Repeat,
		"Replace":   strings.Replace,
		"Split":     strings.Split,
		"ToLower":   strings.ToLower,
		"ToUpper":   strings.ToUpper,
		"Trim":      strings.Trim,
		"TrimSpace": strings.TrimSpace,
		// Add other strings functions as needed
	}
	t.Funcs(funcMap)
	gt := template.Must(t.Parse(templateText))

	// generate client method
	var builder strings.Builder
	for _, g := range api.Service.Groups {
		for _, r := range g.Routes {
			if err := generateClientMethod(gt, &builder, g, r); err != nil {
				return err
			}
		}
	}
	// generate client file
	return genClientFile(&builder, dir, rootPkg, api.Service.Name)
}

func generateClientMethod(gt *template.Template, builder *strings.Builder, g spec.Group, r spec.Route) error {
	client := getClientName(r)
	var responseString, responseType, returnString, requestString string
	if len(r.ResponseTypeName()) > 0 {
		responseType = responseGoTypeName(r, typesPacket)
		responseString = "(resp " + responseType + ", err error)"
		returnString = "return"
	} else {
		responseString = "error"
		returnString = "return nil"
	}
	if len(r.RequestTypeName()) > 0 {
		requestString = "req *" + requestGoTypeName(r, typesPacket)
	}
	data := map[string]any{
		"client":         strings.Title(client),
		"function":       strings.Title(strings.TrimSuffix(client, "Client")),
		"responseString": responseString,
		"responseType":   responseType,
		"httpMethod":     mapping[r.Method],
		"hasRequest":     len(r.RequestTypeName()) > 0,
		"hasResponse":    len(r.ResponseTypeName()) > 0,
		"returnString":   returnString,
		"request":        requestString,
		"hasDoc":         len(r.JoinedDoc()) > 0,
		"doc":            strings.Trim(r.JoinedDoc(), "\""),
		"url":            g.Annotation.Properties["prefix"] + r.Path,
	}

	return gt.Execute(builder, data)
}

func genClientFile(builder *strings.Builder, dir, rootPkg string, name string) error {
	imports := genClientImports(rootPkg)
	subDir := clientDir
	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          subDir,
		filename:        "client.go",
		templateName:    "clientTemplate",
		category:        category,
		templateFile:    clientTemplateFile,
		builtinTemplate: clientTemplate,
		data: map[string]any{
			"imports":       imports,
			"client":        strings.Title(name),
			"clientMethods": builder.String(),
		},
	})
}

func genClientImports(parentPkg string) string {
	var imports []string
	imports = append(imports, fmt.Sprintf("\"%s\"\n", pathx.JoinPackages(parentPkg, typesDir)))
	imports = append(imports, fmt.Sprintf("\"%s/core/logx\"", vars.ProjectOpenSourceURL))
	return strings.Join(imports, "\n\t")
}

func getClientName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	return handler + "Client"
}
