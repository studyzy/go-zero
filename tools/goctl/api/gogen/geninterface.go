package gogen

import (
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/internal/version"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
)

const (
	interfaceFilename = "interface"
	interfaceTemplate = `package client

import (
	{{.importPackages}}
)

type Client interface {
	{{.methods}}
}
`
	interfaceMethodTemplate = `
{{if .hasDoc}}// {{.function}} {{.doc}}{{end}}
{{.function}}(ctx context.Context,{{.request}}) {{.responseType}}
`
)

func genMethod(gt *template.Template, builder *strings.Builder, route spec.Route) error {
	//{{.function}}({{.request}}) {{.responseType}}
	client := getClientName(route)
	request := ""
	requestType := requestGoTypeName(route, typesPacket)
	if len(requestType) > 0 {
		request = "req *" + requestType
	}
	var responseString string
	if len(route.ResponseTypeName()) > 0 {
		resp := responseGoTypeName(route, typesPacket)
		responseString = "(" + resp + ", error)"
	} else {
		responseString = "error"
	}
	data := map[string]any{
		"client":       strings.Title(client),
		"function":     strings.Title(strings.TrimSuffix(client, "Client")),
		"responseType": responseString,
		"hasRequest":   len(route.RequestTypeName()) > 0,
		"request":      request,
		"hasDoc":       len(route.JoinedDoc()) > 0,
		"doc":          strings.Trim(route.JoinedDoc(), "\""),
	}
	return gt.Execute(builder, data)
}

func genInterface(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	var builder strings.Builder
	templateText, err := pathx.LoadTemplate(category, interfaceMethodTemplateFile, interfaceMethodTemplate)
	if err != nil {
		return err
	}
	gt := template.Must(template.New("groupTemplate").Parse(templateText))

	for _, g := range api.Service.Groups {
		for _, r := range g.Routes {
			err = genMethod(gt, &builder, r)
			if err != nil {
				return err
			}
		}
	}

	var hasTimeout bool

	interfaceFilename, err := format.FileNamingFormat(cfg.NamingFormat, interfaceFilename)
	if err != nil {
		return err
	}

	interfaceFilename = interfaceFilename + ".go"
	filename := path.Join(dir, clientDir, interfaceFilename)
	os.Remove(filename)

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          clientDir,
		filename:        interfaceFilename,
		templateName:    "interfaceTemplate",
		category:        category,
		templateFile:    interfaceTemplateFile,
		builtinTemplate: interfaceTemplate,
		data: map[string]any{
			"hasTimeout":     hasTimeout,
			"importPackages": genInterfaceImports(rootPkg, api),
			"methods":        strings.TrimSpace(builder.String()),
			"version":        version.BuildVersion,
		},
	})
}

func genInterfaceImports(parentPkg string, api *spec.ApiSpec) string {
	var imports []string
	imports = append(imports, "\"context\"")
	imports = append(imports, "")
	imports = append(imports, fmt.Sprintf("\"%s\"\n", pathx.JoinPackages(parentPkg, typesDir)))
	return strings.Join(imports, "\n\t")
}
