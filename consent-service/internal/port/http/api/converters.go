package api

import (
	"github.com/beaesthetic/consent-service/internal/domain"
)

// Pointer helpers

func ptr[T any](v T) *T {
	return &v
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// Response builders

func errResponse(msg string) InternalServerErrorJSONResponse {
	return InternalServerErrorJSONResponse{Error: ptr(msg)}
}

func notFoundResponse(msg string) NotFoundJSONResponse {
	return NotFoundJSONResponse{Error: ptr(msg)}
}

func badRequestResponse(msg string) BadRequestJSONResponse {
	return BadRequestJSONResponse{Error: ptr(msg)}
}

// Domain to API converters

func domainPolicyToAPI(p domain.Policy) Policy {
	versions := make([]PolicyVersion, len(p.Versions))
	for i, v := range p.Versions {
		versions[i] = PolicyVersion{
			Version:         ptr(v.Version),
			PublishedAt:     ptr(v.PublishedAt),
			ContentHtml:     ptr(v.ContentHTML),
			ContentMarkdown: ptr(v.ContentMarkdown),
			PdfUrl:          ptr(v.PDFURL),
			IsActive:        ptr(v.IsActive),
		}
	}

	return Policy{
		Slug:        ptr(p.Slug),
		Name:        ptr(p.Name),
		Description: ptr(p.Description),
		Versions:    &versions,
		CreatedAt:   ptr(p.CreatedAt),
		UpdatedAt:   ptr(p.UpdatedAt),
	}
}

func domainConsentToAPI(c domain.Consent) Consent {
	return Consent{
		Id:               ptr(c.ID),
		Subject:          ptr(c.Subject),
		PolicySlug:       ptr(c.PolicySlug),
		PolicyVersion:    ptr(c.PolicyVersion),
		AcceptedAt:       ptr(c.AcceptedAt),
		AcceptanceMethod: ptr(ConsentAcceptanceMethod(c.AcceptanceMethod)),
		LinkToken:        c.LinkToken,
		RevokedAt:        c.RevokedAt,
		RevokedBy:        c.RevokedBy,
	}
}

func domainLinkToAPI(l domain.ConsentLink) ConsentLink {
	return ConsentLink{
		Token:     ptr(l.Token),
		Subject:   ptr(l.Subject),
		Policies:  &l.Policies,
		CreatedAt: ptr(l.CreatedAt),
		ExpiresAt: ptr(l.ExpiresAt),
		UsedAt:    l.UsedAt,
		CreatedBy: ptr(l.CreatedBy),
	}
}
