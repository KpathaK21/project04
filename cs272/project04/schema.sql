CREATE TABLE sqlite_documents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT UNIQUE NOT NULL
);

CREATE TABLE term_frequencies (
    document_id INTEGER,
    term TEXT,
    frequency INTEGER,
    PRIMARY KEY (document_id, term),
    FOREIGN KEY (document_id) REFERENCES sqlite_documents(id)
);
