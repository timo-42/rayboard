CREATE VIRTUAL TABLE attachment_fts USING fts5(
	attachment_id UNINDEXED,
	file_name,
	content_type
);
