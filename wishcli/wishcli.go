package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Luzifer/AMZWishlist"
	"github.com/crackcomm/go-clitable"
)

func main() {
	wishlist := os.Args[1]
	elements := amzwishlist.ScrapeWishlist(wishlist)

	table := clitable.New([]string{"Title", "Price", "Requested", "Received"})
	for _, wish := range elements {
		row := make(map[string]interface{})
		row["Title"] = truncate(wish.Title)
		row["Price"] = wish.Price
		row["Requested"] = wish.Requested
		row["Received"] = wish.Received
		table.AddRow(row)
	}
	table.Print()
}

func truncate(s string) string {
	if len(s) > 75 {
		return fmt.Sprintf("%s...", strings.TrimSpace(s[:75]))
	}
	return s
}
