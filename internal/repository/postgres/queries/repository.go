package queries

const (
	GetRepositoryList = `
		SELECT 
			repository_id,
			repository_name,
			repository_url,
			is_active
		FROM
			reposcan.repositories
		WHERE
			deleted_by IS NULL
		ORDER BY
			repository_id DESC
		LIMIT $1
		OFFSET $2
	`

	GetRepositoryById = `
		SELECT 
			repository_id,
			repository_name,
			repository_url,
			is_active,
			created_by,
			created_at,
			modified_by,
			modified_at,
			deleted_by,
			deleted_at
		FROM
			reposcan.repositories
		WHERE
			repository_id = $1
		AND deleted_by IS NULL
	`

	InsertNewRepository = `
		INSERT INTO reposcan.repositories(
			repository_name,
			repository_url,
			created_by,
			created_at,
			modified_by,
			modified_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			repository_id,
			repository_name,
			repository_url,
			is_active
	`

	EditRepository = `
		UPDATE reposcan.repositories
		SET
			repository_name = CASE WHEN $2 = true THEN $3
							ELSE repository_name END,
			repository_url = CASE WHEN $4 = true THEN $5
							ELSE repository_url END,
			is_active = CASE WHEN $6 = true THEN
								CASE WHEN $7 THEN '1'::BIT
									ELSE '0'::BIT END
							ELSE is_active END,
			modified_by = $8,
			modified_at = $9
		WHERE
			repository_id = $1
		RETURNING
			repository_id,
			repository_name,
			repository_url,
			is_active
	`

	DeleteRepository = `
		UPDATE reposcan.repositories
		SET
			is_active = '0'::BIT,
			deleted_by = $2,
			deleted_at = $3
		WHERE
			repository_id = $1
	`
)
