package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	applicationv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/application/v2"
	legacydomain "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	domain "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
	appointmentcontracts "github.com/petretiandrea/beaesthetic-backend/core-contracts/appointment"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestReminderBeforeFromProtoAlwaysUsesBackendDefault(t *testing.T) {
	frontendValue := int32(0)
	if value := reminderBeforeFromProto(&frontendValue); value != 24*time.Hour {
		t.Fatalf("reminder before = %s, want 24h", value)
	}
}

func TestAppointmentReminderProtoMapsLifecycleState(t *testing.T) {
	scheduledAt := time.Date(2026, 8, 8, 9, 0, 0, 0, time.UTC)
	sentRequestedAt := scheduledAt.Add(time.Hour)
	reminder := &domain.AppointmentReminder{
		Status:          domain.ReminderStatusSendRequested,
		RemindBefore:    24 * time.Hour,
		ScheduledAt:     &scheduledAt,
		SentRequestedAt: &sentRequestedAt,
	}

	mapped := appointmentReminderProto(reminder)
	if mapped.GetStatus() != appointmentcontracts.AppointmentReminderStatus_APPOINTMENT_REMINDER_STATUS_SEND_REQUESTED {
		t.Fatalf("status = %s", mapped.GetStatus())
	}
	if mapped.GetRemindBeforeSeconds() != 86400 || !mapped.GetScheduledAt().AsTime().Equal(scheduledAt) || !mapped.GetSentRequestedAt().AsTime().Equal(sentRequestedAt) {
		t.Fatalf("mapped reminder = %#v", mapped)
	}
}

func TestUpdateCalendarEventCommandBuildsCommonFieldsCommand(t *testing.T) {
	command, err := (&Server{}).updateCalendarEventCommand(context.Background(), &appointmentcontracts.UpdateCalendarEventRequest{
		Id:         "event-1",
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"title"}},
	})
	if err != nil {
		t.Fatalf("updateCalendarEventCommand() error = %v", err)
	}
	update, ok := command.(applicationv2.UpdateCalendarFieldsCommand)
	if !ok {
		t.Fatalf("command type = %T, want UpdateCalendarFieldsCommand", command)
	}
	if update.Changes.Title == nil || *update.Changes.Title != "" {
		t.Fatalf("title change = %#v, want an explicit empty value", update.Changes.Title)
	}
}

func TestUpdateCalendarEventCommandRejectsMixedDetailMasks(t *testing.T) {
	_, err := (&Server{}).updateCalendarEventCommand(context.Background(), &appointmentcontracts.UpdateCalendarEventRequest{
		Id: "event-1",
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
			"appointment.services",
			"time_block.reason",
		}},
	})
	if err == nil {
		t.Fatal("updateCalendarEventCommand() error = nil, want mixed detail mask error")
	}
}

func TestUpdateCalendarEventCommandClearsManualLocation(t *testing.T) {
	command, err := (&Server{}).updateCalendarEventCommand(context.Background(), &appointmentcontracts.UpdateCalendarEventRequest{
		Id:         "event-1",
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"manual_event.location"}},
		Detail: &appointmentcontracts.UpdateCalendarEventRequest_ManualEvent{
			ManualEvent: &appointmentcontracts.UpdateManualEventDetail{},
		},
	})
	if err != nil {
		t.Fatalf("updateCalendarEventCommand() error = %v", err)
	}
	update, ok := command.(applicationv2.UpdateManualEventCommand)
	if !ok {
		t.Fatalf("command type = %T, want UpdateManualEventCommand", command)
	}
	if update.Location == nil || *update.Location != nil {
		t.Fatalf("location change = %#v, want explicit nil location", update.Location)
	}
}

func TestListServicesRequestFromQuery(t *testing.T) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest("GET", "/v1/services?query=facial&limit=4", nil)

	request, err := listServicesRequestFromQuery(context)
	if err != nil {
		t.Fatalf("listServicesRequestFromQuery() error = %v", err)
	}
	if request.GetQuery() != "facial" || request.GetLimit() != 4 {
		t.Fatalf("request = %#v, want query facial and limit 4", request)
	}
}

func TestListServicesRequestFromQueryRejectsInvalidLimit(t *testing.T) {
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest("GET", "/v1/services?limit=0", nil)

	if _, err := listServicesRequestFromQuery(context); err == nil {
		t.Fatal("listServicesRequestFromQuery() error = nil, want invalid limit error")
	}
}

func TestServiceServiceLimitsAnUnfilteredCatalog(t *testing.T) {
	repository := &serviceRepositoryStub{}
	service := application.NewServiceService(repository)

	if _, err := service.SearchServices(context.Background(), "", 2); err != nil {
		t.Fatalf("SearchServices() error = %v", err)
	}
	if repository.query != "" || repository.limit != 2 {
		t.Fatalf("search = (%q, %d), want unfiltered search limited to 2", repository.query, repository.limit)
	}
}

func TestCatalogServiceProtoDoesNotExposePrice(t *testing.T) {
	service := catalogServiceProto(legacydomain.AppointmentService{
		ID:   "service-1",
		Name: "Facial treatment",
		Tags: []string{"facial"},
	})

	encoded, err := protojson.Marshal(service)
	if err != nil {
		t.Fatalf("protojson.Marshal() error = %v", err)
	}
	if strings.Contains(string(encoded), "price") {
		t.Fatalf("catalog service response exposes price: %s", encoded)
	}
}

func TestListServicesProtoReturnsTheCatalogWithoutPrice(t *testing.T) {
	repository := &serviceRepositoryStub{searchResults: []legacydomain.AppointmentService{{
		ID:   "service-1",
		Name: "Facial treatment",
		Tags: []string{"facial"},
	}}}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/v1/services?query=facial&limit=2", nil)

	(&Server{services: application.NewServiceService(repository)}).listServicesProto(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if repository.query != "facial" || repository.limit != 2 {
		t.Fatalf("search = (%q, %d), want (facial, 2)", repository.query, repository.limit)
	}
	if strings.Contains(recorder.Body.String(), "price") {
		t.Fatalf("response exposes price: %s", recorder.Body.String())
	}
}

type serviceRepositoryStub struct {
	searchResults []legacydomain.AppointmentService
	query         string
	limit         int
}

func (s *serviceRepositoryStub) SaveService(_ context.Context, service legacydomain.AppointmentService) (legacydomain.AppointmentService, error) {
	return service, nil
}

func (s *serviceRepositoryStub) FindServices(_ context.Context) ([]legacydomain.AppointmentService, error) {
	return s.searchResults, nil
}

func (s *serviceRepositoryStub) SearchServices(_ context.Context, query string, limit int) ([]legacydomain.AppointmentService, error) {
	s.query = query
	s.limit = limit
	return s.searchResults, nil
}

func (s *serviceRepositoryStub) FindService(_ context.Context, _ string) (*legacydomain.AppointmentService, error) {
	return nil, nil
}
