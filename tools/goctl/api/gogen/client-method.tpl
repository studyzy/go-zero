var {{.function}}Url= "{{.url}}"
{{if .hasDoc}}{{.doc}}{{end}}
func (c *ApiClient) {{.function}}(ctx context.Context, {{.request}}) {{.responseType}} {
	result,err:=c.Do(ctx, {{.httpMethod}}, {{.function}}Url{{if .hasRequest}}, req{{else}}, nil{{end}})
	if err!=nil{
	            logx.Error(err)
        return resp,err
	}
    err = json.Unmarshal(result, resp)
    if err != nil {
        return resp, fmt.Errorf("json unmarshal failed. error: %v", err)
    }
    return resp,nil
}
