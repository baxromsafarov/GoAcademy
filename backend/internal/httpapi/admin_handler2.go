package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/goacademy/backend/internal/admin"
	"github.com/goacademy/backend/internal/httpapi/respond"
)

// tracks ----------------------------------------------------------------------

type adminTrackItemRequest struct {
	ContentType string `json:"content_type"`
	ContentID   string `json:"content_id"`
}

type adminTrackRequest struct {
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Level       string                  `json:"level"`
	Position    int                     `json:"position"`
	Language    string                  `json:"language"`
	Items       []adminTrackItemRequest `json:"items"`
}

func (r adminTrackRequest) toInput() admin.TrackInput {
	items := make([]admin.TrackItemInput, 0, len(r.Items))
	for _, it := range r.Items {
		items = append(items, admin.TrackItemInput{ContentType: it.ContentType, ContentID: it.ContentID})
	}
	return admin.TrackInput{
		Title: r.Title, Description: r.Description, Level: r.Level,
		Position: r.Position, Language: r.Language, Items: items,
	}
}

func (h *adminHandler) createTrack(w http.ResponseWriter, r *http.Request) {
	var req adminTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	t, err := h.svc.CreateTrack(r.Context(), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toTrackMeta(t))
}

func (h *adminHandler) updateTrack(w http.ResponseWriter, r *http.Request) {
	var req adminTrackRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	t, err := h.svc.UpdateTrack(r.Context(), chi.URLParam(r, "id"), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toTrackMeta(t))
}

func (h *adminHandler) deleteTrack(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteTrack(r.Context(), chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// cheatsheets -----------------------------------------------------------------

type adminCheatsheetRequest struct {
	Title        string `json:"title"`
	Category     string `json:"category"`
	BodyMarkdown string `json:"body_markdown"`
	Language     string `json:"language"`
}

func (r adminCheatsheetRequest) toInput() admin.CheatsheetInput {
	return admin.CheatsheetInput{Title: r.Title, Category: r.Category, BodyMarkdown: r.BodyMarkdown, Language: r.Language}
}

func (h *adminHandler) createCheatsheet(w http.ResponseWriter, r *http.Request) {
	var req adminCheatsheetRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	c, err := h.svc.CreateCheatsheet(r.Context(), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toCheatsheetDetailResponse(c))
}

func (h *adminHandler) updateCheatsheet(w http.ResponseWriter, r *http.Request) {
	var req adminCheatsheetRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	c, err := h.svc.UpdateCheatsheet(r.Context(), chi.URLParam(r, "id"), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toCheatsheetDetailResponse(c))
}

func (h *adminHandler) deleteCheatsheet(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteCheatsheet(r.Context(), chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// projects --------------------------------------------------------------------

type adminProjectStepRequest struct {
	Text string `json:"text"`
}

type adminProjectRequest struct {
	Title               string                    `json:"title"`
	DescriptionMarkdown string                    `json:"description_markdown"`
	Difficulty          string                    `json:"difficulty"`
	Language            string                    `json:"language"`
	Tags                []string                  `json:"tags"`
	Steps               []adminProjectStepRequest `json:"steps"`
}

func (r adminProjectRequest) toInput() admin.ProjectInput {
	steps := make([]admin.ProjectStepInput, 0, len(r.Steps))
	for _, st := range r.Steps {
		steps = append(steps, admin.ProjectStepInput{Text: st.Text})
	}
	return admin.ProjectInput{
		Title: r.Title, DescriptionMarkdown: r.DescriptionMarkdown, Difficulty: r.Difficulty,
		Language: r.Language, Tags: r.Tags, Steps: steps,
	}
}

func (h *adminHandler) createProject(w http.ResponseWriter, r *http.Request) {
	var req adminProjectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	p, err := h.svc.CreateProject(r.Context(), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toProjectMeta(p))
}

func (h *adminHandler) updateProject(w http.ResponseWriter, r *http.Request) {
	var req adminProjectRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	p, err := h.svc.UpdateProject(r.Context(), chi.URLParam(r, "id"), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toProjectMeta(p))
}

func (h *adminHandler) deleteProject(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteProject(r.Context(), chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// glossary --------------------------------------------------------------------

type adminGlossaryRequest struct {
	Term               string `json:"term"`
	DefinitionMarkdown string `json:"definition_markdown"`
	Language           string `json:"language"`
}

func (r adminGlossaryRequest) toInput() admin.GlossaryInput {
	return admin.GlossaryInput{Term: r.Term, DefinitionMarkdown: r.DefinitionMarkdown, Language: r.Language}
}

func (h *adminHandler) createGlossary(w http.ResponseWriter, r *http.Request) {
	var req adminGlossaryRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	g, err := h.svc.CreateGlossaryTerm(r.Context(), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toGlossaryItem(g))
}

func (h *adminHandler) updateGlossary(w http.ResponseWriter, r *http.Request) {
	var req adminGlossaryRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	g, err := h.svc.UpdateGlossaryTerm(r.Context(), chi.URLParam(r, "id"), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toGlossaryItem(g))
}

func (h *adminHandler) deleteGlossary(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteGlossaryTerm(r.Context(), chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// badges ----------------------------------------------------------------------

type adminBadgeRequest struct {
	Code           string          `json:"code"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	Icon           string          `json:"icon"`
	CriteriaType   string          `json:"criteria_type"`
	CriteriaParams json.RawMessage `json:"criteria_params"`
}

func (r adminBadgeRequest) toInput() admin.BadgeInput {
	return admin.BadgeInput{
		Code: r.Code, Title: r.Title, Description: r.Description, Icon: r.Icon,
		CriteriaType: r.CriteriaType, CriteriaParams: r.CriteriaParams,
	}
}

func (h *adminHandler) createBadge(w http.ResponseWriter, r *http.Request) {
	var req adminBadgeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	b, err := h.svc.CreateBadge(r.Context(), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toAdminBadgeResponse(b))
}

func (h *adminHandler) updateBadge(w http.ResponseWriter, r *http.Request) {
	var req adminBadgeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	b, err := h.svc.UpdateBadge(r.Context(), chi.URLParam(r, "id"), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toAdminBadgeResponse(b))
}

func (h *adminHandler) deleteBadge(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteBadge(r.Context(), chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// daily challenges ------------------------------------------------------------

type adminDailyChallengeRequest struct {
	ChallengeDate string `json:"challenge_date"`
	ContentType   string `json:"content_type"`
	ContentID     string `json:"content_id"`
	BonusXP       int    `json:"bonus_xp"`
}

func (r adminDailyChallengeRequest) toInput() admin.DailyChallengeInput {
	return admin.DailyChallengeInput{
		ChallengeDate: r.ChallengeDate, ContentType: r.ContentType, ContentID: r.ContentID, BonusXP: r.BonusXP,
	}
}

func (h *adminHandler) createDailyChallenge(w http.ResponseWriter, r *http.Request) {
	var req adminDailyChallengeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	d, err := h.svc.CreateDailyChallenge(r.Context(), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusCreated, toAdminDailyChallengeResponse(d))
}

func (h *adminHandler) updateDailyChallenge(w http.ResponseWriter, r *http.Request) {
	var req adminDailyChallengeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	d, err := h.svc.UpdateDailyChallenge(r.Context(), chi.URLParam(r, "id"), req.toInput())
	if err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	respond.JSON(w, http.StatusOK, toAdminDailyChallengeResponse(d))
}

func (h *adminHandler) deleteDailyChallenge(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteDailyChallenge(r.Context(), chi.URLParam(r, "id")); err != nil {
		respond.Error(w, r, h.logger, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
