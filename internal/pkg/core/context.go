package core

import (
	"bytes"
	stdctx "context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	"edu-schedule-system/internal/code"
	"edu-schedule-system/internal/pkg/trace"
	"edu-schedule-system/internal/pkg/validation"
	"edu-schedule-system/internal/proposal"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

type HandlerFunc func(c Context)

type Trace = trace.T

const (
	_Alias           = "_alias_"
	_TraceName       = "_trace_"
	_LoggerName      = "_logger_"
	_BodyName        = "_body_"
	_PayloadName     = "_payload_"
	_HTTPStatusName  = "_http_status_"
	_SessionUserInfo = "_session_user_info"
	_AbortErrorName  = "_abort_error_"
	_IsRecordMetrics = "_is_record_metrics_"
)

type actorContextKey struct{}

var contextPool = &sync.Pool{
	New: func() interface{} {
		return new(context)
	},
}

func newContext(ctx *gin.Context) Context {
	context := contextPool.Get().(*context)
	context.ctx = ctx
	return context
}

func releaseContext(ctx Context) {
	c := ctx.(*context)
	c.ctx = nil
	contextPool.Put(c)
}

var _ Context = (*context)(nil)

type Context interface {
	init()

	// ShouldBindQuery deserialize querystring
	// tag: `form:"xxx"`
	ShouldBindQuery(obj interface{}) error

	// ShouldBindPostForm deserialize postform (querystring is ignored)
	// tag: `form:"xxx"`
	ShouldBindPostForm(obj interface{}) error

	// ShouldBindForm deserialize querystring and postform
	// tag: `form:"xxx"`
	ShouldBindForm(obj interface{}) error

	// ShouldBindJSON deserialize postjson
	// tag: `json:"xxx"`
	ShouldBindJSON(obj interface{}) error

	// ShouldBindURI deserialize path params
	// tag: `uri:"xxx"`
	ShouldBindURI(obj interface{}) error

	// Redirect redirect
	Redirect(code int, location string)

	// Trace get Trace
	Trace() Trace
	setTrace(trace Trace)
	disableTrace()

	// Logger get Logger
	Logger() *zap.Logger
	setLogger(logger *zap.Logger)

	// getPayload get payload
	getPayload() interface{}

	// Payload success response
	Payload(payload interface{})
	PayloadWithStatus(status int, payload interface{})
	responseStatus() int
	HTML(name string, obj interface{})
	File(filePath string)
	FormFile(name string) (*multipart.FileHeader, error)
	SaveUploadedFile(file *multipart.FileHeader, dst string) error

	Header() http.Header
	GetHeader(key string) string
	SetHeader(key, value string)

	SessionUserInfo() proposal.SessionUserInfo
	setSessionUserInfo(info proposal.SessionUserInfo)

	// AbortWithError error response
	AbortWithError(err BusinessError)
	abortError() BusinessError

	Alias() string
	setAlias(path string)

	isRecordMetrics() bool
	ableRecordMetrics()
	disableRecordMetrics()

	RequestInputParams() url.Values
	RequestPostFormParams() url.Values
	RequestPathParams(key string) string

	Request() *http.Request
	RawData() []byte
	Method() string
	Host() string
	Path() string
	RoutePath() string
	URI() string

	RequestContext() StdContext
	ResponseWriter() gin.ResponseWriter
	Param(key string) string
}

type context struct {
	ctx *gin.Context
}

type StdContext struct {
	stdctx.Context
	Trace
	*zap.Logger
}

// WithActor returns a child context carrying the authenticated session actor.
func WithActor(ctx stdctx.Context, actor proposal.SessionUserInfo) stdctx.Context {
	return stdctx.WithValue(ctx, actorContextKey{}, actor)
}

// ActorFromContext extracts the authenticated session actor from a context.
func ActorFromContext(ctx stdctx.Context) proposal.SessionUserInfo {
	if ctx == nil {
		return proposal.SessionUserInfo{}
	}
	actor, _ := ctx.Value(actorContextKey{}).(proposal.SessionUserInfo)
	return actor
}

func (c StdContext) TraceValue() trace.T {
	return c.Trace
}

func (c *context) init() {
	if !traceBodyEnabled() || !shouldCaptureBody(c.ctx.Request) {
		return
	}

	body, err := c.ctx.GetRawData()
	if err != nil {
		panic(err)
	}

	c.ctx.Set(_BodyName, body)                               // cache body for trace use
	c.ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // re-construct req body
}

func shouldCaptureBody(req *http.Request) bool {
	if req == nil {
		return false
	}
	if req.ContentLength <= 0 {
		return false
	}
	maxBytes := traceBodyMaxBytes()
	if maxBytes > 0 && req.ContentLength > maxBytes {
		return false
	}

	ct := strings.ToLower(strings.TrimSpace(req.Header.Get("Content-Type")))
	if strings.HasPrefix(ct, "multipart/") || strings.HasPrefix(ct, "application/octet-stream") {
		return false
	}

	return true
}

// ShouldBindQuery deserialize querystring
// tag: `form:"xxx"`
func (c *context) ShouldBindQuery(obj interface{}) error {
	err := c.ctx.ShouldBindWith(obj, binding.Query)
	if err != nil {
		return c.handleValidationError(err)
	}
	return nil
}

// ShouldBindPostForm deserialize postform (querystring is ignored)
// tag: `form:"xxx"`
func (c *context) ShouldBindPostForm(obj interface{}) error {
	err := c.ctx.ShouldBindWith(obj, binding.FormPost)
	if err != nil {
		return c.handleValidationError(err)
	}
	return nil
}

// ShouldBindForm deserialize querystring and postform
// tag: `form:"xxx"`
func (c *context) ShouldBindForm(obj interface{}) error {
	err := c.ctx.ShouldBindWith(obj, binding.Form)
	if err != nil {
		return c.handleValidationError(err)
	}
	return nil
}

// ShouldBindJSON deserialize postjson
// tag: `json:"xxx"`
func (c *context) ShouldBindJSON(obj interface{}) error {
	err := c.ctx.ShouldBindWith(obj, binding.JSON)
	if err != nil {
		return c.handleValidationError(err)
	}
	return nil
}

// ShouldBindURI deserialize path params
// tag: `uri:"xxx"`
func (c *context) ShouldBindURI(obj interface{}) error {
	err := c.ctx.ShouldBindUri(obj)
	if err != nil {
		return c.handleValidationError(err)
	}
	return nil
}

func (c *context) handleValidationError(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(validation.Error(err))
}

// Redirect redirect
func (c *context) Redirect(code int, location string) {
	c.ctx.Redirect(code, location)
}

func (c *context) Trace() Trace {
	t, ok := c.ctx.Get(_TraceName)
	if !ok || t == nil {
		return nil
	}

	return t.(Trace)
}

func (c *context) setTrace(trace Trace) {
	c.ctx.Set(_TraceName, trace)
}

func (c *context) disableTrace() {
	c.setTrace(nil)
}

func (c *context) Logger() *zap.Logger {
	logger, ok := c.ctx.Get(_LoggerName)
	if !ok {
		return nil
	}

	return logger.(*zap.Logger)
}

func (c *context) setLogger(logger *zap.Logger) {
	c.ctx.Set(_LoggerName, logger)
}

func (c *context) getPayload() interface{} {
	if payload, ok := c.ctx.Get(_PayloadName); ok {
		return payload
	}
	return nil
}

func (c *context) Payload(payload interface{}) {
	c.setPayload(http.StatusOK, payload)
}

func (c *context) PayloadWithStatus(status int, payload interface{}) {
	c.setPayload(status, payload)
}

func (c *context) setPayload(status int, payload interface{}) {
	traceID := ""
	if t := c.Trace(); t != nil {
		traceID = t.ID()
	}

	response := &code.Response{
		Code:    0, // 0 means success
		Message: "success",
		Data:    payload,
		TraceID: traceID,
	}
	c.ctx.Set(_HTTPStatusName, status)
	c.ctx.Set(_PayloadName, response)
}

func (c *context) responseStatus() int {
	status, ok := c.ctx.Get(_HTTPStatusName)
	if !ok {
		return http.StatusOK
	}
	statusInt, ok := status.(int)
	if !ok || statusInt < 100 {
		return http.StatusOK
	}
	return statusInt
}

func (c *context) File(filePath string) {
	c.ctx.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", path.Base(filePath)))
	c.ctx.Writer.Header().Add("Content-Type", "application/octet-stream")
	c.ctx.File(filePath)
}

func (c *context) HTML(name string, obj interface{}) {
	c.ctx.HTML(200, name+".html", obj)
}

func (c *context) FormFile(name string) (*multipart.FileHeader, error) {
	return c.ctx.FormFile(name)
}

func (c *context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	return c.ctx.SaveUploadedFile(file, dst)
}

func (c *context) Header() http.Header {
	header := c.ctx.Request.Header

	clone := make(http.Header, len(header))
	for k, v := range header {
		value := make([]string, len(v))
		copy(value, v)

		clone[k] = value
	}
	return clone
}

func (c *context) GetHeader(key string) string {
	return c.ctx.GetHeader(key)
}

func (c *context) SetHeader(key, value string) {
	c.ctx.Header(key, value)
}

func (c *context) SessionUserInfo() proposal.SessionUserInfo {
	val, ok := c.ctx.Get(_SessionUserInfo)
	if !ok {
		return proposal.SessionUserInfo{}
	}

	return val.(proposal.SessionUserInfo)
}

func (c *context) setSessionUserInfo(info proposal.SessionUserInfo) {
	c.ctx.Set(_SessionUserInfo, info)
}

func (c *context) AbortWithError(err BusinessError) {
	if err != nil {
		httpCode := err.HTTPCode()
		if httpCode == 0 {
			httpCode = http.StatusInternalServerError
		}

		c.ctx.AbortWithStatus(httpCode)
		c.ctx.Set(_AbortErrorName, err)
	}
}

func (c *context) abortError() BusinessError {
	err, ok := c.ctx.Get(_AbortErrorName)
	if !ok || err == nil {
		return nil
	}
	if be, ok := err.(BusinessError); ok {
		return be
	}
	return nil
}

func (c *context) Alias() string {
	path, ok := c.ctx.Get(_Alias)
	if !ok {
		return ""
	}

	return path.(string)
}

func (c *context) setAlias(path string) {
	if path = strings.TrimSpace(path); path != "" {
		c.ctx.Set(_Alias, path)
	}
}

func (c *context) isRecordMetrics() bool {
	isRecordMetrics, ok := c.ctx.Get(_IsRecordMetrics)
	if !ok {
		return false
	}

	return isRecordMetrics.(bool)
}

func (c *context) ableRecordMetrics() {
	c.ctx.Set(_IsRecordMetrics, true)
}

func (c *context) disableRecordMetrics() {
	c.ctx.Set(_IsRecordMetrics, false)
}

// RequestInputParams returns all params
func (c *context) RequestInputParams() url.Values {
	_ = c.ctx.Request.ParseForm()
	return c.ctx.Request.Form
}

// RequestPostFormParams returns post form params
func (c *context) RequestPostFormParams() url.Values {
	_ = c.ctx.Request.ParseForm()
	return c.ctx.Request.PostForm
}

// RequestPathParams returns path param
func (c *context) RequestPathParams(key string) string {
	return c.ctx.Param(key)
}

// Request returns request
func (c *context) Request() *http.Request {
	return c.ctx.Request
}

func (c *context) RawData() []byte {
	body, ok := c.ctx.Get(_BodyName)
	if !ok {
		return nil
	}

	return body.([]byte)
}

// Method returns request method
func (c *context) Method() string {
	return c.ctx.Request.Method
}

// Host returns request host
func (c *context) Host() string {
	return c.ctx.Request.Host
}

// Path returns request path without querystring
func (c *context) Path() string {
	return c.ctx.Request.URL.Path
}

func (c *context) RoutePath() string {
	return c.ctx.FullPath()
}

// URI returns unescaped uri
func (c *context) URI() string {
	uri, _ := url.QueryUnescape(c.ctx.Request.URL.RequestURI())
	return uri
}

// RequestContext wraps Trace + Logger
func (c *context) RequestContext() StdContext {
	base := c.ctx.Request.Context()
	if actor := c.SessionUserInfo(); actor.Id > 0 {
		base = WithActor(base, actor)
	}
	return StdContext{
		base,
		c.Trace(),
		c.Logger(),
	}
}

// ResponseWriter returns response writer
func (c *context) ResponseWriter() gin.ResponseWriter {
	return c.ctx.Writer
}

// Param returns URL param
func (c *context) Param(key string) string {
	return c.ctx.Param(key)
}
