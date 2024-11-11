package client

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"

	{{.imports}}
	"github.com/zeromicro/go-zero/core/logx"
    "github.com/zeromicro/go-zero/rest/httpc"
)

type ApiClient struct {
    url string
    cs httpc.Service
}


func NewApiClient(url string,c *http.Client) *ApiClient{
	return &ApiClient{
		url: url,
		cs:  httpc.NewServiceWithClient("{{.client}}",c),
	}
}

// Do calls the url with the given requestBody
func (cli *ApiClient) Do(ctx context.Context, method, url string, requestBody any) ([]byte, error) {
	return doRequest(ctx, method, cli.url+url, requestBody, cli.cs.Do)
}
func call[T any](ctx context.Context, c *ApiClient, httpMethod string, url string, req any) (resp T, err error) {
	result, err := c.Do(ctx, httpMethod, url, req)
	if err != nil {
		logx.Error(err)
		return resp, err
	}
	err = json.Unmarshal(result, resp)
	if err != nil {
		return resp, fmt.Errorf("json unmarshal failed. error: %v", err)
	}
	return resp, nil
}
func doRequest(
	ctx context.Context, method, url string, requestBody any,
	do func(ctx context.Context, method, url string, data any) (*http.Response, error),
) ([]byte, error) {

	res, err := do(ctx, method, url, requestBody)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	param, _ := json.Marshal(requestBody)
	logx.Debugf("call http api %s [%s] request: %s", method, url, string(param))

	if res.StatusCode == http.StatusOK {
		responseData, err1 := io.ReadAll(res.Body)
		if err1 != nil {
			logx.Error(err1)
			return nil, errors.New("server request failed, status: " + res.Status)
		}
		logx.Debugf("call http api %s [%s] response: %s", method, url, string(responseData))
		return responseData, nil
	}

	var bz []byte
	if res.Body != nil {
		bz, err = io.ReadAll(res.Body)
		logx.Error(string(bz), err)
		return bz, errors.New("server request failed, status: " + res.Status)
	}
	logx.Error("server request failed", res.StatusCode, res.Status)
	return nil, errors.New("server request failed, status: " + res.Status)
}

{{.clientMethods}}
