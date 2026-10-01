CREATE UNIQUE INDEX IF NOT EXISTS services_level1_source_names_unique
    ON services_level1_source_names
        (service_level1_id, original_name, source_sheet, source_row, source_range);
