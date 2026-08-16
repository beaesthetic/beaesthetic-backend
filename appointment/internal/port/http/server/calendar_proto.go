package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	applicationv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/application/v2"
	domain "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
	appointmentcontracts "github.com/petretiandrea/beaesthetic-backend/core-contracts/appointment"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultReminderBeforeSeconds = int32(24 * 60 * 60)

var (
	protoJSONUnmarshal = protojson.UnmarshalOptions{DiscardUnknown: false}
	protoJSONMarshal   = protojson.MarshalOptions{UseProtoNames: false, EmitUnpopulated: false}
)

func registerCalendarProtoRoutes(r gin.IRouter, handler *Server) {
	r.POST("/v1/calendar-events", handler.createCalendarEventProto)
	r.GET("/v1/calendar-events/:id", handler.getCalendarEventProto)
	r.GET("/v1/calendar-events", handler.listCalendarEventsProto)
	r.PATCH("/v1/calendar-events/:id", handler.updateCalendarEventProto)
	r.DELETE("/v1/calendar-events/:id", handler.cancelCalendarEventProto)
	r.POST("/v1/calendar-events/:id/reminder/resend", handler.requestReminderResendProto)
}

func (s *Server) createCalendarEventProto(ctx *gin.Context) {
	var request appointmentcontracts.CreateCalendarEventRequest
	if !s.readProtoJSON(ctx, &request) {
		return
	}
	event, err := s.createCalendarEvent(ctx.Request.Context(), &request)
	if err != nil {
		s.writeCalendarError(ctx, err)
		return
	}
	s.writeProtoJSON(ctx, http.StatusCreated, &appointmentcontracts.CreateCalendarEventResult{CalendarEventId: event.ID})
}

func (s *Server) getCalendarEventProto(ctx *gin.Context) {
	view, err := s.calendar.GetCalendarEventView(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		s.writeCalendarError(ctx, err)
		return
	}
	if view == nil {
		s.writeProtoError(ctx, http.StatusNotFound, "calendar event not found")
		return
	}
	s.writeProtoJSON(ctx, http.StatusOK, &appointmentcontracts.GetCalendarEventResponse{Event: calendarEventProto(*view)})
}

