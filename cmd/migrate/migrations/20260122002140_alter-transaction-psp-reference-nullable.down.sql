UPDATE transactions SET psp_reference = 'unknown' WHERE psp_reference IS NULL;
ALTER TABLE transactions ALTER COLUMN psp_reference SET NOT NULL;