package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	dataDir := "."
	bookFile := filepath.Join(dataDir, "books.json")
	recFile := filepath.Join(dataDir, "borrow_records.json")
	store := NewStore(bookFile, recFile)

	cmd := os.Args[1]

	switch cmd {
	case "donate":
		handleDonate(store)
	case "borrow":
		handleBorrow(store)
	case "return":
		handleReturn(store)
	case "overdue":
		handleOverdue(store)
	case "monthly":
		handleMonthly(store)
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("社区图书角管理工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  donate  <书籍编号> <书名> --donor <捐赠人> --phone <电话> --type <类型> --date <日期>  登记捐赠入库")
	fmt.Println("  borrow  <书籍编号> --member <借阅人> --phone <电话> --date <日期>                    借阅书籍")
	fmt.Println("  return  <书籍编号> --member-phone <电话> --date <日期>                               归还书籍")
	fmt.Println("  overdue                                                                           查看超期未还")
	fmt.Println("  monthly --month <年-月>                                                            月度统计")
	fmt.Println()
	fmt.Println("书籍类型: 文学/历史/科普/儿童/生活/其他")
}

func handleDonate(s *Store) {
	args := os.Args[2:]
	if len(args) < 2 {
		fmt.Println("用法: donate <书籍编号> <书名> --donor <捐赠人> --phone <电话> --type <类型> --date <日期>")
		os.Exit(1)
	}

	bookID := args[0]
	title := args[1]

	flags := parseFlags(args[2:])
	donor := flags["donor"]
	phone := flags["phone"]
	bookType := flags["type"]
	date := flags["date"]

	if bookType == "" {
		fmt.Println("错误: 必须指定 --type")
		os.Exit(1)
	}
	if date == "" {
		fmt.Println("错误: 必须指定 --date")
		os.Exit(1)
	}

	if err := CmdDonate(s, bookID, title, donor, phone, bookType, date); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}

func handleBorrow(s *Store) {
	args := os.Args[2:]
	if len(args) < 1 {
		fmt.Println("用法: borrow <书籍编号> --member <借阅人> --phone <电话> --date <日期>")
		os.Exit(1)
	}

	bookID := args[0]
	flags := parseFlags(args[1:])
	member := flags["member"]
	phone := flags["phone"]
	date := flags["date"]

	if member == "" || phone == "" || date == "" {
		fmt.Println("错误: 必须指定 --member, --phone, --date")
		os.Exit(1)
	}

	if err := CmdBorrow(s, bookID, member, phone, date); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}

func handleReturn(s *Store) {
	args := os.Args[2:]
	if len(args) < 1 {
		fmt.Println("用法: return <书籍编号> --member-phone <电话> --date <日期>")
		os.Exit(1)
	}

	bookID := args[0]
	flags := parseFlags(args[1:])
	phone := flags["member-phone"]
	date := flags["date"]

	if phone == "" || date == "" {
		fmt.Println("错误: 必须指定 --member-phone, --date")
		os.Exit(1)
	}

	if err := CmdReturn(s, bookID, phone, date); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}

func handleOverdue(s *Store) {
	if err := CmdOverdue(s); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}

func handleMonthly(s *Store) {
	args := os.Args[2:]
	flags := parseFlags(args)
	month := flags["month"]

	if month == "" {
		fmt.Println("错误: 必须指定 --month (格式: 2024-03)")
		os.Exit(1)
	}

	if err := CmdMonthly(s, month); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags(args []string) map[string]string {
	result := make(map[string]string)
	for i := 0; i < len(args); i++ {
		if len(args[i]) > 2 && args[i][:2] == "--" {
			key := args[i][2:]
			if i+1 < len(args) && len(args[i+1]) > 0 && args[i+1][:2] != "--" {
				result[key] = args[i+1]
				i++
			} else {
				result[key] = ""
			}
		}
	}
	return result
}
