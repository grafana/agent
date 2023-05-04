// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package injectproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/efficientgo/tools/core/pkg/merrors"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

const (
	queryParam    = "query"
	matchersParam = "match[]"
)

type routes struct {
	upstream *url.URL
	handler  http.Handler
	label    string

	mux            *http.ServeMux
	modifiers      map[string]func(*http.Response) error
	errorOnReplace bool
}

type options struct {
	enableLabelAPIs  bool
	passthroughPaths []string
	errorOnReplace   bool
}

type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

// WithEnabledLabelsAPI enables proxying to labels API. If false, "501 Not implemented" will be return for those.
func WithEnabledLabelsAPI() Option {
	return optionFunc(func(o *options) {
		o.enableLabelAPIs = true
	})
}

// WithPassthroughPaths configures routes to register given paths as passthrough handlers for all HTTP methods.
// that, if requested, will be forwarded without enforcing label. Use with care.
// NOTE: Passthrough "all" paths like "/" or "" and regex are not allowed.
func WithPassthroughPaths(paths []string) Option {
	return optionFunc(func(o *options) {
		o.passthroughPaths = paths
	})
}

// WithErrorOnReplace causes the proxy to return 400 if a label matcher we want to
// inject is present in the query already and matches something different
func WithErrorOnReplace() Option {
	return optionFunc(func(o *options) {
		o.errorOnReplace = true
	})
}

// strictMux is a mux that wraps standard HTTP handler with safer handler that allows safe user provided handler registrations.
type strictMux struct {
	seen map[string]struct{}

	m *http.ServeMux
}

func newStrictMux() *strictMux {
	return &strictMux{
		seen: map[string]struct{}{},
		m:    http.NewServeMux(),
	}

}

// Handle is like HTTP mux handle but it does not allow to register paths that are shared with previously registered paths.
// It also makes sure the trailing / is registered too.
// For example if /api/v1/federate was registered consequent registrations like /api/v1/federate/ or /api/v1/federate/some will
// return error. In the mean time request with both /api/v1/federate and /api/v1/federate/ will point to the handled passed by /api/v1/federate
// registration.
// This allows to de-risk ability for user to mis-configure and leak inject isolation.
func (s *strictMux) Handle(pattern string, handler http.Handler) error {
	sanitized := pattern
	for next := strings.TrimSuffix(sanitized, "/"); next != sanitized; sanitized = next {
	}

	if _, ok := s.seen[sanitized]; ok {
		return errors.Errorf("pattern %q was already registered", sanitized)
	}

	for p := range s.seen {
		if strings.HasPrefix(sanitized+"/", p+"/") {
			return errors.Errorf("pattern %q is registered, cannot register path %q that shares it", p, sanitized)
		}
	}

	s.m.Handle(sanitized, handler)
	s.m.Handle(sanitized+"/", handler)
	s.seen[sanitized] = struct{}{}

	return nil
}

