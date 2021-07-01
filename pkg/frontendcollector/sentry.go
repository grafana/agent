package frontendcollector

import (
	"fmt"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type LogContext = map[string]interface{}

type Measurement struct {
	Value float32 `json:"value"`
}

type Measurements = map[string]Measurement

type FrontendSentryExceptionValue struct {
	Value      string            `json:"value,omitempty"`
	Type       string            `json:"type,omitempty"`
	Stacktrace sentry.Stacktrace `json:"stacktrace,omitempty"`
}

type FrontendSentryException struct {
	Values []FrontendSentryExceptionValue `json:"values,omitempty"`
}

type FrontendSentryEvent struct {
	*sentry.Event
	Exception    *FrontendSentryException `json:"exception,omitempty"`
	Measurements Measurements             `json:"measurements,omitempty"`
}

func (value *FrontendSentryExceptionValue) FmtMessage() string {
	return fmt.Sprintf("%s: %s", value.Type, value.Value)
}

func fmtLine(frame sentry.Frame) string {
	module := ""
	if len(frame.Module) > 0 {
		module = frame.Module + "|"
	}
	return fmt.Sprintf("\n  at %s (%s%s:%v:%v)", frame.Function, module, frame.Filename, frame.Lineno, frame.Colno)
}

func (value *FrontendSentryExceptionValue) FmtStacktrace(store *SourceMapStore, l log.Logger) string {
	var stacktrace = value.FmtMessage()
	for _, frame := range value.Stacktrace.Frames {
		mappedFrame, err := store.resolveSourceLocation(frame)
		if err != nil {
			level.Error(l).Log("msg", "Error resolving stack trace frame source location", "err", err)
			stacktrace += fmtLine(frame) // even if reading source map fails for unexpected reason, still better to log compiled location than nothing at all
		} else {
			if mappedFrame != nil {
				stacktrace += fmtLine(*mappedFrame)
			} else {
				stacktrace += fmtLine(frame)
			}
		}
	}
	return stacktrace
}

func (exception *FrontendSentryException) FmtStacktraces(store *SourceMapStore, l log.Logger) string {
	var stacktraces []string
	for _, value := range exception.Values {
		stacktraces = append(stacktraces, value.FmtStacktrace(store, l))
	}
	return strings.Join(stacktraces, "\n\n")
}

func addEventContextToLogContext(rootPrefix string, logCtx LogContext, eventCtx map[string]interface{}) {
	for key, element := range eventCtx {
		prefix := fmt.Sprintf("%s__%s", rootPrefix, key)
		switch v := element.(type) {
		case LogContext:
			if key == "trace" {
				logCtx["traceID"] = v["rootTraceId"]
				logCtx["spanID"] = v["rootSpanId"]
			} else if key == "measurements" {
				for name, measurement := range v {
					logCtx[name] = measurement
				}
			} else {
				addEventContextToLogContext(prefix, logCtx, v)
			}
		default:
			logCtx[prefix] = fmt.Sprintf("%v", v)
		}
	}
}

func (event *FrontendSentryEvent) ToLogContext(store *SourceMapStore, l log.Logger) LogContext {
	var ctx = make(LogContext)
	ctx["url"] = event.Request.URL
	ctx["user_agent"] = event.Request.Headers["User-Agent"]
	ctx["event_id"] = event.EventID
	ctx["original_timestamp"] = event.Timestamp
	ctx["kind"] = event.GetKind()
	msg := event.GetMessage()
	if len(msg) > 0 {
		ctx["msg"] = msg
	}
	if event.Exception != nil {
		ctx["stacktrace"] = event.Exception.FmtStacktraces(store, l)
	}
	if len(event.Release) > 0 {
		ctx["release"] = event.Release
	}
	if len(event.Environment) > 0 {
		ctx["environment"] = event.Environment
	}
	if len(event.Dist) > 0 {
		ctx["dist"] = event.Dist
	}
	if len(event.Platform) > 0 {
		ctx["platform"] = event.Platform
	}
	addEventContextToLogContext("context", ctx, event.Contexts)
	if len(event.User.Email) > 0 {
		ctx["user_email"] = event.User.Email
	}
	if len(event.User.ID) > 0 {
		ctx["user_id"] = event.User.ID
	}
	if len(event.User.Username) > 0 {
		ctx["user_username"] = event.User.Username
	}
	if len(event.User.IPAddress) > 0 {
		ctx["user_ip_adddress"] = event.User.IPAddress
	}

	if event.Measurements != nil {
		for name, measurement := range event.Measurements {
			ctx[name] = measurement.Value
		}
	}

	return ctx
}

func (event *FrontendSentryEvent) GetMessage() string {
	if len(event.Message) > 0 {
		return event.Message
	} else if event.Exception != nil && len(event.Exception.Values) > 0 {
		return event.Exception.Values[0].FmtMessage()
	}
	return ""
}

func (event *FrontendSentryEvent) GetKind() string {
	if event.Exception != nil {
		return "exception"
	}
	if event.Measurements != nil || event.Message == "measurements" {
		return "perf"
	}
	if len(event.Message) > 0 {
		return "log"
	}
	return ""
}

func LogContextToKeyVals(ctx LogContext) []interface{} {
	keyvals := []interface{}{}
	for key, value := range ctx {
		keyvals = append(keyvals, key, value)
	}
	return keyvals
}
