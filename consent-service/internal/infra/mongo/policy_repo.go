package mongo

import (
	"context"
	"time"

	"github.com/beaesthetic/consent-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PolicyRepository implements domain.PolicyRepository using MongoDB
type PolicyRepository struct {
	collection *mongo.Collection
}

// NewPolicyRepository creates a new PolicyRepository
func NewPolicyRepository(db *mongo.Database) *PolicyRepository {
	return &PolicyRepository{
		collection: db.Collection("policies"),
	}
}

// FindBySlug finds a policy by tenant and slug
func (r *PolicyRepository) FindBySlug(tenantID, slug string) (*domain.Policy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"tenant_id": tenantID, "slug": slug}
	var policy domain.Policy
	err := r.collection.FindOne(ctx, filter).Decode(&policy)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrPolicyNotFound
		}
		return nil, err
	}

	return &policy, nil
}

// FindBySlugs finds multiple policies by tenant and slugs
func (r *PolicyRepository) FindBySlugs(tenantID string, slugs []string) ([]domain.Policy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"tenant_id": tenantID,
		"slug":      bson.M{"$in": slugs},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var policies []domain.Policy
	if err := cursor.All(ctx, &policies); err != nil {
		return nil, err
	}

	return policies, nil
}

// FindAll finds all policies for a tenant
func (r *PolicyRepository) FindAll(tenantID string) ([]domain.Policy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"tenant_id": tenantID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var policies []domain.Policy
	if err := cursor.All(ctx, &policies); err != nil {
		return nil, err
	}

	return policies, nil
}

// FindAllActive finds all policies with at least one active version for a tenant
func (r *PolicyRepository) FindAllActive(tenantID string) ([]domain.Policy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"tenant_id": tenantID,
		"versions": bson.M{
			"$elemMatch": bson.M{
				"is_active": true,
			},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var policies []domain.Policy
	if err := cursor.All(ctx, &policies); err != nil {
		return nil, err
	}

	return policies, nil
}

// Save saves a new policy
func (r *PolicyRepository) Save(policy *domain.Policy) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, policy)
	return err
}

// Update updates an existing policy
func (r *PolicyRepository) Update(policy *domain.Policy) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": policy.ID}, policy)
	return err
}

// EnsureIndexes creates the necessary indexes for the policies collection
func (r *PolicyRepository) EnsureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "slug", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
