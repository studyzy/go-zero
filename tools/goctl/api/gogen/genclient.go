package gogen

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/vars"
)

//go:embed client.tpl
var clientTemplate string

//go:embed client-method.tpl
var clientMethodTemplate string

func genClient(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	var builder strings.Builder
	templateText, err := pathx.LoadTemplate(category, clientMethodTemplateFile, clientMethodTemplate)
	if err != nil {
		return err
	}

	gt := template.Must(template.New("groupTemplate").Parse(templateText))
	for _, g := range api.Service.Groups {
		for _, r := range g.Routes {
			client := getClientName(r)
			var responseString string
			var returnString string
			var requestString string
			if len(r.ResponseTypeName()) > 0 {
				resp := responseGoTypeName(r, typesPacket)
				responseString = "(resp " + resp + ", err error)"
				returnString = "return"
			} else {
				responseString = "error"
				returnString = "return nil"
			}
			if len(r.RequestTypeName()) > 0 {
				requestString = "req *" + requestGoTypeName(r, typesPacket)
			}

			data := map[string]any{
				"client":       strings.Title(client),
				"function":     strings.Title(strings.TrimSuffix(client, "Client")),
				"responseType": responseString,
				"httpMethod":   mapping[r.Method],
				"hasRequest":   len(r.RequestTypeName()) > 0,
				"returnString": returnString,
				"request":      requestString,
				"hasDoc":       len(r.JoinedDoc()) > 0,
				"doc":          getDoc(r.JoinedDoc()),
				"url":          g.Annotation.Properties["prefix"] + r.Path,
			}
			if err := gt.Execute(&builder, data); err != nil {
				return err
			}
		}
	}
	// generate 1 client.go file

	return genClientFile(&builder, dir, rootPkg, cfg)
}

func genClientFile(builder *strings.Builder, dir, rootPkg string, cfg *config.Config) error {
	client := "Client"
	goFile, err := format.FileNamingFormat(cfg.NamingFormat, client)
	if err != nil {
		return err
	}

	imports := genClientImports(rootPkg)
	subDir := clientDir
	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          subDir,
		filename:        goFile + ".go",
		templateName:    "clientTemplate",
		category:        category,
		templateFile:    clientTemplateFile,
		builtinTemplate: clientTemplate,
		data: map[string]any{
			"imports":       imports,
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
