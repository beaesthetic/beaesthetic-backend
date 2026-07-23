package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type RiverJobInserter struct {
	db     *ContextDB
	client *river.Client[pgx.Tx]
}

func NewRiverJobInserter(db *ContextDB, client *river.Client[pgx.Tx]) *RiverJobInserter {
	return &RiverJobInserter{db: db, client: client}
}

func (i *RiverJobInserter) Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) error {
	tx, ok := i.db.executor(ctx).(pgx.Tx)
	if !ok {
		return fmt.Errorf("river job insert requires a postgres transaction")
	}
	_, err := i.client.InsertTx(ctx, tx, args, opts)
	return err
}

func (i *RiverJobInserter) CancelByKey(ctx context.Context, kind string, queue string, key string) error {
	tx, ok := i.db.executor(ctx).(pgx.Tx)
	if !ok {
		return fmt.Errorf("river job cancel requires a postgres transaction")
	}
	metadata, err := json.Marshal(map[string]string{"idempotencyKey": key})
	if err != nil {
		return err
	}
	jobs, err := i.client.JobListTx(ctx, tx, river.NewJobListParams().
		First(100).
		Kinds(kind).
		Queues(queue).
		States(
			rivertype.JobStateAvailable,
			rivertype.JobStatePending,
			rivertype.JobStateRetryable,
			rivertype.JobStateRunning,
			rivertype.JobStateScheduled,
		).
		Metadata(string(metadata)),
	)
	if err != nil {
		return err
	}
	for _, job := range jobs.Jobs {
		if _, err := i.client.JobCancelTx(ctx, tx, job.ID); err != nil {
			return err
		}
	}
	return nil
}