func (s *Server) listCalendarEventsProto(ctx *gin.Context) {
	query, err := calendarEventsListQueryFromProto(ctx)
	if err != nil {
		s.writeProtoError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	views, err := s.calendar.ListCalendarEventViews(ctx.Request.Context(), query)
	if err != nil {
		s.writeCalendarError(ctx, err)
		return
	}
	out := make([]*appointmentcontracts.CalendarEvent, 0, len(views))
	for _, view := range views {
		out = append(out, calendarEventProto(view))
	}
	s.writeProtoJSON(ctx, http.StatusOK, &appointmentcontracts.ListCalendarEventsResponse{Events: out})
}

func (s *Server) updateCalendarEventProto(ctx *gin.Context) {
	var request appointmentcontracts.UpdateCalendarEventRequest
	if !s.readProtoJSON(ctx, &request) {
		return
	}
	request.Id = ctx.Param("id")
	command, err := s.updateCalendarEventCommand(ctx.Request.Context(), &request)
	if err != nil {
		s.writeProtoError(ctx, http.StatusBadRequest, err.Error())
		return
	}
	event, err := s.calendar.Update(ctx.Request.Context(), command)
	if err != nil {
		s.writeCalendarError(ctx, err)
		return
	}
	view, err := s.calendar.GetCalendarEventView(ctx.Request.Context(), event.ID)
	if err != nil {
		s.writeCalendarError(ctx, err)
		return
	}
	if view == nil {
		s.writeProtoError(ctx, http.StatusNotFound, "calendar event not found")
		return
	}
	s.writeProtoJSON(ctx, http.StatusOK, &appointmentcontracts.GetCalendarEventResponse{Event: calendarEventProto(*view)})
}

func (s *Server) cancelCalendarEventProto(ctx *gin.Context) {
	var request appointmentcontracts.CancelCalendarEventRequest
	if ctx.Request.Body != nil && ctx.Request.ContentLength != 0 {
		if !s.readProtoJSON(ctx, &request) {
			return
		}
	}
	reason := cancelReasonFromProto(request.GetReason())
	if reason == "" {
		reason = domain.CancelReasonDeleted
	}
	_, err := s.calendar.CancelEvent(ctx.Request.Context(), applicationv2.CancelEventCommand{CalendarEventID: ctx.Param("id"), Reason: reason})
	if err != nil {
		s.writeCalendarError(ctx, err)
		return
	}
	s.writeProtoJSON(ctx, http.StatusOK, &appointmentcontracts.CancelCalendarEventResponse{})
}

func (s *Server) requestReminderResendProto(ctx *gin.Context) {
	eventID := ctx.Param("id")
	var request appointmentcontracts.RequestReminderResendRequest
	if ctx.Request.Body != nil && ctx.Request.ContentLength != 0 {
		if !s.readProtoJSON(ctx, &request) {
			return
		}
		if request.GetCalendarEventId() != "" {
			eventID = request.GetCalendarEventId()
		}
	}
	idempotencyKey := strings.TrimSpace(request.GetIdempotencyKey())
	if idempotencyKey == "" {
		idempotencyKey = uuid.NewString()
	}
	if err := s.reminders.RequestReminderResend(ctx.Request.Context(), eventID, idempotencyKey); err != nil {
		s.writeCalendarError(ctx, err)
		return
	}
	view, err := s.calendar.GetCalendarEventView(ctx.Request.Context(), eventID)
	if err != nil {
		s.writeCalendarError(ctx, err)
		return
	}
	if view == nil {
		s.writeProtoError(ctx, http.StatusNotFound, "calendar event not found")
		return
	}
	response := appointmentcontracts.GetCalendarEventResponse{Event: calendarEventProto(*view)}
	s.writeProtoJSON(ctx, http.StatusOK, &response)
}

func (s *Server) createCalendarEvent(ctx context.Context, request *appointmentcontracts.CreateCalendarEventRequest) (domain.CalendarEvent, error) {
	base, err := createBaseFromProto(request.GetCalendarId(), request.GetTimeRange(), request.GetTitle(), request.GetDescription(), request.GetVisibility())
	if err != nil {
		return domain.CalendarEvent{}, err
	}
	switch detail := request.GetDetail().(type) {
	case *appointmentcontracts.CreateCalendarEventRequest_Appointment:
		remindBefore := reminderBeforeFromProto(detail.Appointment.RemindBeforeSeconds)
		services, err := s.serviceItemsFromProto(ctx, detail.Appointment.GetServices())
		if err != nil {
			return domain.CalendarEvent{}, err
		}
		event, err := s.calendar.Create(ctx, v2CreateAppointmentCommand(base, detail.Appointment.GetCustomerId(), services, remindBefore))
		if err != nil {
			return domain.CalendarEvent{}, err
		}
		return *event, nil
	case *appointmentcontracts.CreateCalendarEventRequest_ManualEvent:
		location := optionalString(detail.ManualEvent.GetLocation())
		event, err := s.calendar.Create(ctx, v2CreateManualEventCommand(base, detail.ManualEvent.GetTitle(), detail.ManualEvent.GetDescription(), location))
		if err != nil {
			return domain.CalendarEvent{}, err
		}
		return *event, nil
	case *appointmentcontracts.CreateCalendarEventRequest_TimeBlock:
		event, err := s.calendar.Create(ctx, v2CreateTimeBlockCommand(base, detail.TimeBlock.GetReason()))
		if err != nil {
			return domain.CalendarEvent{}, err
		}
		return *event, nil
	default:
		return domain.CalendarEvent{}, fmt.Errorf("event detail is required")
	}
}

func (s *Server) updateCalendarEventCommand(ctx context.Context, request *appointmentcontracts.UpdateCalendarEventRequest) (applicationv2.UpdateEventCommand, error) {
	calendarEventID := request.GetId()
	if calendarEventID == "" {
		return nil, fmt.Errorf("id is required")
	}
	paths := updatePaths(request.GetUpdateMask())
	if err := validateUpdateMask(paths); err != nil {
		return nil, err
	}
	changes := applicationv2.CalendarEventChanges{}
	if hasUpdatePath(paths, "time_range", "timeRange") {
		update, err := timeRangeUpdateFromProto(request.GetTimeRange())
		if err != nil {
			return nil, err
		}
		changes.TimeRange = update
	}
	if hasUpdatePath(paths, "title") {
		changes.Title = maskedString(paths, "title", "title", request.Title)
	}
	if hasUpdatePath(paths, "description") {
		changes.Description = maskedString(paths, "description", "description", request.Description)
	}
	if hasUpdatePath(paths, "visibility") {
		visibility := visibilityFromProto(request.GetVisibility())
		changes.Visibility = &visibility
	}
	detailType, err := updateDetailType(paths)
	if err != nil {
		return nil, err
	}
	switch detail := request.GetDetail().(type) {
	case *appointmentcontracts.UpdateCalendarEventRequest_Appointment:
		if detailType != "appointment" {
			return nil, fmt.Errorf("updateMask must include only appointment detail fields")
		}
		services, err := s.serviceItemsFromProto(ctx, detail.Appointment.GetServices())
		if err != nil {
			return nil, err
		}
		return applicationv2.UpdateAppointmentCommand{
			CalendarEventID: calendarEventID,
			Changes:         changes,
			Services:        services,
		}, nil
	case *appointmentcontracts.UpdateCalendarEventRequest_ManualEvent:
		if detailType != "manual_event" {
			return nil, fmt.Errorf("updateMask must include only manualEvent detail fields")
		}
		command := applicationv2.UpdateManualEventCommand{
			CalendarEventID: calendarEventID,
			Changes:         changes,
			Title:           maskedString(paths, "manual_event.title", "manualEvent.title", detail.ManualEvent.Title),
			Description:     maskedString(paths, "manual_event.description", "manualEvent.description", detail.ManualEvent.Description),
		}
		if hasUpdatePath(paths, "manual_event.location", "manualEvent.location") {
			location := optionalString(detail.ManualEvent.GetLocation())
			command.Location = &location
		}
		return command, nil
	case *appointmentcontracts.UpdateCalendarEventRequest_TimeBlock:
		if detailType != "time_block" {
			return nil, fmt.Errorf("updateMask must include only timeBlock detail fields")
		}
		return applicationv2.UpdateTimeBlockCommand{
			CalendarEventID: calendarEventID,
			Changes:         changes,
			Reason:          detail.TimeBlock.GetReason(),
		}, nil
	default:
		if detailType != "" {
			return nil, fmt.Errorf("%s detail is required by updateMask", detailType)
		}
		return applicationv2.UpdateCalendarFieldsCommand{CalendarEventID: calendarEventID, Changes: changes}, nil
	}
}

func (s *Server) serviceItemsFromProto(ctx context.Context, selections []*appointmentcontracts.AppointmentServiceSelection) ([]domain.ServiceItem, error) {
	out := make([]domain.ServiceItem, 0, len(selections))
	for index, selection := range selections {
		switch value := selection.GetValue().(type) {
		case *appointmentcontracts.AppointmentServiceSelection_CustomServiceName:
			item, err := domain.NewServiceItem(nil, value.CustomServiceName, nil, index)
			if err != nil {
				return nil, err
			}
			out = append(out, item)
		case *appointmentcontracts.AppointmentServiceSelection_CatalogServiceId:
			service, err := s.services.FindService(ctx, value.CatalogServiceId)
			if err != nil {
				return nil, err
			}
			if service == nil {
				return nil, fmt.Errorf("service %s not found", value.CatalogServiceId)
			}
			serviceID := service.ID
			price := service.Price
			item, err := domain.NewServiceItem(&serviceID, service.Name, &price, index)
			if err != nil {
				return nil, err
			}
			out = append(out, item)
		default:
			return nil, fmt.Errorf("service selection is required")
		}
	}
	return out, nil
}

func (s *Server) readProtoJSON(ctx *gin.Context, message proto.Message) bool {
	defer ctx.Request.Body.Close()
	payload, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		s.writeProtoError(ctx, http.StatusBadRequest, err.Error())
		return false
	}
	if err := protoJSONUnmarshal.Unmarshal(payload, message); err != nil {
		s.writeProtoError(ctx, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

func (s *Server) writeProtoJSON(ctx *gin.Context, status int, message proto.Message) {
	payload, err := protoJSONMarshal.Marshal(message)
	if err != nil {
		s.log.Error("marshal proto json response", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.Data(status, "application/json; charset=utf-8", payload)
}

func (s *Server) writeCalendarError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, applicationv2.ErrCalendarEventNotFound):
		s.writeProtoError(ctx, http.StatusNotFound, err.Error())
	case errors.Is(err, applicationv2.ErrAppointmentNotRemindable),
		errors.Is(err, applicationv2.ErrInvalidReminderRequest):
		s.writeProtoError(ctx, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrMissingRequiredData),
		errors.Is(err, domain.ErrInvalidCalendarID),
		errors.Is(err, domain.ErrInvalidTimeRange),
		errors.Is(err, domain.ErrInvalidEventType),
		errors.Is(err, domain.ErrInvalidEventDetail),
		errors.Is(err, domain.ErrInvalidVisibility),
		errors.Is(err, domain.ErrInvalidReminder):
		s.writeProtoError(ctx, http.StatusBadRequest, err.Error())
	default:
		s.writeProtoError(ctx, http.StatusInternalServerError, err.Error())
	}
}

func (s *Server) writeProtoError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"message": message})
}

