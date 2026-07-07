package server

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	httpserver "github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/server/generated"
	"go.uber.org/zap"
)

func (s *Server) CreateAgendaActivity(ctx context.Context, request httpserver.CreateAgendaActivityRequestObject) (httpserver.CreateAgendaActivityResponseObject, error) {
	if request.Body == nil {
		return httpserver.CreateAgendaActivity400JSONResponse(errorResponse("request body is required")), nil
	}
	typ, title, desc, start, end, attendee, services, err := parseCreate(*request.Body)
	if err != nil {
		s.logHandlerError("CreateAgendaActivity", err)
		return httpserver.CreateAgendaActivity400JSONResponse(errorResponse(err.Error())), nil
	}
	event, err := s.appointments.CreateAgenda(ctx, typ, title, desc, start, end, attendee, services)
	if err != nil {
		s.logHandlerError("CreateAgendaActivity", err)
		return httpserver.CreateAgendaActivity400JSONResponse(errorResponse(err.Error())), nil
	}
	id := uuid.MustParse(event.ID)
	return httpserver.CreateAgendaActivity201JSONResponse{Id: &id}, nil
}

func (s *Server) GetActivityById(ctx context.Context, request httpserver.GetActivityByIdRequestObject) (httpserver.GetActivityByIdResponseObject, error) {
	event, err := s.appointments.GetAgenda(ctx, request.ActivityId.String())
	if err != nil {
		return nil, err
	}
	if event == nil {
		return httpserver.GetActivityById404JSONResponse(errorResponse("activity not found")), nil
	}
	return httpserver.GetActivityById200JSONResponse(toActivityResponse(*event)), nil
}

func (s *Server) GetAppointmentsByTimeRangeOrCustomer(ctx context.Context, request httpserver.GetAppointmentsByTimeRangeOrCustomerRequestObject) (httpserver.GetAppointmentsByTimeRangeOrCustomerResponseObject, error) {
	attendee := ""
	if request.Params.AttendeeId != nil {
		attendee = *request.Params.AttendeeId
	}
	events, err := s.appointments.SearchAgenda(ctx, attendee, request.Params.Start, request.Params.End)
	if err != nil {
		return nil, err
	}
	out := make([]httpserver.ActivityResponse, 0, len(events))
	for _, e := range events {
		out = append(out, toActivityResponse(e))
	}
	return httpserver.GetAppointmentsByTimeRangeOrCustomer200JSONResponse(out), nil
}

func (s *Server) UpdateActivity(ctx context.Context, request httpserver.UpdateActivityRequestObject) (httpserver.UpdateActivityResponseObject, error) {
	if request.Body == nil {
		return httpserver.UpdateActivity400JSONResponse(errorResponse("request body is required")), nil
	}
	req := request.Body
	var services []domain.AppointmentServiceRef
	if req.Services != nil {
		for _, name := range *req.Services {
			services = append(services, domain.AppointmentServiceRef{Name: name})
		}
	}
	event, err := s.appointments.UpdateAgenda(ctx, request.ActivityId.String(), req.Title, req.Description, req.Start, req.End, services)
	if err != nil {
		s.logHandlerError("UpdateActivity", err)
		return httpserver.UpdateActivity400JSONResponse(errorResponse(err.Error())), nil
	}
	if event == nil {
		return httpserver.UpdateActivity400JSONResponse(errorResponse("activity not found")), nil
	}
	return httpserver.UpdateActivity200JSONResponse(toActivityResponse(*event)), nil
}

func (s *Server) DeleteActivity(ctx context.Context, request httpserver.DeleteActivityRequestObject) (httpserver.DeleteActivityResponseObject, error) {
	reason := domain.CancelReasonDeleted
	if request.Params.Reason != nil && *request.Params.Reason == httpserver.CustomerCancel {
		reason = domain.CancelReasonCustomer
	}
	event, err := s.appointments.DeleteAgenda(ctx, request.ActivityId.String(), reason)
	if err != nil {
		s.logHandlerError("DeleteActivity", err)
		return httpserver.DeleteActivity400JSONResponse(errorResponse(err.Error())), nil
	}
	if event == nil {
		return httpserver.DeleteActivity404JSONResponse(errorResponse("activity not found")), nil
	}
	return httpserver.DeleteActivity204Response{}, nil
}

