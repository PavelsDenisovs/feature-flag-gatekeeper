package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/application"
	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/flag/domain"
	"github.com/PavelsDenisovs/feature-flag-gatekeeper/internal/problems"
)

func RegisterEndpoints(
	mux *http.ServeMux,
	service application.FlagService,
) {
	h := handler{svc: service}
	mux.HandleFunc("POST /flags/{flag_key}/evaluate", h.Evaluate)
}

type handler struct {
	svc application.FlagService
}

func (h *handler) Evaluate(w http.ResponseWriter, r *http.Request) {
	date := time.Now().Format("2006-01-02")
	flagKey := r.PathValue("flag_key")
	if flagKey == "" {
		handleErrorResponse(w, http.StatusBadRequest, problems.ProblemParams{
			Type:     fmt.Sprintf("tag:feature-flag-gatekeeper,%v:flag/evaluate/flag-key-absent", date),
			Title:    "flag_key is not present in the request URL",
			Status:   http.StatusBadRequest,
			Instance: "/flag",
		})
		return
	}

	// 5KB
	limit := 1024 * 5
	raw := make([]byte, limit+1)

	n, err := io.ReadFull(r.Body, raw)
	if n > limit {
		handleErrorResponse(w, http.StatusRequestEntityTooLarge, problems.ProblemParams{
			Type:     fmt.Sprintf("tag:feature-flag-gatekeeper,%v:problems/flag/evaluate/big-request", date),
			Title:    http.StatusText(http.StatusRequestEntityTooLarge),
			Status:   http.StatusRequestEntityTooLarge,
			Detail:   fmt.Sprintf("Request exceeds its size limit of %d KB", limit/1024),
			Instance: fmt.Sprintf("/flag/%v", flagKey),
		})
		return
	}
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		handleErrorResponse(w, http.StatusInternalServerError, problems.ProblemParams{
			Type:     fmt.Sprintf("tag:feature-flag-gatekeeper,%v:flag/evaluate/internal-server-error", date),
			Title:    "Failed to read request body",
			Status:   http.StatusInternalServerError,
			Instance: fmt.Sprintf("/flag/%v", flagKey),
		})
		return
	}

	if n == 0 {
		handleErrorResponse(w, http.StatusBadRequest, problems.ProblemParams{
			Type:     fmt.Sprintf("tag:feature-flag-gatekeeper,%v:flag/evaluate/body-absent", date),
			Title:    "Body is absent",
			Status:   http.StatusBadRequest,
			Instance: fmt.Sprintf("/flag/%v", flagKey),
		})
		return
	}

	body := struct {
		SubjectKey string            `json:"subject_key"`
		Attributes map[string]string `json:"attributes"`
	}{}

	err = json.Unmarshal(raw[:n], &body)
	if err != nil {
		handleErrorResponse(w, http.StatusBadRequest, problems.ProblemParams{
			Type:     fmt.Sprintf("tag:feature-flag-gatekeeper,%v:flag/evaluate/invalid-request", date),
			Title:    "Failed to unmarshal request data into req struct",
			Status:   http.StatusBadRequest,
			Instance: fmt.Sprintf("/flag/%v", flagKey),
		})
		return
	}

	ctx := r.Context()
	res, err := h.svc.Evaluate(ctx,
		application.EvaluateRequest{
			SubjectKey: body.SubjectKey,
			FlagKey:    flagKey,
			Attributes: body.Attributes,
		},
	)
	if err != nil {
		if errors.Is(err, domain.ErrFlagNotFound) {
			handleErrorResponse(w, http.StatusNotFound, problems.ProblemParams{
				Type:     fmt.Sprintf("tag:feature-flag-gatekeeper,%v:flag/evaluate/flag-not-found", date),
				Title:    "Flag not found",
				Status:   http.StatusNotFound,
				Detail:   fmt.Sprintf("Flag with key %s is not found for evaluation", flagKey),
				Instance: fmt.Sprintf("/flag/%v", flagKey),
			})
			return
		}
		if errors.Is(err, domain.ErrEvaluationFailed) {
			handleErrorResponse(w, http.StatusBadRequest, problems.ProblemParams{
				Type:     fmt.Sprintf("tag:feature-flag-gatekeeper,%v:flag/evaluate/evaluation-failed", date),
				Title:    "Evaluation failed",
				Status:   http.StatusBadRequest,
				Detail:   fmt.Sprintf("Flag with key %s failed evaluation: %v", flagKey, err),
				Instance: fmt.Sprintf("/flag/%v", flagKey),
			})
			return
		}

		handleErrorResponse(w, http.StatusInternalServerError, problems.ProblemParams{
			Type:     fmt.Sprintf("tag:feature-flag-gatekeeper,%v:flag/evaluate/internal-server-error", date),
			Title:    "Failed to evaluate flag",
			Status:   http.StatusInternalServerError,
			Instance: fmt.Sprintf("/flag/%v", flagKey),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"enabled": %t}`, res.Enabled)))
}

func handleErrorResponse(w http.ResponseWriter, status int, params problems.ProblemParams) {
	p, err := problems.New(params)
	if err != nil {
		slog.Error(err.Error())
	}
	if err := problems.WriteProblem(w, status, p); err != nil {
		slog.Error(err.Error())
	}
}