func calendarEventsListQueryFromProto(ctx *gin.Context) (applicationv2.ListCalendarEventsQuery, error) {
	var query applicationv2.ListCalendarEventsQuery
	calendarID, err := domain.NormalizeCalendarID(ctx.Query("calendarId"))
	if err != nil {
		return query, err
	}
	query.CalendarID = calendarID
	query.CustomerID = ctx.Query("customerId")
	if raw := ctx.Query("startAt"); raw != "" {
		start, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return query, fmt.Errorf("invalid startAt")
		}
		query.Start = &start
	}
	if raw := ctx.Query("endAt"); raw != "" {
		end, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return query, fmt.Errorf("invalid endAt")
		}
		query.End = &end
	}
	for _, raw := range ctx.QueryArray("eventTypes") {
		eventType, err := calendarEventTypeFromString(raw)
		if err != nil {
			return query, err
		}
		query.EventTypes = append(query.EventTypes, eventType)
	}
	return query, nil
}

func createBaseFromProto(calendarID string, timeRange *appointmentcontracts.TimeRange, title string, description string, visibility appointmentcontracts.CalendarEventVisibility) (calendarEventBase, error) {
	calendarID, err := domain.NormalizeCalendarID(calendarID)
	if err != nil {
		return calendarEventBase{}, err
	}
	rangeUpdate, err := timeRangeUpdateFromProto(timeRange)
	if err != nil {
		return calendarEventBase{}, err
	}
	return calendarEventBase{
		CalendarID:  calendarID,
		Start:       rangeUpdate.Start,
		End:         rangeUpdate.End,
		Timezone:    rangeUpdate.Timezone,
		AllDay:      rangeUpdate.AllDay,
		Title:       title,
		Description: description,
		Visibility:  visibilityFromProto(visibility),
	}, nil
}

