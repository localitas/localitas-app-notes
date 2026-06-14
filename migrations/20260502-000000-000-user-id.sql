ALTER TABLE notes ADD COLUMN user_id TEXT NOT NULL DEFAULT '';
UPDATE notes SET user_id = '2b9af8b9-856a-4710-9ab6-58fe4eccdf24' WHERE user_id = '';
CREATE INDEX IF NOT EXISTS idx_notes_user ON notes(user_id);
