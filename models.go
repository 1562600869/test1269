package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var ValidBookTypes = []string{"文学", "历史", "科普", "儿童", "生活", "其他"}

const (
	StatusAvailable = "在库"
	StatusBorrowed  = "借出"
)

var ValidBookStatuses = []string{StatusAvailable, StatusBorrowed}

func IsValidBookType(t string) bool {
	for _, v := range ValidBookTypes {
		if v == t {
			return true
		}
	}
	return false
}

func IsValidBookStatus(s string) bool {
	for _, v := range ValidBookStatuses {
		if v == s {
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

func (s *Store) validateAll() error {
	for _, b := range s.Books {
		if !IsValidBookType(b.Type) {
			return fmt.Errorf("书籍 [%s] %s 的类型「%s」非法，可选值: %v", b.ID, b.Title, b.Type, ValidBookTypes)
		}
		if !IsValidBookStatus(b.Status) {
			return fmt.Errorf("书籍 [%s] %s 的状态「%s」非法，可选值: %v", b.ID, b.Title, b.Status, ValidBookStatuses)
		}
	}
	return nil
}

func (s *Store) save() error {
	if err := s.validateAll(); err != nil {
		return fmt.Errorf("数据校验失败: %w", err)
	}

	bookData, err := json.MarshalIndent(s.Books, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化书籍数据失败: %w", err)
	}
	recData, err := json.MarshalIndent(s.Records, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化借阅数据失败: %w", err)
	}

	bookTmp := s.bookFile + ".tmp"
	recTmp := s.recFile + ".tmp"
	bookOrigStat, bookHadOrig := fileStat(s.bookFile)
	var bookOrig []byte
	if bookHadOrig {
		bookOrig, _ = os.ReadFile(s.bookFile)
	}

	bookDir := filepath.Dir(s.bookFile)
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	recDir := filepath.Dir(s.recFile)
	if err := os.MkdirAll(recDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(bookTmp, bookData, 0644); err != nil {
		return fmt.Errorf("写入书籍临时文件失败: %w", err)
	}
	if err := os.Rename(bookTmp, s.bookFile); err != nil {
		os.Remove(bookTmp)
		return fmt.Errorf("提交书籍文件失败: %w", err)
	}

	if err := os.WriteFile(recTmp, recData, 0644); err != nil {
		rollbackFile(s.bookFile, bookOrig, bookHadOrig, bookOrigStat)
		return fmt.Errorf("写入借阅临时文件失败: %w", err)
	}
	if err := os.Rename(recTmp, s.recFile); err != nil {
		os.Remove(recTmp)
		rollbackFile(s.bookFile, bookOrig, bookHadOrig, bookOrigStat)
		return fmt.Errorf("提交借阅文件失败: %w", err)
	}

	_ = bookOrigStat
	return nil
}

func fileStat(path string) (os.FileInfo, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	return info, true
}

func rollbackFile(path string, orig []byte, hadOrig bool, origStat os.FileInfo) {
	if !hadOrig {
		os.Remove(path)
		return
	}
	os.WriteFile(path, orig, 0644)
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