func timeRangeUpdateFromProto(timeRange *appointmentcontracts.TimeRange) (*applicationv2.TimeRangeUpdate, error) {
	if timeRange == nil || timeRange.GetStartAt() == nil || timeRange.GetEndAt() == nil {
		return nil, fmt.Errorf("timeRange.startAt and timeRange.endAt are required")
	}
	return &applicationv2.TimeRangeUpdate{
		Start:    timeRange.GetStartAt().AsTime(),
		End:      timeRange.GetEndAt().AsTime(),
		Timezone: timeRange.GetTimezone(),
		AllDay:   timeRange.GetAllDay(),
	}, nil
}

func calendarEventProto(view applicationv2.CalendarEventView) *appointmentcontracts.CalendarEvent {
	event := view.Event
	out := &appointmentcontracts.CalendarEvent{
		Id:         event.ID,
		CalendarId: event.CalendarID,
		EventType:  calendarEventTypeProto(event.Type),
		TimeRange: &appointmentcontracts.TimeRange{
			StartAt:  timestamppb.New(event.Range.Start),
			EndAt:    timestamppb.New(event.Range.End),
			Timezone: event.Range.Timezone,
			AllDay:   event.Range.AllDay,
		},
		Title:       event.Title,
		Description: event.Description,
		Visibility:  visibilityProto(event.Visibility),
		Version:     event.Version,
		CreatedAt:   timestamppb.New(event.CreatedAt),
		UpdatedAt:   timestamppb.New(event.UpdatedAt),
	}
	if event.Cancellation != nil {
		out.Cancellation = &appointmentcontracts.CalendarEventCancellation{
			Reason:     cancelReasonProto(event.Cancellation.Reason),
			CanceledAt: timestamppb.New(event.Cancellation.CanceledAt),
		}
	}
	switch detail := event.Detail.(type) {
	case domain.Appointment:
		services := make([]*appointmentcontracts.AppointmentServiceItem, 0, len(detail.Services))
		for _, item := range detail.Services {
			services = append(services, &appointmentcontracts.AppointmentServiceItem{
				ServiceId:   stringValue(item.ServiceID),
				ServiceName: item.ServiceName,
				Price:       float64Value(item.Price),
				Position:    int32(item.Position),
			})
		}
		out.Detail = &appointmentcontracts.CalendarEvent_Appointment{Appointment: &appointmentcontracts.AppointmentDetail{
			Customer: &appointmentcontracts.CustomerRef{CustomerId: detail.Customer.ID, DisplayName: detail.Customer.DisplayName},
			Services: services,
			Reminder: appointmentReminderProto(view.Reminder),
		}}
	case domain.ManualEvent:
		out.Detail = &appointmentcontracts.CalendarEvent_ManualEvent{ManualEvent: &appointmentcontracts.ManualEventDetail{
			Title:       detail.Title,
			Description: detail.Description,
			Location:    stringValue(detail.Location),
		}}
	case domain.TimeBlock:
		out.Detail = &appointmentcontracts.CalendarEvent_TimeBlock{TimeBlock: &appointmentcontracts.TimeBlockDetail{Reason: detail.Reason}}
	}
	return out
}

