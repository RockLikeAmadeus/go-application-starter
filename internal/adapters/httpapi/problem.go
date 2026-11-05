package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi/openapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// Problem type codes. The HTTP client adapter maps these back to domain errors.
const (
	ProblemNotFound   = "not_found"
	ProblemValidation = "validation"
	ProblemLimit      = "limit"
	ProblemFeature    = "feature"
	ProblemBadRequest = "bad_request"
	ProblemInternal   = "internal"
)

// problemFor maps a domain error to an RFC 9457 problem document.
func problemFor(err error) openapi.Problem {
	var validationErr *domain.ValidationError
	var limitErr *domain.LimitError
	var featureErr *domain.FeatureError
	switch {
	case errors.As(err, &validationErr):
		return problem(http.StatusUnprocessableEntity, ProblemValidation, validationErr.Error(), &validationErr.Field)
	case errors.As(err, &limitErr):
		p := problem(http.StatusForbidden, ProblemLimit, limitErr.Error(), nil)
		p.Limit, p.Max, p.IsReadOnly = &limitErr.Limit, &limitErr.Max, &limitErr.IsReadOnly
		return p
	case errors.As(err, &featureErr):
		p := problem(http.StatusForbidden, ProblemFeature, featureErr.Error(), nil)
		availableOn := availableOnDTO(featureErr.AvailableOn)
		p.Feature, p.Name, p.AvailableOn = &featureErr.Feature, &featureErr.Name, &availableOn
		return p
	case errors.Is(err, domain.ErrNotFound):
		return problem(http.StatusNotFound, ProblemNotFound, err.Error(), nil)
	default:
		return problem(http.StatusInternalServerError, ProblemInternal, "internal server error", nil)
	}
}

func problem(status int, typ, detail string, field *string) openapi.Problem {
	return openapi.Problem{Status: status, Title: http.StatusText(status), Type: &typ, Detail: &detail, Field: field}
}

func writeProblem(w http.ResponseWriter, p openapi.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}

func responseErrorHandler(log *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		p := problemFor(err)
		if p.Status >= 500 {
			log.ErrorContext(r.Context(), "request failed", "method", r.Method, "path", r.URL.Path, "err", err)
		}
		writeProblem(w, p)
	}
}

func requestErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	writeProblem(w, problem(http.StatusBadRequest, ProblemBadRequest, err.Error(), nil))
}
