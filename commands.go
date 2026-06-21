package main

import (
	"fmt"
	"time"
)

func CmdDonate(s *Store, bookID, title, donor, phone, bookType, date string) error {
	if !IsValidBookType(bookType) {
		return fmt.Errorf("无效的书籍类型: %s，可选值: %v", bookType, ValidBookTypes)
	}
	if s.FindBookByID(bookID) != nil {
		return fmt.Errorf("书籍编号 %s 已存在", bookID)
	}

	book := Book{
		ID:         bookID,
		Title:      title,
		Type:       bookType,
		Status:     "在库",
		Donor:      donor,
		DonorPhone: phone,
		DonateDate: date,
	}

	s.Lock()
	s.load()
	s.Books = append(s.Books, book)
	if err := s.save(); err != nil {
		s.Unlock()
		return err
	}
	s.Unlock()

	fmt.Printf("✅ 捐赠入库成功: [%s] %s (%s) — 捐赠人: %s\n", bookID, title, bookType, donor)
	return nil
}

func CmdBorrow(s *Store, bookID, member, phone, date string) error {
	s.Lock()
	s.load()

	book := s.FindBookByID(bookID)
	if book == nil {
		s.Unlock()
		return fmt.Errorf("书籍编号 %s 不存在", bookID)
	}
	if book.Status != "在库" {
		s.Unlock()
		return fmt.Errorf("书籍 [%s] %s 当前状态为「%s」，无法借出", bookID, book.Title, book.Status)
	}

	active := s.FindActiveBorrow(bookID, phone)
	if active != nil {
		s.Unlock()
		return fmt.Errorf("会员 %s(%s) 已借阅 [%s] %s 且尚未归还", member, phone, bookID, book.Title)
	}

	record := BorrowRecord{
		BookID:      bookID,
		BookTitle:   book.Title,
		Member:      member,
		MemberPhone: phone,
		BorrowDate:  date,
		ReturnDate:  "",
		Returned:    false,
	}

	book.Status = "借出"
	s.Records = append(s.Records, record)

	if err := s.save(); err != nil {
		s.Unlock()
		return err
	}
	s.Unlock()

	fmt.Printf("✅ 借阅成功: [%s] %s — 借阅人: %s(%s)\n", bookID, book.Title, member, phone)
	return nil
}

func CmdReturn(s *Store, bookID, phone, date string) error {
	s.Lock()
	s.load()

	book := s.FindBookByID(bookID)
	if book == nil {
		s.Unlock()
		return fmt.Errorf("书籍编号 %s 不存在", bookID)
	}

	var target *BorrowRecord
	for i := range s.Records {
		if s.Records[i].BookID == bookID && s.Records[i].MemberPhone == phone && !s.Records[i].Returned {
			target = &s.Records[i]
			break
		}
	}
	if target == nil {
		s.Unlock()
		return fmt.Errorf("未找到手机号 %s 对 [%s] 的有效借阅记录", phone, bookID)
	}

	target.Returned = true
	target.ReturnDate = date
	book.Status = "在库"

	if err := s.save(); err != nil {
		s.Unlock()
		return err
	}
	s.Unlock()

	fmt.Printf("✅ 归还成功: [%s] %s — 借阅人: %s(%s)，归还日期: %s\n", bookID, book.Title, target.Member, phone, date)
	return nil
}

func CmdOverdue(s *Store) error {
	s.Lock()
	s.load()
	s.Unlock()

	now := time.Now()
	found := false

	fmt.Println("📋 超期未还列表:")
	fmt.Printf("%-8s %-20s %-10s %-15s %-12s %s\n", "编号", "书名", "借阅人", "手机号", "借阅日期", "超期天数")
	fmt.Println("--------------------------------------------------------------------------")

	for _, r := range s.Records {
		if r.Returned {
			continue
		}
		borrowDate, err := time.Parse("2006-01-02", r.BorrowDate)
		if err != nil {
			continue
		}
		days := int(now.Sub(borrowDate).Hours() / 24)
		if days > 30 {
			overdueDays := days - 30
			fmt.Printf("%-8s %-20s %-10s %-15s %-12s %d天\n", r.BookID, r.BookTitle, r.Member, r.MemberPhone, r.BorrowDate, overdueDays)
			found = true
		}
	}

	if !found {
		fmt.Println("暂无超期未还记录")
	}
	return nil
}

func CmdMonthly(s *Store, month string) error {
	s.Lock()
	s.load()
	s.Unlock()

	typeStats := make(map[string]struct {
		DonateCount int
		BorrowCount int
	})
	for _, t := range ValidBookTypes {
		s := typeStats[t]
		typeStats[t] = s
	}

	donateCount := 0
	for _, b := range s.Books {
		if len(b.DonateDate) >= 7 && b.DonateDate[:7] == month {
			entry := typeStats[b.Type]
			entry.DonateCount++
			typeStats[b.Type] = entry
			donateCount++
		}
	}

	borrowCount := 0
	for _, r := range s.Records {
		if len(r.BorrowDate) >= 7 && r.BorrowDate[:7] == month {
			book := s.FindBookByID(r.BookID)
			if book == nil {
				continue
			}
			entry := typeStats[book.Type]
			entry.BorrowCount++
			typeStats[book.Type] = entry
			borrowCount++
		}
	}

	fmt.Printf("📊 %s 月度统计:\n", month)
	fmt.Printf("%-10s %-10s %-10s\n", "类型", "捐赠数", "借阅数")
	fmt.Println("------------------------------")
	for _, t := range ValidBookTypes {
		stats := typeStats[t]
		fmt.Printf("%-10s %-10d %-10d\n", t, stats.DonateCount, stats.BorrowCount)
	}
	fmt.Println("------------------------------")
	fmt.Printf("%-10s %-10d %-10d\n", "合计", donateCount, borrowCount)
	return nil
}