func NewRoutes(upstream *url.URL, label string, opts ...Option) (*routes, error) {
	opt := options{}
	for _, o := range opts {
		o.apply(&opt)
	}

	proxy := httputil.NewSingleHostReverseProxy(upstream)

	r := &routes{upstream: upstream, handler: proxy, label: label, errorOnReplace: opt.errorOnReplace}
	mux := newStrictMux()

	errs := merrors.New(
		mux.Handle("/federate", r.enforceLabel(enforceMethods(r.matcher, "GET"))),
		mux.Handle("/api/v1/query", r.enforceLabel(enforceMethods(r.query, "GET", "POST"))),
		mux.Handle("/api/v1/query_range", r.enforceLabel(enforceMethods(r.query, "GET", "POST"))),
		mux.Handle("/api/v1/alerts", r.enforceLabel(enforceMethods(r.passthrough, "GET"))),
		mux.Handle("/api/v1/rules", r.enforceLabel(enforceMethods(r.passthrough, "GET"))),
		mux.Handle("/api/v1/series", r.enforceLabel(enforceMethods(r.matcher, "GET", "POST"))),
		mux.Handle("/api/v1/query_exemplars", r.enforceLabel(enforceMethods(r.query, "GET", "POST"))),
	)

	if opt.enableLabelAPIs {
		errs.Add(
			mux.Handle("/api/v1/labels", r.enforceLabel(enforceMethods(r.matcher, "GET", "POST"))),
			// Full path is /api/v1/label/<label_name>/values but http mux does not support patterns.
			// This is fine though as we don't care about name for matcher injector.
			mux.Handle("/api/v1/label/", r.enforceLabel(enforceMethods(r.matcher, "GET"))),
		)
	}

	errs.Add(
		mux.Handle("/api/v2/silences", r.enforceLabel(enforceMethods(r.silences, "GET", "POST"))),
		mux.Handle("/api/v2/silence/", r.enforceLabel(enforceMethods(r.deleteSilence, "DELETE"))),
		mux.Handle("/api/v2/alerts/groups", r.enforceLabel(enforceMethods(r.enforceFilterParameter, "GET"))),
	)

	errs.Add(
		mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		})),
	)

	if err := errs.Err(); err != nil {
		return nil, err
	}

	// Validate paths.
	for _, path := range opt.passthroughPaths {
		u, err := url.Parse(fmt.Sprintf("http://example.com%v", path))
		if err != nil {
			return nil, fmt.Errorf("path %q is not a valid URI path, got %v", path, opt.passthroughPaths)
		}
		if u.Path != path {
			return nil, fmt.Errorf("path %q is not a valid URI path, got %v", path, opt.passthroughPaths)
		}
		if u.Path == "" || u.Path == "/" {
			return nil, fmt.Errorf("path %q is not allowed, got %v", u.Path, opt.passthroughPaths)
		}
	}

	// Register optional passthrough paths.
	for _, path := range opt.passthroughPaths {
		if err := mux.Handle(path, http.HandlerFunc(r.passthrough)); err != nil {
			return nil, err
		}
	}

	r.mux = mux.m
	r.modifiers = map[string]func(*http.Response) error{
		"/api/v1/rules":  modifyAPIResponse(r.filterRules),
		"/api/v1/alerts": modifyAPIResponse(r.filterAlerts),
	}
	proxy.ModifyResponse = r.ModifyResponse
	return r, nil
}

func (r *routes) enforceLabel(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		lvalue := req.FormValue(r.label)
		if lvalue == "" {
			http.Error(w, fmt.Sprintf("Bad request. The %q query parameter must be provided.", r.label), http.StatusBadRequest)
			return
		}
		req = req.WithContext(withLabelValue(req.Context(), lvalue))

		// Remove the proxy label from the query parameters.
		q := req.URL.Query()
		if q.Get(r.label) != "" {
			q.Del(r.label)
		}
		req.URL.RawQuery = q.Encode()
		// Remove the proxy label from the PostForm.
		if req.Method == http.MethodPost {
			if err := req.ParseForm(); err != nil {
				http.Error(w, fmt.Sprintf("Failed to parse the PostForm: %v", err), http.StatusInternalServerError)
				return
			}
			if req.PostForm.Get(r.label) != "" {
				req.PostForm.Del(r.label)
				newBody := req.PostForm.Encode()
				// We are replacing request body, close previous one (req.FormValue ensures it is read fully and not nil).
				_ = req.Body.Close()
				req.Body = ioutil.NopCloser(strings.NewReader(newBody))
				req.ContentLength = int64(len(newBody))
			}
		}

		h.ServeHTTP(w, req)
	})
}

func (r *routes) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *routes) ModifyResponse(resp *http.Response) error {
	m, found := r.modifiers[resp.Request.URL.Path]
	if !found {
		// Return the server's response unmodified.
		return nil
	}
	return m(resp)
}

func enforceMethods(h http.HandlerFunc, methods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for _, m := range methods {
			if m == req.Method {
				h(w, req)
				return
			}
		}
		http.NotFound(w, req)
	}
}

type ctxKey int

const keyLabel ctxKey = iota

func mustLabelValue(ctx context.Context) string {
	label, ok := ctx.Value(keyLabel).(string)
	if !ok {
		panic(fmt.Sprintf("can't find the %q value in the context", keyLabel))
	}
	if label == "" {
		panic(fmt.Sprintf("empty %q value in the context", keyLabel))
	}
	return label
}

func withLabelValue(ctx context.Context, label string) context.Context {
	return context.WithValue(ctx, keyLabel, label)
}

