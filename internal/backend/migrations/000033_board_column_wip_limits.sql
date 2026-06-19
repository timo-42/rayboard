ALTER TABLE board_columns
ADD COLUMN wip_limit INTEGER CHECK (wip_limit IS NULL OR wip_limit >= 0);
