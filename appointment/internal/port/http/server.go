package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/api"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	"go.uber.org/zap"
)

type Server struct {
	service *application.Service
	log     *zap.Logger
}

func NewServer(service *application.Service, log *zap.Logger) *Server {
	return &Server{service: service, log: log}
}

var _ api.ServerInterface = (*Server)(nil)

func NewRouter(server api.ServerInterface) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	api.RegisterHandlers(r, server)
	return r
}

func (s *Server) CreateAgendaActivity(c *gin.Context) {
	var req api.CreateAgendaActivityJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	typ, title, desc, start, end, attendee, services, err := parseCreate(req)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	event, err := s.service.CreateAgenda(c.Request.Context(), typ, title, desc, start, end, attendee, services)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	c.JSON(201, gin.H{"id": event.ID})
}
func (s *Server) GetActivityById(c *gin.Context, activityId uuid.UUID) {
	event, err := s.service.GetAgenda(c.Request.Context(), activityId.String())
	if err != nil {
		c.JSON(500, gin.H{"message": "internal error"})
		return
	}
	if event == nil {
		c.JSON(404, gin.H{"message": "activity not found"})
		return
	}
	c.JSON(200, toActivityResponse(*event))
}
func (s *Server) GetAppointmentsByTimeRangeOrCustomer(c *gin.Context, params api.GetAppointmentsByTimeRangeOrCustomerParams) {
	attendee := ""
	if params.AttendeeId != nil {
		attendee = *params.AttendeeId
	}
	events, err := s.service.SearchAgenda(c.Request.Context(), attendee, params.Start, params.End)
	if err != nil {
		c.JSON(500, gin.H{"message": "internal error"})
		return
	}
	out := make([]api.ActivityResponse, 0, len(events))
	for _, e := range events {
		out = append(out, toActivityResponse(e))
	}
	c.JSON(200, out)
}
func (s *Server) UpdateActivity(c *gin.Context, activityId uuid.UUID) {
	var req api.UpdateActivityJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	var services []domain.AppointmentServiceRef
	if req.Services != nil {
		for _, name := range *req.Services {
			services = append(services, domain.AppointmentServiceRef{Name: name})
		}
	}
	event, err := s.service.UpdateAgenda(c.Request.Context(), activityId.String(), req.Title, req.Description, req.Start, req.End, services)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	if event == nil {
		c.JSON(404, gin.H{"message": "activity not found"})
		return
	}
	c.JSON(200, toActivityResponse(*event))
}
func (s *Server) DeleteActivity(c *gin.Context, activityId uuid.UUID, params api.DeleteActivityParams) {
	reason := domain.CancelReasonDeleted
	if params.Reason != nil && *params.Reason == api.CustomerCancel {
		reason = domain.CancelReasonCustomer
	}
	event, err := s.service.DeleteAgenda(c.Request.Context(), activityId.String(), reason)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	if event == nil {
		c.JSON(404, gin.H{"message": "activity not found"})
		return
	}
	c.Status(204)
}
func (s *Server) ResendReminder(c *gin.Context, activityId uuid.UUID) {
	event, err := s.service.GetAgenda(c.Request.Context(), activityId.String())
	if err != nil {
		c.JSON(500, gin.H{"message": "internal error"})
		return
	}
	if event == nil {
		c.JSON(404, gin.H{"message": "activity not found"})
		return
	}
	c.JSON(200, toActivityResponse(*event))
}

func (s *Server) CreateService(c *gin.Context) {
	var req api.CreateServiceJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	tags := []string{}
	if req.Tags != nil {
		tags = *req.Tags
	}
	svc, err := s.service.CreateService(c.Request.Context(), req.Name, float64(req.Price), tags, req.Color)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	c.JSON(201, toService(svc))
}
func (s *Server) UpdateService(c *gin.Context, serviceId string) {
	var req api.UpdateServiceRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	var price *float64
	if req.Price != nil {
		v := float64(*req.Price)
		price = &v
	}
	var tags []string
	if req.Tags != nil {
		tags = *req.Tags
	}
	svc, err := s.service.UpdateService(c.Request.Context(), serviceId, price, tags, req.Color)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	if svc.ID == "" {
		c.JSON(404, gin.H{"message": "service not found"})
		return
	}
	c.JSON(200, toService(svc))
}
func (s *Server) GetAllServices(c *gin.Context) {
	services, err := s.service.AllServices(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"message": "internal error"})
		return
	}
	c.JSON(200, toServices(services))
}
func (s *Server) SearchService(c *gin.Context, params api.SearchServiceParams) {
	text := ""
	limit := 10
	if params.Text != nil {
		text = *params.Text
	}
	if params.Limit != nil {
		limit = *params.Limit
	}
	services, err := s.service.SearchServices(c.Request.Context(), text, limit)
	if err != nil {
		c.JSON(500, gin.H{"message": "internal error"})
		return
	}
	c.JSON(200, toServices(services))
}
func (s *Server) GetCustomerRanking(c *gin.Context, params api.GetCustomerRankingParams) {
	c.JSON(200, gin.H{"items": []any{}})
}
func (s *Server) GetCustomerCancelRanking(c *gin.Context, params api.GetCustomerCancelRankingParams) {
	c.JSON(200, gin.H{"items": []any{}})
}
func (s *Server) GetInsightOverview(c *gin.Context) { c.JSON(200, api.StatisticsResponseBody{}) }

