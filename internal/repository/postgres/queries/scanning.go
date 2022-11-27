package queries

const (
	GetScanningList = `
		SELECT 
			s.scanning_id,
			r.repository_name,
			r.repository_url,
			s.findings,
			s.scanning_status,
			s.queued_at,
			s.scanning_at,
			s.finished_at
		FROM
			reposcan.repositories r
		JOIN
			reposcan.scannings s
		ON
			r.repository_id = s.repository_id
		WHERE
			('all'=$1 OR s.scanning_status = $1::reposcan.scanning_status)
		AND	s.deleted_by IS NULL
		ORDER BY
			s.created_at %v
		LIMIT $2
		OFFSET $3
	`

	InsertNewScanning = `
		INSERT INTO reposcan.scannings(
			repository_id,
			queued_at,
			created_by,
			created_at,
			modified_by,
			modified_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			scanning_id,
			repository_id,
			findings,
			scanning_status,
			queued_at,
			scanning_at,
			finished_at
	`

	UpdateScanningInProgressById = `
		UPDATE reposcan.scannings
		SET
			scanning_status = $2,
			scanning_at = $4,
			modified_by = $3,
			modified_at = $4
		WHERE
			scanning_id = $1
		RETURNING
			scanning_id,
			repository_id,
			findings,
			scanning_status,
			queued_at,
			scanning_at,
			finished_at
	`

	UpdateScanningFinishedById = `
		UPDATE reposcan.scannings
		SET
			scanning_status = $2,
			finished_at = $5,
			findings = $3,
			modified_by = $4,
			modified_at = $5
		WHERE
			scanning_id = $1
		RETURNING
			scanning_id,
			repository_id,
			findings,
			scanning_status,
			queued_at,
			scanning_at,
			finished_at
	`
)
