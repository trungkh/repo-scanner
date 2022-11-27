package model

import (
	"time"
)

type (
	Repository struct {
		Id         int64      `json:"repository_id" db:"repository_id" sqlq:"@{ primary: true; sortable: true; conds: $key; }"`
		Name       string     `json:"repository_name" db:"repository_name" sqlq:"@{ sortable: true; conds: $key, $text; }"`
		Url        string     `json:"repository_url" db:"repository_url" sqlq:"@{ sortable: false; }"`
		IsActive   bool       `json:"is_active" db:"is_active" sqlq:"@{ sortable: true; conds: $basic; }"`
		CreatedBy  string     `json:"created_by" db:"created_by" sqlq:"@{ sortable: true; conds: $key, $text; }"`
		CreatedAt  time.Time  `json:"created_at" db:"created_at" sqlq:"@{ sortable: true; conds: $number; }"`
		ModifiedBy string     `json:"modified_by" db:"modified_by" sqlq:"@{ sortable: true; conds: $key, $text; }"`
		ModifiedAt time.Time  `json:"modified_at" db:"modified_at" sqlq:"@{ sortable: true; conds: $number; }"`
		DeletedBy  *string    `json:"deleted_by" db:"deleted_by" sqlq:"@{ sortable: true; conds: $key, $text, $nullable; }"`
		DeletedAt  *time.Time `json:"deleted_at" db:"deleted_at" sqlq:"@{ sortable: true; soft-del: true; conds: $number, $nullable; }"`
	}

	RepositoryListRequest struct {
		Limit int64 `json:"limit" validate:"numeric,min=1,max=10"` // limit item per page
		Page  int64 `json:"page" validate:"numeric,min=1"`
	}
	RepositoryListResponse struct {
		Id       int64  `json:"repository_id" db:"repository_id"`
		Name     string `json:"repository_name" db:"repository_name"`
		Url      string `json:"repository_url" db:"repository_url"`
		IsActive bool   `json:"is_active" db:"is_active"`
	}

	AddRepositoryRequest struct {
		Name string `json:"repository_name" validate:"required"`
		Url  string `json:"repository_url" validate:"required"`
	}
	AddRepositoryResponse struct {
		Id       int64  `json:"repository_id" db:"repository_id"`
		Name     string `json:"repository_name" db:"repository_name"`
		Url      string `json:"repository_url" db:"repository_url"`
		IsActive bool   `json:"is_active" db:"is_active"`
	}

	EditRepositoryRequest struct {
		Id       int64   `json:"_"`
		Name     *string `json:"repository_name"`
		Url      *string `json:"repository_url"`
		IsActive *bool   `json:"is_active"`
	}
	EditRepositoryResponse struct {
		Id       int64  `json:"repository_id" db:"repository_id"`
		Name     string `json:"repository_name" db:"repository_name"`
		Url      string `json:"repository_url" db:"repository_url"`
		IsActive bool   `json:"is_active" db:"is_active"`
	}
)
