package mongo

import (
	"context"
	"time"

	"github.com/beaesthetic/consent-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

// FindBySlug finds a policy by its slug
func (r *PolicyRepository) FindBySlug(slug string) (*domain.Policy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var policy domain.Policy
	err := r.collection.FindOne(ctx, bson.M{"_id": slug}).Decode(&policy)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrPolicyNotFound
		}
		return nil, err
	}

	return &policy, nil
}

// FindAll finds all policies
func (r *PolicyRepository) FindAll() ([]domain.Policy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
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

// FindAllActive finds all policies that have at least one active version
func (r *PolicyRepository) FindAllActive() ([]domain.Policy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
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

	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": policy.Slug}, policy)
	return err
}
