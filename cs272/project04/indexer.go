package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SQLiteDocument represents a document in the SQLite database.
type SQLiteDocument struct {
	ID   uint   `gorm:"primaryKey"`  // Unique identifier for each document
	Path string `gorm:"unique"`      // File path or URL of the document; must be unique
}

// Word represents a unique word in the database.
type Word struct {
	ID   uint   `gorm:"primaryKey"` // Unique identifier for each word
	Term string `gorm:"unique"`     // Unique term (word) in the database
}

// TermFrequency represents the frequency of a term in a specific document.
type TermFrequency struct {
	DocumentID uint `gorm:"primaryKey"` // Foreign key to SQLiteDocument
	WordID     uint `gorm:"primaryKey"` // Foreign key to Word
	Frequency  int  `gorm:"not null"`   // Frequency count of the term in the document
}

// Link represents a URL link in the database.
type Link struct {
	ID   uint   `gorm:"primaryKey"` // Unique identifier for each link
	URL  string `gorm:"unique"`     // URL link; must be unique
	Title string
}

// TableName specifies a custom table name for TermFrequency.
func (TermFrequency) TableName() string {
	return "term_freq"
}

// TableName specifies a custom table name for Word.
func (Word) TableName() string {
	return "words"
}

// TableName specifies a custom table name for SQLiteDocument.
func (SQLiteDocument) TableName() string {
	return "documents"
}

// TableName specifies a custom table name for Link.
func (Link) TableName() string {
	return "links"
}

// createTables initializes the tables in the SQLite database by first dropping existing tables (if any)
// and then creating new ones based on the provided models.
func createTables(db *gorm.DB) error {
	// Drop the existing tables if they exist
	if db.Migrator().HasTable(&TermFrequency{}) {
		if err := db.Migrator().DropTable(&TermFrequency{}); err != nil {
			return err
		}
	}
	if db.Migrator().HasTable(&Word{}) {
		if err := db.Migrator().DropTable(&Word{}); err != nil {
			return err
		}
	}
	if db.Migrator().HasTable(&SQLiteDocument{}) {
		if err := db.Migrator().DropTable(&SQLiteDocument{}); err != nil {
			return err
		}
	}
	if db.Migrator().HasTable(&Link{}) {
		if err := db.Migrator().DropTable(&Link{}); err != nil {
			return err
		}
	}
	// Create the tables based on the models
	if err := db.AutoMigrate(&SQLiteDocument{}, &Word{}, &TermFrequency{}, &Link{}); err != nil {
		return err
	}
	return nil
}

// SQLiteStorage is a struct that manages interactions with SQLite for storing
// term frequencies, documents, and links.
type SQLiteStorage struct {
	db *gorm.DB // GORM database connection instance
}

// NewSQLiteStorage initializes a new SQLiteStorage instance and creates tables if necessary.
func NewSQLiteStorage(dbName string) (*SQLiteStorage, error) {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Create tables immediately after opening the database
	if err := createTables(db); err != nil {
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

// AddDocument inserts a document and its term frequencies into the database.
// It uses a transaction to ensure atomicity and rollback on failure.
func (s *SQLiteStorage) AddDocument(path string, termFreq map[string]int) error {
    tx := s.db.Begin() // Begin transaction
    document := SQLiteDocument{Path: path}

    // Insert the document record
    if err := tx.Create(&document).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to add document: %w", err)
    }

    termInserts := make([]TermFrequency, 0, len(termFreq)) // Prepare term frequency entries
		
    for term, freq := range termFreq {
        var word Word
        // Check if the term already exists in the database
        if err := tx.Where("term = ?", term).First(&word).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                word = Word{Term: term} // Create new word record
                if err := tx.Create(&word).Error; err != nil {
                    tx.Rollback()
                    return fmt.Errorf("failed to add word: %w", err)
                }
            } else {
                tx.Rollback()
                return fmt.Errorf("failed to query word: %w", err)
            }
        }

        // Append the term frequency record to the list for bulk insertion
        termInserts = append(termInserts, TermFrequency{
            DocumentID: document.ID,
            WordID:     word.ID,
            Frequency:  freq,
        })
    }

    // Insert all term frequencies in a single batch
    if err := tx.Create(&termInserts).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to add term frequencies: %w", err)
    }

    return tx.Commit().Error // Commit transaction
}


// GetTermFrequency retrieves the term frequency for a specified term across all documents.
func (s *SQLiteStorage) GetTermFrequency(term string) (map[string]int, error) {
	var results []struct {
		DocumentID uint
		Frequency  int
	}

	// Find the word ID for the specified term
	var word Word
	if err := s.db.Where("Term = ?", term).First(&word).Error; err != nil {
		return nil, fmt.Errorf("failed to get term: %w", err)
	}

	// Query the term frequencies for the term across documents
	if err := s.db.Raw("SELECT document_id, frequency FROM term_freq WHERE word_id = ?", word.ID).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get term frequencies: %w", err)
	}

	// Map the results to document paths and their frequencies
	termFrequency := make(map[string]int)
	for _, result := range results {
		var document SQLiteDocument
		// Retrieve the document path for each DocumentID
		if err := s.db.First(&document, result.DocumentID).Error; err == nil {
			termFrequency[document.Path] = result.Frequency
		}
	}

	return termFrequency, nil
}

// GetDocumentFrequency retrieves the document frequency (DF) of a specific term.
// Document frequency is the number of documents containing the term.
func (s *SQLiteStorage) GetDocumentFrequency(term string) (int, error) {
	var word Word
	// Retrieve the word ID for the term
	if err := s.db.Where("Term = ?", term).First(&word).Error; err != nil {
		return 0, fmt.Errorf("failed to get term: %w", err)
	}

	var count int64
	// Count documents containing the word ID
	if err := s.db.Model(&TermFrequency{}).Where("word_id = ?", word.ID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get document frequency: %w", err)
	}
	return int(count), nil
}

// GetDocumentCount retrieves the total number of documents stored in the database.
func (s *SQLiteStorage) GetDocumentCount() (int, error) {
	var count int64
	// Count all documents in the SQLiteDocument table
	if err := s.db.Model(&SQLiteDocument{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get document count: %w", err)
	}
	return int(count), nil
}

// AddLink inserts a unique URL link with a title into the database.
func (s *SQLiteStorage) AddLink(url string, title string) error {
    link := Link{URL: url, Title: title}
    if err := s.db.Create(&link).Error; err != nil {
        return fmt.Errorf("failed to add link: %w", err)
    }
    return nil
}

// GetTitle retrieves the title for a specific URL.
func (s *SQLiteStorage) GetTitle(url string) (string, error) {
    var link Link
    if err := s.db.Where("url = ?", url).First(&link).Error; err != nil {
        return "", fmt.Errorf("title not found for URL %s: %w", url, err)
    }
    return link.Title, nil
}

// GetAllLinks retrieves all links and their titles from the database using GORM.
func (s *SQLiteStorage) GetAllLinks() ([]Link, error) {
    var links []Link
    result := s.db.Find(&links) 

    if result.Error != nil {
        return nil, result.Error
    }
    return links, nil
}