func appointmentReminderProto(reminder *domain.AppointmentReminder) *appointmentcontracts.AppointmentReminder {
	if reminder == nil {
		return nil
	}
	out := &appointmentcontracts.AppointmentReminder{
		Status:              appointmentReminderStatusProto(reminder.Status),
		RemindBeforeSeconds: int32(reminder.RemindBefore.Seconds()),
		FailureReason:       stringValue(reminder.FailureReason),
	}
	if reminder.ScheduledAt != nil {
		out.ScheduledAt = timestamppb.New(*reminder.ScheduledAt)
	}
	if reminder.SentRequestedAt != nil {
		out.SentRequestedAt = timestamppb.New(*reminder.SentRequestedAt)
	}
	if reminder.SentAt != nil {
		out.SentAt = timestamppb.New(*reminder.SentAt)
	}
	if reminder.FailedAt != nil {
		out.FailedAt = timestamppb.New(*reminder.FailedAt)
	}
	return out
}

func appointmentReminderStatusProto(status domain.ReminderStatus) appointmentcontracts.AppointmentReminderStatus {
	switch status {
	case domain.ReminderStatusPending:
		return appointmentcontracts.AppointmentReminderStatus_APPOINTMENT_REMINDER_STATUS_PENDING
	case domain.ReminderStatusScheduled:
		return appointmentcontracts.AppointmentReminderStatus_APPOINTMENT_REMINDER_STATUS_SCHEDULED
	case domain.ReminderStatusUnprocessable:
		return appointmentcontracts.AppointmentReminderStatus_APPOINTMENT_REMINDER_STATUS_UNPROCESSABLE
	case domain.ReminderStatusSendRequested:
		return appointmentcontracts.AppointmentReminderStatus_APPOINTMENT_REMINDER_STATUS_SEND_REQUESTED
	case domain.ReminderStatusSent:
		return appointmentcontracts.AppointmentReminderStatus_APPOINTMENT_REMINDER_STATUS_SENT
	case domain.ReminderStatusFailed:
		return appointmentcontracts.AppointmentReminderStatus_APPOINTMENT_REMINDER_STATUS_FAILED
	case domain.ReminderStatusDeleted:
		return appointmentcontracts.AppointmentReminderStatus_APPOINTMENT_REMINDER_STATUS_DELETED
	default:
		return appointmentcontracts.AppointmentReminderStatus_APPOINTMENT_REMINDER_STATUS_UNSPECIFIED
	}
}

