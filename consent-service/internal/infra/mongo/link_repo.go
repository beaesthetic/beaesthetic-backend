package mongo

import (
	"context"
	"time"

	"github.com/beaesthetic/consent-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConsentLinkRepository implements domain.ConsentLinkRepository using MongoDB
type ConsentLinkRepository struct {
	collection *mongo.Collection
}

// NewConsentLinkRepository creates a new ConsentLinkRepository
func NewConsentLinkRepository(db *mongo.Database) *ConsentLinkRepository {
	return &ConsentLinkRepository{
		collection: db.Collection("consent_links"),
	}
}

// FindByToken finds a consent link by its token
func (r *ConsentLinkRepository) FindByToken(token string) (*domain.ConsentLink, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var link domain.ConsentLink
	err := r.collection.FindOne(ctx, bson.M{"_id": token}).Decode(&link)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrLinkNotFound
		}
		return nil, err
	}

	return &link, nil
}

// Save saves a new consent link
func (r *ConsentLinkRepository) Save(link *domain.ConsentLink) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, link)
	return err
}

// Update updates an existing consent link
func (r *ConsentLinkRepository) Update(link *domain.ConsentLink) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": link.Token}, link)
	return err
}

// Delete deletes a consent link
func (r *ConsentLinkRepository) Delete(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": token})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrLinkNotFound
	}
	return nil
}

// EnsureIndexes creates the necessary indexes for the consent_links collection
func (r *ConsentLinkRepository) EnsureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			// TTL index for automatic cleanup of expired links
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "subject", Value: 1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
