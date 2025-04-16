DROP TABLE IF EXISTS tasks;
CREATE TABLE tasks (
  id         INT AUTO_INCREMENT NOT NULL,
  title      VARCHAR(128) NOT NULL,
  done       BOOLEAN NOT NULL,
  PRIMARY KEY (`id`)
);

INSERT INTO tasks
  (title, done)
VALUES
  ('cat', FALSE),
  ('Giant Steps', TRUE),
  ('Jeru', TRUE),
  ('Sarah Vaughan', FALSE);