func parseCreate(req api.CreateAgendaActivityJSONRequestBody) (domain.EventType, string, string, time.Time, time.Time, domain.Attendee, []domain.AppointmentServiceRef, error) {
	disc, err := req.Discriminator()
	if err != nil {
		return "", "", "", time.Time{}, time.Time{}, domain.Attendee{}, nil, err
	}
	switch disc {
	case "appointment":
		v, err := req.AsAppointmentEvent()
		if err != nil {
			return "", "", "", time.Time{}, time.Time{}, domain.Attendee{}, nil, err
		}
		services := []domain.AppointmentServiceRef{}
		if v.Appointment.Services != nil {
			for _, name := range *v.Appointment.Services {
				services = append(services, domain.AppointmentServiceRef{Name: name})
			}
		}
		return domain.EventTypeAppointment, v.Title, "", v.Start, v.End, domain.Attendee{ID: v.AttendeeId.String(), DisplayName: ""}, services, nil
	default:
		v, err := req.AsGenericEvent()
		if err != nil {
			return "", "", "", time.Time{}, time.Time{}, domain.Attendee{}, nil, err
		}
		desc := ""
		if v.Description != nil {
			desc = *v.Description
		}
		return domain.EventTypeGeneric, v.Title, desc, v.Start, v.End, domain.Attendee{ID: v.AttendeeId.String(), DisplayName: "self"}, nil, nil
	}
}
func toServices(in []domain.AppointmentService) []api.Service {
	out := make([]api.Service, 0, len(in))
	for _, s := range in {
		out = append(out, toService(s))
	}
	return out
}
func toService(s domain.AppointmentService) api.Service {
	return api.Service{Id: s.ID, Name: s.Name, Price: float32(s.Price), Tags: s.Tags, Color: s.Color}
}
func toActivityResponse(e domain.AgendaEvent) api.ActivityResponse {
	if e.Type == domain.EventTypeAppointment {
		services := []string{}
		for _, s := range e.Services {
			services = append(services, s.Name)
		}
		r := api.AppointmentEventResponse{Id: uuid.MustParse(e.ID), Type: api.AppointmentEventResponseTypeAppointment, Start: e.Start, End: e.End, Attendee: attendee(e), IsCanceled: e.CancelReason != nil, ReminderSent: e.ReminderStatus == domain.ReminderSent}
		r.Appointment.Services = &services
		title := e.Title
		r.Title = &title
		mins := int(e.RemindBefore.Minutes())
		status := api.AppointmentEventResponseReminderStatusNOTSENT
		r.Reminder.ReminderMinutes = &mins
		r.Reminder.Status = &status
		if e.CancelReason != nil {
			v := string(*e.CancelReason)
			r.CancelReason = &v
		}
		var out api.ActivityResponse
		_ = out.FromAppointmentEventResponse(r)
		return out
	}
	r := api.EventResponse{Id: uuid.MustParse(e.ID), Type: api.EventResponseTypeEvent, Start: e.Start, End: e.End, Attendee: attendee(e), IsCanceled: e.CancelReason != nil, ReminderSent: e.ReminderStatus == domain.ReminderSent}
	title := e.Title
	desc := e.Description
	r.Title = &title
	r.Description = &desc
	mins := int(e.RemindBefore.Minutes())
	status := api.EventResponseReminderStatusNOTSENT
	r.Reminder.ReminderMinutes = &mins
	r.Reminder.Status = &status
	if e.CancelReason != nil {
		v := string(*e.CancelReason)
		r.CancelReason = &v
	}
	var out api.ActivityResponse
	_ = out.FromEventResponse(r)
	return out
}
func attendee(e domain.AgendaEvent) api.Attendee {
	parts := strings.SplitN(e.Attendee.DisplayName, " ", 2)
	name := e.Attendee.DisplayName
	surname := ""
	if len(parts) > 0 {
		name = parts[0]
	}
	if len(parts) > 1 {
		surname = parts[1]
	}
	return api.Attendee{Id: uuid.MustParse(e.Attendee.ID), Name: name, Surname: surname}
}
