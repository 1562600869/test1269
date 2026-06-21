package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var ValidBookTypes = []string{"文学", "历史", "科普", "儿童", "生活", "其他"}

func IsValidBookType(t string) bool {
	for _, v := range ValidBookTypes {
		if v == t {
			return true
		}
	}
	return false
}

type Book struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Donor      string `json:"donor"`
	DonorPhone string `json:"donor_phone"`
	DonateDate string `json:"donate_date"`
}

type BorrowRecord struct {
	BookID      string `json:"book_id"`
	BookTitle   string `json:"book_title"`
	Member      string `json:"member"`
	MemberPhone string `json:"member_phone"`
	BorrowDate  string `json:"borrow_date"`
	ReturnDate  string `json:"return_date"`
	Returned    bool   `json:"returned"`
}

type Store struct {
	Books    []Book         `json:"books"`
	Records  []BorrowRecord `json:"records"`
	mu       sync.Mutex
	bookFile string
	recFile  string
}

func NewStore(bookFile, recFile string) *Store {
	s := &Store{
		bookFile: bookFile,
		recFile:  recFile,
	}
	s.load()
	return s
}

func (s *Store) load() {
	s.Books = nil
	s.Records = nil

	if data, err := os.ReadFile(s.bookFile); err == nil {
		json.Unmarshal(data, &s.Books)
	}
	if data, err := os.ReadFile(s.recFile); err == nil {
		json.Unmarshal(data, &s.Records)
	}

	if s.Books == nil {
		s.Books = []Book{}
	}
	if s.Records == nil {
		s.Records = []BorrowRecord{}
	}
}

func (s *Store) save() error {
	bookData, err := json.MarshalIndent(s.Books, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化书籍数据失败: %w", err)
	}
	if err := os.WriteFile(s.bookFile, bookData, 0644); err != nil {
		return fmt.Errorf("写入书籍数据失败: %w", err)
	}

	recData, err := json.MarshalIndent(s.Records, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化借阅数据失败: %w", err)
	}
	if err := os.WriteFile(s.recFile, recData, 0644); err != nil {
		return fmt.Errorf("写入借阅数据失败: %w", err)
	}

	return nil
}

func (s *Store) Lock() {
	s.mu.Lock()
}

func (s *Store) Unlock() {
	s.mu.Unlock()
}

func (s *Store) FindBookByID(id string) *Book {
	for i := range s.Books {
		if s.Books[i].ID == id {
			return &s.Books[i]
		}
	}
	return nil
}

func (s *Store) FindActiveBorrow(bookID, memberPhone string) *BorrowRecord {
	for i := range s.Records {
		if s.Records[i].BookID == bookID && s.Records[i].MemberPhone == memberPhone && !s.Records[i].Returned {
			return &s.Records[i]
		}
	}
	return nil
}
