package model

import (
	"time"

	"github.com/jmoiron/sqlx/types"
)

type (
	Scanning struct {
		Id           int64          `json:"scanning_id" db:"scanning_id" sqlq:"@{ primary: true; sortable: true; conds: $key; }"`
		RepositoryId int64          `json:"repository_id" db:"repository_id" sqlq:"@{ foreign: true; sortable: true; conds: $key; }"`
		Findings     types.JSONText `json:"findings" db:"findings"`
		Status       string         `json:"scanning_status" db:"scanning_status" sqlq:"@{ sortable: true; conds: $key; }"`
		QueuedAt     time.Time      `json:"queued_at" db:"queued_at" sqlq:"@{ sortable: true; conds: $number; }"`
		ScanningAt   *time.Time     `json:"scanning_at" db:"scanning_at" sqlq:"@{ sortable: true; conds: $number, $nullable; }"`
		FinishedAt   *time.Time     `json:"finished_at" db:"finished_at" sqlq:"@{ sortable: true; conds: $number, $nullable; }"`
		IsActive     bool           `json:"is_active" db:"is_active" sqlq:"@{ sortable: true; conds: $basic; }"`
		CreatedBy    string         `json:"created_by" db:"created_by" sqlq:"@{ sortable: true; conds: $key, $text; }"`
		CreatedAt    time.Time      `json:"created_at" db:"created_at" sqlq:"@{ sortable: true; conds: $number; }"`
		ModifiedBy   string         `json:"modified_by" db:"modified_by" sqlq:"@{ sortable: true; conds: $key, $text; }"`
		ModifiedAt   time.Time      `json:"modified_at" db:"modified_at" sqlq:"@{ sortable: true; conds: $number; }"`
		DeletedBy    *string        `json:"deleted_by" db:"deleted_by" sqlq:"@{ sortable: true; conds: $key, $text, $nullable; }"`
		DeletedAt    *time.Time     `json:"deleted_at" db:"deleted_at" sqlq:"@{ sortable: true; soft-del: true; conds: $number, $nullable; }"`
	}

	ScanningListRequest struct {
		Limit  int64  `json:"limit" validate:"numeric,min=1,max=10"` // limit item per page
		Page   int64  `json:"page" validate:"numeric,min=1"`
		Sort   string `json:"sort" validate:"oneof=asc desc"`
		Status string `json:"status" validate:"oneof=all queued in_progress success failure"`
	}
	ScanningListResponse struct {
		Id         int64          `json:"scanning_id" db:"scanning_id"`
		Name       string         `json:"repository_name" db:"repository_name"`
		Url        string         `json:"repository_url" db:"repository_url"`
		Findings   types.JSONText `json:"findings" db:"findings"`
		Status     string         `json:"scanning_status" db:"scanning_status"`
		QueuedAt   time.Time      `json:"queued_at" db:"queued_at"`
		ScanningAt *time.Time     `json:"scanning_at" db:"scanning_at"`
		FinishedAt *time.Time     `json:"finished_at" db:"finished_at"`
	}

	ScanningResponse struct {
		Id         int64          `json:"scanning_id" db:"scanning_id"`
		RepoId     int64          `json:"repository_id" db:"repository_id"`
		Findings   types.JSONText `json:"findings" db:"findings"`
		Status     string         `json:"scanning_status" db:"scanning_status"`
		QueuedAt   time.Time      `json:"queued_at" db:"queued_at"`
		ScanningAt *time.Time     `json:"scanning_at" db:"scanning_at"`
		FinishedAt *time.Time     `json:"finished_at" db:"finished_at"`
	}
)