func updatePaths(mask *fieldmaskpb.FieldMask) map[string]struct{} {
	out := map[string]struct{}{}
	if mask == nil || len(mask.Paths) == 0 {
		return out
	}
	for _, path := range mask.Paths {
		out[path] = struct{}{}
	}
	return out
}

func hasUpdatePath(paths map[string]struct{}, candidates ...string) bool {
	for _, candidate := range candidates {
		if _, ok := paths[candidate]; ok {
			return true
		}
	}
	return false
}

func validateUpdateMask(paths map[string]struct{}) error {
	if len(paths) == 0 {
		return fmt.Errorf("updateMask is required")
	}
	allowed := map[string]struct{}{
		"time_range":               {},
		"timeRange":                {},
		"title":                    {},
		"description":              {},
		"visibility":               {},
		"appointment.services":     {},
		"manual_event.title":       {},
		"manualEvent.title":        {},
		"manual_event.description": {},
		"manualEvent.description":  {},
		"manual_event.location":    {},
		"manualEvent.location":     {},
		"time_block.reason":        {},
		"timeBlock.reason":         {},
	}
	for path := range paths {
		if _, ok := allowed[path]; !ok {
			return fmt.Errorf("unsupported updateMask path %q", path)
		}
	}
	return nil
}

func updateDetailType(paths map[string]struct{}) (string, error) {
	detailType := ""
	for path := range paths {
		candidate := ""
		switch {
		case strings.HasPrefix(path, "appointment."):
			candidate = "appointment"
		case strings.HasPrefix(path, "manual_event."), strings.HasPrefix(path, "manualEvent."):
			candidate = "manual_event"
		case strings.HasPrefix(path, "time_block."), strings.HasPrefix(path, "timeBlock."):
			candidate = "time_block"
		}
		if candidate == "" {
			continue
		}
		if detailType != "" && detailType != candidate {
			return "", fmt.Errorf("updateMask cannot contain fields from different event details")
		}
		detailType = candidate
	}
	return detailType, nil
}

