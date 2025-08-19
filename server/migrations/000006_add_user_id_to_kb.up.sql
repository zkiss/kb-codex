-- Add user_id column to knowledge_bases table
ALTER TABLE knowledge_bases ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE knowledge_bases ADD CONSTRAINT fk_knowledge_bases_user_id FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE knowledge_bases DROP CONSTRAINT knowledge_bases_name_key;
ALTER TABLE knowledge_bases ADD CONSTRAINT knowledge_bases_name_user_id_key UNIQUE (name, user_id);