func (s *Server) ResendReminder(ctx context.Context, request httpserver.ResendReminderRequestObject) (httpserver.ResendReminderResponseObject, error) {
	event, err := s.appointments.GetAgenda(ctx, request.ActivityId.String())
	if err != nil {
		return nil, err
	}
	if event == nil {
		return httpserver.ResendReminder404JSONResponse(errorResponse("activity not found")), nil
	}
	return httpserver.ResendReminder200JSONResponse(toActivityResponse(*event)), nil
}

func (s *Server) CreateService(ctx context.Context, request httpserver.CreateServiceRequestObject) (httpserver.CreateServiceResponseObject, error) {
	if request.Body == nil {
		return httpserver.CreateService400JSONResponse(errorResponse("request body is required")), nil
	}
	tags := []string{}
	if request.Body.Tags != nil {
		tags = *request.Body.Tags
	}
	svc, err := s.services.CreateService(ctx, request.Body.Name, float64(request.Body.Price), tags, request.Body.Color)
	if err != nil {
		s.logHandlerError("CreateService", err)
		return httpserver.CreateService400JSONResponse(errorResponse(err.Error())), nil
	}
	return httpserver.CreateService201JSONResponse(toService(svc)), nil
}

func (s *Server) UpdateService(ctx context.Context, request httpserver.UpdateServiceRequestObject) (httpserver.UpdateServiceResponseObject, error) {
	if request.Body == nil {
		return httpserver.UpdateService400JSONResponse(errorResponse("request body is required")), nil
	}
	var price *float64
	if request.Body.Price != nil {
		v := float64(*request.Body.Price)
		price = &v
	}
	var tags []string
	if request.Body.Tags != nil {
		tags = *request.Body.Tags
	}
	svc, err := s.services.UpdateService(ctx, request.ServiceId, price, tags, request.Body.Color)
	if err != nil {
		s.logHandlerError("UpdateService", err)
		return httpserver.UpdateService400JSONResponse(errorResponse(err.Error())), nil
	}
	if svc.ID == "" {
		return httpserver.UpdateService400JSONResponse(errorResponse("service not found")), nil
	}
	return httpserver.UpdateService200JSONResponse(toService(svc)), nil
}

func (s *Server) GetAllServices(ctx context.Context, request httpserver.GetAllServicesRequestObject) (httpserver.GetAllServicesResponseObject, error) {
	services, err := s.services.AllServices(ctx)
	if err != nil {
		return nil, err
	}
	return httpserver.GetAllServices200JSONResponse(toServices(services)), nil
}

func (s *Server) SearchService(ctx context.Context, request httpserver.SearchServiceRequestObject) (httpserver.SearchServiceResponseObject, error) {
	text := ""
	limit := 10
	if request.Params.Text != nil {
		text = *request.Params.Text
	}
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}
	services, err := s.services.SearchServices(ctx, text, limit)
	if err != nil {
		return nil, err
	}
	return httpserver.SearchService200JSONResponse(toServices(services)), nil
}

func (s *Server) GetCustomerRanking(ctx context.Context, request httpserver.GetCustomerRankingRequestObject) (httpserver.GetCustomerRankingResponseObject, error) {
	items := []httpserver.CustomerRankItemDto{}
	return httpserver.GetCustomerRanking200JSONResponse{CustomerRankingResponseBodyJSONResponse: httpserver.CustomerRankingResponseBodyJSONResponse{Items: &items}}, nil
}

