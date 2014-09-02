package amzwishlist

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bjarneh/latinx"
	"github.com/crackcomm/go-clitable"
	"gopkg.in/xmlpath.v2"

	"code.google.com/p/go.net/html"
)

type Wish struct {
	Title, Price        string
	Requested, Received int
}

func main() {
	wishlist := os.Args[1]
	elements := ScrapeWishlist(wishlist)

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
		return fmt.Sprintf("%s...", s[:75])
	}
	return s
}

func ScrapeWishlist(wishlist string) []Wish {
	url := fmt.Sprintf("http://www.amazon.de/registry/wishlist/%s/?reveal=unpurchased&filter=all&sort=date-added&layout=compact", wishlist)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, read_err := ioutil.ReadAll(resp.Body)
	if read_err != nil {
		panic(read_err)
	}

	good, converr := latinx.Decode(latinx.ISO_8859_15, body)
	if converr != nil {
		panic(converr)
	}

	reader := strings.NewReader(sanitizeHtml(string(good)))
	xmlroot, xmlerr := xmlpath.ParseHTML(reader)
	if xmlerr != nil {
		panic(xmlerr)
	}

	pathTable := xmlpath.MustCompile("//*[contains(@class,'g-compact-items')]/tbody/tr[preceding::tr]")
	iterator := pathTable.Iter(xmlroot)
	wishes := []Wish{}

	for iterator.Next() {
		line := iterator.Node()

		title := getXpathString("td[@class='g-title']/a[@class='a-link-normal']", line)
		price := getXpathString("td[contains(@class,'g-price')]/span", line)
		requested := getXpathInt("td[@class='g-requested']/span", line)
		gotten := getXpathInt("td[@class='g-received']/span", line)
		result := Wish{title, price, requested, gotten}
		wishes = append(wishes, result)
	}

	return wishes
}

func getXpathString(xpath string, node *xmlpath.Node) string {
	path := xmlpath.MustCompile(xpath)
	if res, ok := path.String(node); ok {
		return strings.Trim(res, " \n")
	} else {
		panic(fmt.Sprintf("No string found for %s", xpath))
	}
}

func getXpathInt(xpath string, node *xmlpath.Node) int {
	str := getXpathString(xpath, node)
	i, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return i
}

func sanitizeHtml(rawHTML string) string {
	reader := strings.NewReader(rawHTML)
	root, err := html.Parse(reader)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer
	html.Render(&b, root)
	fixedHtml := b.String()

	return fixedHtml
}