func maskedString(paths map[string]struct{}, snakePath string, camelPath string, value *string) *string {
	if !hasUpdatePath(paths, snakePath, camelPath) {
		return nil
	}
	if value == nil {
		empty := ""
		return &empty
	}
	return value
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func float64Value(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func visibilityFromProto(value appointmentcontracts.CalendarEventVisibility) domain.Visibility {
	switch value {
	case appointmentcontracts.CalendarEventVisibility_CALENDAR_EVENT_VISIBILITY_PUBLIC:
		return domain.VisibilityPublic
	default:
		return domain.VisibilityPrivate
	}
}

func visibilityProto(value domain.Visibility) appointmentcontracts.CalendarEventVisibility {
	switch value {
	case domain.VisibilityPublic:
		return appointmentcontracts.CalendarEventVisibility_CALENDAR_EVENT_VISIBILITY_PUBLIC
	default:
		return appointmentcontracts.CalendarEventVisibility_CALENDAR_EVENT_VISIBILITY_PRIVATE
	}
}

func calendarEventTypeProto(value domain.CalendarEventType) appointmentcontracts.CalendarEventType {
	switch value {
	case domain.CalendarEventTypeAppointment:
		return appointmentcontracts.CalendarEventType_CALENDAR_EVENT_TYPE_APPOINTMENT
	case domain.CalendarEventTypeManual:
		return appointmentcontracts.CalendarEventType_CALENDAR_EVENT_TYPE_MANUAL
	case domain.CalendarEventTypeTimeBlock:
		return appointmentcontracts.CalendarEventType_CALENDAR_EVENT_TYPE_TIME_BLOCK
	default:
		return appointmentcontracts.CalendarEventType_CALENDAR_EVENT_TYPE_UNSPECIFIED
	}
}

func calendarEventTypeFromString(value string) (domain.CalendarEventType, error) {
	switch strings.TrimSpace(value) {
	case "appointment", "CALENDAR_EVENT_TYPE_APPOINTMENT":
		return domain.CalendarEventTypeAppointment, nil
	case "manual", "CALENDAR_EVENT_TYPE_MANUAL":
		return domain.CalendarEventTypeManual, nil
	case "time_block", "CALENDAR_EVENT_TYPE_TIME_BLOCK":
		return domain.CalendarEventTypeTimeBlock, nil
	default:
		return "", fmt.Errorf("invalid eventTypes value %q", value)
	}
}

func cancelReasonFromProto(value appointmentcontracts.CancelReason) domain.CancelReason {
	switch value {
	case appointmentcontracts.CancelReason_CANCEL_REASON_CUSTOMER_CANCEL:
		return domain.CancelReasonCustomer
	case appointmentcontracts.CancelReason_CANCEL_REASON_DELETED:
		return domain.CancelReasonDeleted
	default:
		return ""
	}
}

func cancelReasonProto(value domain.CancelReason) appointmentcontracts.CancelReason {
	switch value {
	case domain.CancelReasonCustomer:
		return appointmentcontracts.CancelReason_CANCEL_REASON_CUSTOMER_CANCEL
	case domain.CancelReasonDeleted:
		return appointmentcontracts.CancelReason_CANCEL_REASON_DELETED
	default:
		return appointmentcontracts.CancelReason_CANCEL_REASON_UNSPECIFIED
	}
}

type calendarEventBase struct {
	CalendarID  string
	Start       time.Time
	End         time.Time
	Timezone    string
	AllDay      bool
	Title       string
	Description string
	Visibility  domain.Visibility
}

func reminderBeforeFromProto(_ *int32) time.Duration {
	return time.Duration(defaultReminderBeforeSeconds) * time.Second
}

func v2CreateAppointmentCommand(base calendarEventBase, customerID string, services []domain.ServiceItem, remindBefore time.Duration) applicationv2.CreateAppointmentCommand {
	return applicationv2.CreateAppointmentCommand{
		CalendarID:   base.CalendarID,
		Start:        base.Start,
		End:          base.End,
		Timezone:     base.Timezone,
		AllDay:       base.AllDay,
		Title:        base.Title,
		Description:  base.Description,
		Visibility:   base.Visibility,
		CustomerID:   customerID,
		Services:     services,
		RemindBefore: remindBefore,
	}
}

func v2CreateManualEventCommand(base calendarEventBase, title string, description string, location *string) applicationv2.CreateManualEventCommand {
	return applicationv2.CreateManualEventCommand{
		CalendarID:    base.CalendarID,
		Start:         base.Start,
		End:           base.End,
		Timezone:      base.Timezone,
		AllDay:        base.AllDay,
		Title:         base.Title,
		Description:   base.Description,
		Visibility:    base.Visibility,
		ManualTitle:   title,
		ManualDetails: description,
		Location:      location,
	}
}

func v2CreateTimeBlockCommand(base calendarEventBase, reason string) applicationv2.CreateTimeBlockCommand {
	return applicationv2.CreateTimeBlockCommand{
		CalendarID:  base.CalendarID,
		Start:       base.Start,
		End:         base.End,
		Timezone:    base.Timezone,
		AllDay:      base.AllDay,
		Title:       base.Title,
		Description: base.Description,
		Visibility:  base.Visibility,
		Reason:      reason,
	}
}