func (s *Server) GetCustomerCancelRanking(ctx context.Context, request httpserver.GetCustomerCancelRankingRequestObject) (httpserver.GetCustomerCancelRankingResponseObject, error) {
	items := []httpserver.CustomerCancellationRankItemDto{}
	return httpserver.GetCustomerCancelRanking200JSONResponse{CustomerCancellationRankingResponseBodyJSONResponse: httpserver.CustomerCancellationRankingResponseBodyJSONResponse{Items: &items}}, nil
}

func (s *Server) GetInsightOverview(ctx context.Context, request httpserver.GetInsightOverviewRequestObject) (httpserver.GetInsightOverviewResponseObject, error) {
	items := []httpserver.CancellationDayOfWeekCount{}
	return httpserver.GetInsightOverview200JSONResponse{StatisticsResponseBodyJSONResponse: httpserver.StatisticsResponseBodyJSONResponse{CancellationDayOfWeek: &items}}, nil
}

func parseCreate(req httpserver.CreateAgendaActivityJSONRequestBody) (domain.EventType, string, string, time.Time, time.Time, domain.Attendee, []domain.AppointmentServiceRef, error) {
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

func toServices(in []domain.AppointmentService) []httpserver.Service {
	out := make([]httpserver.Service, 0, len(in))
	for _, s := range in {
		out = append(out, toService(s))
	}
	return out
}

func toService(s domain.AppointmentService) httpserver.Service {
	return httpserver.Service{Id: s.ID, Name: s.Name, Price: float32(s.Price), Tags: s.Tags, Color: s.Color}
}

func toActivityResponse(e domain.AgendaEvent) httpserver.ActivityResponse {
	if e.Type == domain.EventTypeAppointment {
		services := []string{}
		for _, s := range e.Services {
			services = append(services, s.Name)
		}
		r := httpserver.AppointmentEventResponse{Id: uuid.MustParse(e.ID), Type: httpserver.Appointment, Start: e.Start, End: e.End, Attendee: attendee(e), IsCanceled: e.CancelReason != nil, ReminderSent: e.ReminderStatus == domain.ReminderSent}
		r.Appointment.Services = &services
		title := e.Title
		r.Title = &title
		mins := int(e.RemindBefore.Minutes())
		status := httpserver.AppointmentEventResponseReminderStatusNOTSENT
		r.Reminder.ReminderMinutes = &mins
		r.Reminder.Status = &status
		if e.CancelReason != nil {
			v := string(*e.CancelReason)
			r.CancelReason = &v
		}
		var out httpserver.ActivityResponse
		_ = out.FromAppointmentEventResponse(r)
		return out
	}
	r := httpserver.EventResponse{Id: uuid.MustParse(e.ID), Type: httpserver.Event, Start: e.Start, End: e.End, Attendee: attendee(e), IsCanceled: e.CancelReason != nil, ReminderSent: e.ReminderStatus == domain.ReminderSent}
	title := e.Title
	desc := e.Description
	r.Title = &title
	r.Description = &desc
	mins := int(e.RemindBefore.Minutes())
	status := httpserver.EventResponseReminderStatusNOTSENT
	r.Reminder.ReminderMinutes = &mins
	r.Reminder.Status = &status
	if e.CancelReason != nil {
		v := string(*e.CancelReason)
		r.CancelReason = &v
	}
	var out httpserver.ActivityResponse
	_ = out.FromEventResponse(r)
	return out
}

func attendee(e domain.AgendaEvent) httpserver.Attendee {
	parts := strings.SplitN(e.Attendee.DisplayName, " ", 2)
	name := e.Attendee.DisplayName
	surname := ""
	if len(parts) > 0 {
		name = parts[0]
	}
	if len(parts) > 1 {
		surname = parts[1]
	}
	return httpserver.Attendee{Id: uuid.MustParse(e.Attendee.ID), Name: name, Surname: surname}
}

func (s *Server) logHandlerError(operation string, err error) {
	if s.log != nil {
		s.log.Warn("http handler error", zap.String("operation", operation), zap.Error(err))
	}
}
func errorResponse(message string) httpserver.ErrorResponse {
	return httpserver.ErrorResponse{Message: &message}
}
