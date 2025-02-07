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

// NewApiClientWithClient returns a http-api client with the given url.
// opts are used to customize the *http.Client.
func NewApiClientWithClient(url string, c *http.Client, opts ...httpc.Option) *ApiClient {
	return &ApiClient{
		url: url,
		cs:  httpc.NewServiceWithClient("{{.client}}", c, opts...),
	}
}

// NewApiClient returns a http-api client with the given url.
// opts are used to customize the *http.Client.
func NewApiClient(url string, opts ...httpc.Option) *ApiClient {
	return &ApiClient{
		url: url,
		cs:  httpc.NewService("{{.client}}", opts...),
	}
}

// Do calls the URL with the given requestBody
func (cli *ApiClient) Do(ctx context.Context, method, url string, requestBody any) ([]byte, error) {
	return doRequest(ctx, method, cli.url+url, requestBody, cli.cs.Do)
}

// call makes an HTTP request and unmarshals the response into the specified type
func call[T any](ctx context.Context, c *ApiClient, httpMethod string, url string, req any) (resp T, err error) {
	result, err := c.Do(ctx, httpMethod, url, req)
	if err != nil {
		logx.Error(err)
		return resp, err
	}
	if err = json.Unmarshal(result, &resp); err != nil {
		return resp, fmt.Errorf("json unmarshal failed. error: %v", err)
	}
	return resp, nil
}

// doRequest performs the HTTP request and handles the response
func doRequest(
	ctx context.Context, method, url string, requestBody any,
	do func(ctx context.Context, method, url string, data any) (*http.Response, error)) ([]byte, error) {

	res, err := do(ctx, method, url, requestBody)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Log the request
	logx.Debugfn(func() any {
		param, _ := json.Marshal(requestBody)
		return fmt.Sprintf("call http api %s [%s] request: %s", method, url, string(param))
	})

	// Check the response status
	if res.StatusCode != http.StatusOK {
		return handleErrorResponse(res)
	}

	return handleSuccessResponse(res)
}

// handleErrorResponse handles non-OK HTTP responses
func handleErrorResponse(res *http.Response) ([]byte, error) {
	var bz []byte
	var err error
	if res.Body != nil {
		bz, err = io.ReadAll(res.Body)
		logx.Error(string(bz), err)
	} else {
		logx.Error("server request failed", res.StatusCode, res.Status)
	}
	return nil, errors.New("server request failed, status: " + res.Status)
}

// handleSuccessResponse handles OK HTTP responses
func handleSuccessResponse(res *http.Response) ([]byte, error) {
	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		logx.Error(err)
		return nil, errors.New("server request failed, status: " + res.Status)
	}
	logx.Debugf("call http api %s [%s] response: %s", res.Request.Method, res.Request.URL, string(responseData))
	return responseData, nil
}

{{.clientMethods}}
