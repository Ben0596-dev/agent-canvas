package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"agent-canvas/backend/internal/model"
	"agent-canvas/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AgentController struct {
	service *service.AgentService
}

type createAgentRunRequest struct {
	SessionID       string   `json:"sessionId"`
	Message         string   `json:"message" binding:"required"`
	SelectedNodeIDs []string `json:"selectedNodeIds" binding:"max=100"`
}

type agentSessionResponse struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type agentRunResponse struct {
	ID        string               `json:"id"`
	ProjectID string               `json:"projectId"`
	SessionID string               `json:"sessionId"`
	Status    string               `json:"status"`
	Message   string               `json:"message"`
	Proposal  *model.AgentProposal `json:"proposal,omitempty"`
	Error     string               `json:"error,omitempty"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

func NewAgentController(service *service.AgentService) *AgentController {
	return &AgentController{service: service}
}

func (controller *AgentController) Register(router *gin.RouterGroup) {
	router.GET("/projects/:projectId/agent/sessions", controller.listSessions)
	router.GET("/projects/:projectId/agent/sessions/:sessionId/runs", controller.listRuns)
	router.POST("/projects/:projectId/agent/runs", controller.createRun)
	router.GET("/agent/runs/:runId", controller.getRun)
	router.GET("/agent/runs/:runId/events", controller.streamEvents)
	router.POST("/agent/runs/:runId/cancel", controller.cancelRun)
}

func (controller *AgentController) listRuns(c *gin.Context) {
	runs, err := controller.service.ListRuns(c.Request.Context(), c.Param("projectId"), c.Param("sessionId"))
	if err != nil {
		writeAgentError(c, err)
		return
	}
	items := make([]agentRunResponse, 0, len(runs))
	for _, run := range runs {
		items = append(items, newAgentRunResponse(run))
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (controller *AgentController) listSessions(c *gin.Context) {
	sessions, err := controller.service.ListSessions(c.Request.Context(), c.Param("projectId"))
	if err != nil {
		writeAgentError(c, err)
		return
	}
	items := make([]agentSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, agentSessionResponse{
			ID: session.ID, ProjectID: session.ProjectID, Title: session.Title,
			CreatedAt: session.CreatedAt, UpdatedAt: session.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (controller *AgentController) createRun(c *gin.Context) {
	var request createAgentRunRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_request", "message": "message is required"})
		return
	}
	run, err := controller.service.CreateRun(
		c.Request.Context(), c.Param("projectId"), request.SessionID, request.Message, request.SelectedNodeIDs,
	)
	if err != nil {
		writeAgentError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, newAgentRunResponse(run))
}

func (controller *AgentController) getRun(c *gin.Context) {
	run, err := controller.service.GetRun(c.Request.Context(), c.Param("runId"))
	if err != nil {
		writeAgentError(c, err)
		return
	}
	c.JSON(http.StatusOK, newAgentRunResponse(run))
}

func (controller *AgentController) cancelRun(c *gin.Context) {
	run, err := controller.service.CancelRun(c.Request.Context(), c.Param("runId"))
	if err != nil {
		writeAgentError(c, err)
		return
	}
	c.JSON(http.StatusOK, newAgentRunResponse(run))
}

func (controller *AgentController) streamEvents(c *gin.Context) {
	runID := c.Param("runId")
	run, err := controller.service.GetRun(c.Request.Context(), runID)
	if err != nil {
		writeAgentError(c, err)
		return
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	lastStatus := ""
	for {
		if run.Status != lastStatus {
			if writeRunEvents(c, run) {
				return
			}
			lastStatus = run.Status
		}
		select {
		case <-c.Request.Context().Done():
			return
		case <-time.After(100 * time.Millisecond):
		}
		run, err = controller.service.GetRun(c.Request.Context(), runID)
		if err != nil {
			return
		}
	}
}

func writeRunEvents(c *gin.Context, run model.AgentRun) bool {
	switch run.Status {
	case model.AgentRunQueued, model.AgentRunRunning:
		writeSSE(c, "run.started", newAgentRunResponse(run))
	case model.AgentRunSucceeded:
		writeSSE(c, "operation.proposed", run.Proposal)
		writeSSE(c, "run.completed", newAgentRunResponse(run))
		return true
	case model.AgentRunFailed:
		writeSSE(c, "run.failed", newAgentRunResponse(run))
		return true
	case model.AgentRunCancelled:
		writeSSE(c, "run.cancelled", newAgentRunResponse(run))
		return true
	}
	return false
}

func writeSSE(c *gin.Context, event string, data any) {
	payload, _ := json.Marshal(data)
	_, _ = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, payload)
	c.Writer.Flush()
}

func newAgentRunResponse(run model.AgentRun) agentRunResponse {
	response := agentRunResponse{
		ID: run.ID, ProjectID: run.ProjectID, SessionID: run.SessionID,
		Status: run.Status, Message: run.Message, Error: run.ErrorMessage,
		CreatedAt: run.CreatedAt, UpdatedAt: run.UpdatedAt,
	}
	if run.Status == model.AgentRunSucceeded {
		response.Proposal = &run.Proposal
	}
	return response
}

func writeAgentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidAgentMessage):
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_agent_message", "message": err.Error()})
	case errors.Is(err, service.ErrProjectNotFound):
		c.JSON(http.StatusNotFound, gin.H{"code": "project_not_found", "message": "project not found"})
	case errors.Is(err, service.ErrAgentSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"code": "agent_session_not_found", "message": "agent session not found"})
	case errors.Is(err, service.ErrAgentRunNotFound):
		c.JSON(http.StatusNotFound, gin.H{"code": "agent_run_not_found", "message": "agent run not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "agent request failed"})
	}
}
