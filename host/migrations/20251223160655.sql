-- Create themes table
CREATE TABLE themes (
		name TEXT,
		title TEXT,
		category TEXT,
		keyword TEXT

);
-- Create actors table
CREATE TABLE actors (
		name TEXT PRIMARY KEY,
		site TEXT
);
-- Create sessions table
CREATE TABLE sessions (
		id INTEGER PRIMARY KEY,
		theme INTEGER,
		start8601 TEXT,
		stop8601 TEXT
);
-- Create sessions_actors table
CREATE TABLE sessions_actors (
		id INTEGER PRIMARY KEY,
		actor TEXT,
		session INTEGER
);
-- Create themes_locks table
CREATE TABLE themes_locks (
		id INTEGER PRIMARY KEY,
		islocked BOOLEAN,
		theme INTEGER
);
-- Create actors_locks table
CREATE TABLE actors_locks (
		id INTEGER PRIMARY KEY,
		islocked BOOLEAN,
		session INTEGER,
		actor TEXT
);
-- Create tx table
CREATE TABLE tx (
		session INTEGER,
		t8601 TEXT,
		text TEXT
);