func (r *routes) passthrough(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func (r *routes) query(w http.ResponseWriter, req *http.Request) {
	e := NewEnforcer(r.errorOnReplace,
		[]*labels.Matcher{{
			Name:  r.label,
			Type:  labels.MatchEqual,
			Value: mustLabelValue(req.Context()),
		}}...)

	// The `query` can come in the URL query string and/or the POST body.
	// For this reason, we need to try to enforcing in both places.
	// Note: a POST request may include some values in the URL query string
	// and others in the body. If both locations include a `query`, then
	// enforce in both places.
	q, found1, err := enforceQueryValues(e, req.URL.Query())
	if err != nil {
		switch err.(type) {
		case IllegalLabelMatcherError:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case queryParseError:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case enforceLabelError:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	req.URL.RawQuery = q

	var found2 bool
	// Enforce the query in the POST body if needed.
	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		q, found2, err = enforceQueryValues(e, req.PostForm)
		if err != nil {
			switch err.(type) {
			case IllegalLabelMatcherError:
				http.Error(w, err.Error(), http.StatusBadRequest)
			case queryParseError:
				http.Error(w, err.Error(), http.StatusBadRequest)
			case enforceLabelError:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// We are replacing request body, close previous one (ParseForm ensures it is read fully and not nil).
		_ = req.Body.Close()
		req.Body = ioutil.NopCloser(strings.NewReader(q))
		req.ContentLength = int64(len(q))
	}

	// If no query was found, return early.
	if !found1 && !found2 {
		return
	}

	r.handler.ServeHTTP(w, req)
}

func enforceQueryValues(e *Enforcer, v url.Values) (values string, noQuery bool, err error) {
	// If no values were given or no query is present,
	// e.g. because the query came in the POST body
	// but the URL query string was passed, then finish early.
	if v.Get(queryParam) == "" {
		return v.Encode(), false, nil
	}
	expr, err := parser.ParseExpr(v.Get(queryParam))
	if err != nil {
		queryParseError := newQueryParseError(err)
		return "", true, queryParseError
	}

	if err := e.EnforceNode(expr); err != nil {
		if _, ok := err.(IllegalLabelMatcherError); ok {
			return "", true, err
		}
		enforceLabelError := newEnforceLabelError(err)
		return "", true, enforceLabelError
	}

	v.Set(queryParam, expr.String())
	return v.Encode(), true, nil
}

// matcher ensures all the provided match[] if any has label injected. If none was provided, single matcher is injected.
// This works for non-query Prometheus APIs like: /api/v1/series, /api/v1/label/<name>/values, /api/v1/labels and /federate support multiple matchers.
// See e.g https://prometheus.io/docs/prometheus/latest/querying/api/#querying-metadata
func (r *routes) matcher(w http.ResponseWriter, req *http.Request) {
	matcher := &labels.Matcher{
		Name:  r.label,
		Type:  labels.MatchEqual,
		Value: mustLabelValue(req.Context()),
	}
	q := req.URL.Query()

	if err := injectMatcher(q, matcher); err != nil {
		return
	}
	req.URL.RawQuery = q.Encode()
	if req.Method == http.MethodPost {
		if err := req.ParseForm(); err != nil {
			return
		}
		q = req.PostForm
		if err := injectMatcher(q, matcher); err != nil {
			return
		}
		// We are replacing request body, close previous one (ParseForm ensures it is read fully and not nil).
		_ = req.Body.Close()
		newBody := q.Encode()
		req.Body = ioutil.NopCloser(strings.NewReader(newBody))
		req.ContentLength = int64(len(newBody))
	}
	r.handler.ServeHTTP(w, req)
}

func injectMatcher(q url.Values, matcher *labels.Matcher) error {
	matchers := q[matchersParam]
	if len(matchers) == 0 {
		q.Set(matchersParam, matchersToString(matcher))
	} else {
		// Inject label to existing matchers.
		for i, m := range matchers {
			ms, err := parser.ParseMetricSelector(m)
			if err != nil {
				return err
			}
			matchers[i] = matchersToString(append(ms, matcher)...)
		}
		q[matchersParam] = matchers
	}
	return nil
}

func matchersToString(ms ...*labels.Matcher) string {
	var el []string
	for _, m := range ms {
		el = append(el, m.String())
	}
	return fmt.Sprintf("{%v}", strings.Join(el, ","))
}

type queryParseError struct {
	msg string
}

func (e queryParseError) Error() string {
	return e.msg
}

func newQueryParseError(err error) queryParseError {
	return queryParseError{msg: fmt.Sprintf("error parsing query string %q", err.Error())}
}

type enforceLabelError struct {
	msg string
}

func (e enforceLabelError) Error() string {
	return e.msg
}

func newEnforceLabelError(err error) enforceLabelError {
	return enforceLabelError{msg: fmt.Sprintf("error enforcing label %q", err.Error())}
}
