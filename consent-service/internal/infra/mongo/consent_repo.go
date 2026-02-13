package mongo

import (
	"context"
	"time"

	"github.com/beaesthetic/consent-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConsentRepository implements domain.ConsentRepository using MongoDB
type ConsentRepository struct {
	collection *mongo.Collection
}

// NewConsentRepository creates a new ConsentRepository
func NewConsentRepository(db *mongo.Database) *ConsentRepository {
	return &ConsentRepository{
		collection: db.Collection("consents"),
	}
}

// FindByID finds a consent by its ID
func (r *ConsentRepository) FindByID(id string) (*domain.Consent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var consent domain.Consent
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&consent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrConsentNotFound
		}
		return nil, err
	}

	return &consent, nil
}

// FindBySubject finds all consents for a subject within a tenant
func (r *ConsentRepository) FindBySubject(tenantID, subject string) ([]domain.Consent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"tenant_id": tenantID, "subject": subject}
	opts := options.Find().SetSort(bson.D{{Key: "accepted_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var consents []domain.Consent
	if err := cursor.All(ctx, &consents); err != nil {
		return nil, err
	}

	return consents, nil
}

// FindBySubjectAndPolicy finds the most recent consent for a subject and policy within a tenant
func (r *ConsentRepository) FindBySubjectAndPolicy(tenantID, subject, policySlug string) (*domain.Consent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"tenant_id":   tenantID,
		"subject":     subject,
		"policy_slug": policySlug,
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "accepted_at", Value: -1}})

	var consent domain.Consent
	err := r.collection.FindOne(ctx, filter, opts).Decode(&consent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrConsentNotFound
		}
		return nil, err
	}

	return &consent, nil
}

// FindActiveBySubjectAndPolicy finds the active (non-revoked) consent for a subject and policy within a tenant
func (r *ConsentRepository) FindActiveBySubjectAndPolicy(tenantID, subject, policySlug string) (*domain.Consent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"tenant_id":   tenantID,
		"subject":     subject,
		"policy_slug": policySlug,
		"revoked_at":  nil,
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "accepted_at", Value: -1}})

	var consent domain.Consent
	err := r.collection.FindOne(ctx, filter, opts).Decode(&consent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrConsentNotFound
		}
		return nil, err
	}

	return &consent, nil
}

// FindActiveBySubjectAndPolicies finds all active (non-revoked) consents for a subject and multiple policies
func (r *ConsentRepository) FindActiveBySubjectAndPolicies(tenantID, subject string, slugs []string) ([]domain.Consent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"tenant_id":   tenantID,
		"subject":     subject,
		"policy_slug": bson.M{"$in": slugs},
		"revoked_at":  nil,
	}
	opts := options.Find().SetSort(bson.D{{Key: "accepted_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var consents []domain.Consent
	if err := cursor.All(ctx, &consents); err != nil {
		return nil, err
	}

	return consents, nil
}

// Save saves a new consent
func (r *ConsentRepository) Save(consent *domain.Consent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, consent)
	return err
}

// Update updates an existing consent
func (r *ConsentRepository) Update(consent *domain.Consent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": consent.ID}, consent)
	return err
}

// EnsureIndexes creates the necessary indexes for the consents collection
func (r *ConsentRepository) EnsureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "subject", Value: 1},
				{Key: "policy_slug", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "subject", Value: 1},
			},
		},
		{
			Keys: bson.D{{Key: "accepted_at", Value: -1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
