{{if .hasDoc}}// {{.function}} {{.doc}}{{end}}
func (c *ApiClient) {{.function}}(ctx context.Context, {{.request}}) {{.responseString}} {
    const {{.function}}Url= "{{.url}}"
    return call[{{.responseType}}](ctx, c, {{.httpMethod}}, {{.function}}Url{{if .hasRequest}}, req{{else}}, nil{{end}})
}
